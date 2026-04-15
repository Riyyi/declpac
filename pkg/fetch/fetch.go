package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Jguer/dyalpm"
)

var (
	Root       = "/"
	LockFile   = "/var/lib/pacman/db.lock"
	AURInfoURL = "https://aur.archlinux.org/rpc?v=5&type=info"
)

type Fetcher struct {
	aurCache map[string]AURPackage
	handle   dyalpm.Handle
	localDB  dyalpm.Database
	syncDBs  []dyalpm.Database
}

type PackageInfo struct {
	Name      string
	InAUR     bool
	Exists    bool
	Installed bool
	AURInfo   *AURPackage
	syncPkg   dyalpm.Package
}

type AURResponse struct {
	Results []AURPackage `json:"results"`
}

type AURPackage struct {
	Name        string `json:"Name"`
	PackageBase string `json:"PackageBase"`
	Version     string `json:"Version"`
	URL         string `json:"URL"`
}

func New() (*Fetcher, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] Fetcher New: starting...\n")

	handle, err := dyalpm.Initialize(Root, "/var/lib/pacman")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize alpm: %w", err)
	}

	localDB, err := handle.LocalDB()
	if err != nil {
		handle.Release()
		return nil, fmt.Errorf("failed to get local database: %w", err)
	}

	syncDBs, err := handle.SyncDBs()
	if err != nil {
		handle.Release()
		return nil, fmt.Errorf("failed to get sync databases: %w", err)
	}

	if len(syncDBs) == 0 {
		syncDBs, err = registerSyncDBs(handle)
		if err != nil {
			handle.Release()
			return nil, fmt.Errorf("failed to register sync databases: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] Fetcher New: done (%.2fs)\n", time.Since(start).Seconds())
	return &Fetcher{
		aurCache: make(map[string]AURPackage),
		handle:   handle,
		localDB:  localDB,
		syncDBs:  syncDBs,
	}, nil
}

func (f *Fetcher) Close() error {
	if f.handle != nil {
		f.handle.Release()
	}
	return nil
}

func (f *Fetcher) GetAURPackage(name string) (AURPackage, bool) {
	pkg, ok := f.aurCache[name]
	return pkg, ok
}

func (f *Fetcher) BuildLocalPkgMap() (map[string]interface{}, error) {
	localPkgs, err := f.buildLocalPkgMap()
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for k, v := range localPkgs {
		result[k] = v
	}
	return result, nil
}

func registerSyncDBs(handle dyalpm.Handle) ([]dyalpm.Database, error) {
	fmt.Fprintf(os.Stderr, "[debug] registerSyncDBs: starting...\n")

	repos := []string{"core", "extra", "multilib"}
	var dbs []dyalpm.Database

	for _, repo := range repos {
		db, err := handle.RegisterSyncDB(repo, 0)
		if err != nil {
			continue
		}

		count := 0
		db.PkgCache().ForEach(func(pkg dyalpm.Package) error {
			count++
			return nil
		})

		if count > 0 {
			dbs = append(dbs, db)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] registerSyncDBs: done (%d dbs)\n", len(dbs))
	return dbs, nil
}

func (f *Fetcher) buildLocalPkgMap() (map[string]dyalpm.Package, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] buildLocalPkgMap: starting...\n")

	localPkgs := make(map[string]dyalpm.Package)

	err := f.localDB.PkgCache().ForEach(func(pkg dyalpm.Package) error {
		localPkgs[pkg.Name()] = pkg
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate local package cache: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[debug] buildLocalPkgMap: done (%.2fs)\n", time.Since(start).Seconds())
	return localPkgs, nil
}

func (f *Fetcher) checkSyncDBs(pkgNames []string) (map[string]dyalpm.Package, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] checkSyncDBs: starting...\n")

	syncPkgs := make(map[string]dyalpm.Package)
	pkgSet := make(map[string]bool)
	for _, name := range pkgNames {
		pkgSet[name] = true
	}

	for _, db := range f.syncDBs {
		err := db.PkgCache().ForEach(func(pkg dyalpm.Package) error {
			if pkgSet[pkg.Name()] {
				if _, exists := syncPkgs[pkg.Name()]; !exists {
					syncPkgs[pkg.Name()] = pkg
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to iterate sync database %s: %w", db.Name(), err)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] checkSyncDBs: done (%.2fs)\n", time.Since(start).Seconds())
	return syncPkgs, nil
}

func (f *Fetcher) Resolve(packages []string) (map[string]*PackageInfo, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] Resolve: starting...\n")

	result := make(map[string]*PackageInfo)
	for _, pkg := range packages {
		result[pkg] = &PackageInfo{Name: pkg, Exists: true}
	}

	localPkgs, err := f.buildLocalPkgMap()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] Resolve: local pkgs built (%.2fs)\n", time.Since(start).Seconds())

	var notInLocal []string
	for _, pkg := range packages {
		if localPkg, ok := localPkgs[pkg]; ok {
			result[pkg] = &PackageInfo{
				Name:      pkg,
				Exists:    true,
				InAUR:     false,
				Installed: true,
				syncPkg:   localPkg,
			}
		} else {
			notInLocal = append(notInLocal, pkg)
		}
	}

	if len(notInLocal) > 0 {
		syncPkgs, err := f.checkSyncDBs(notInLocal)
		if err != nil {
			return nil, err
		}

		f.ensureAURCache(packages)

		for _, pkg := range packages {
			info := result[pkg]
			if info == nil {
				continue
			}

			if info.Installed {
				if aurInfo, ok := f.aurCache[pkg]; ok {
					info.InAUR = true
					info.AURInfo = &aurInfo
				}
				continue
			}

			if syncPkg, ok := syncPkgs[pkg]; ok {
				info.InAUR = false
				info.Installed = false
				info.syncPkg = syncPkg
				continue
			}

			if aurInfo, ok := f.aurCache[pkg]; ok {
				info.InAUR = true
				info.Installed = false
				info.AURInfo = &aurInfo
				continue
			}

			return nil, fmt.Errorf("package not found: %s", pkg)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] Resolve: done (%.2fs)\n", time.Since(start).Seconds())
	return result, nil
}

func (f *Fetcher) ensureAURCache(packages []string) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] ensureAURCache: starting...\n")

	if len(packages) == 0 {
		return
	}

	var uncached []string
	for _, pkg := range packages {
		if _, ok := f.aurCache[pkg]; !ok {
			uncached = append(uncached, pkg)
		}
	}

	if len(uncached) == 0 {
		fmt.Fprintf(os.Stderr, "[debug] ensureAURCache: done (%.2fs)\n", time.Since(start).Seconds())
		return
	}

	f.fetchAURInfo(uncached)
	fmt.Fprintf(os.Stderr, "[debug] ensureAURCache: done (%.2fs)\n", time.Since(start).Seconds())
}

func (f *Fetcher) fetchAURInfo(packages []string) map[string]AURPackage {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] fetchAURInfo: starting...\n")

	result := make(map[string]AURPackage)

	if len(packages) == 0 {
		return result
	}

	v := url.Values{}
	for _, pkg := range packages {
		v.Add("arg[]", pkg)
	}

	resp, err := http.Get(AURInfoURL + "&" + v.Encode())
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result
	}

	var aurResp AURResponse
	if err := json.Unmarshal(body, &aurResp); err != nil {
		return result
	}

	for _, r := range aurResp.Results {
		f.aurCache[r.Name] = r
		result[r.Name] = r
	}

	fmt.Fprintf(os.Stderr, "[debug] fetchAURInfo: done (%.2fs)\n", time.Since(start).Seconds())
	return result
}

func (f *Fetcher) ListOrphans() ([]string, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] ListOrphans: starting...\n")

	cmd := exec.Command("pacman", "-Qdtq")
	orphans, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	list := strings.TrimSpace(string(orphans))
	if list == "" {
		fmt.Fprintf(os.Stderr, "[debug] ListOrphans: done (%.2fs)\n", time.Since(start).Seconds())
		return nil, nil
	}

	fmt.Fprintf(os.Stderr, "[debug] ListOrphans: done (%.2fs)\n", time.Since(start).Seconds())
	return strings.Split(list, "\n"), nil
}
