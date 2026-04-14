package pacman

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Jguer/dyalpm"
	"github.com/Riyyi/declpac/pkg/output"
)

var (
	Root       = "/"
	LockFile   = "/var/lib/pacman/db.lock"
	AURInfoURL = "https://aur.archlinux.org/rpc?v=5&type=info"
)

type Pac struct {
	aurCache map[string]AURPackage
	handle   dyalpm.Handle
	localDB  dyalpm.Database
	syncDBs  []dyalpm.Database
}

func New() (*Pac, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] New: starting...\n")

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

	fmt.Fprintf(os.Stderr, "[debug] New: done (%.2fs)\n", time.Since(start).Seconds())
	return &Pac{
		aurCache: make(map[string]AURPackage),
		handle:   handle,
		localDB:  localDB,
		syncDBs:  syncDBs,
	}, nil
}

func (p *Pac) Close() error {
	if p.handle != nil {
		p.handle.Release()
	}
	return nil
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

type PackageInfo struct {
	Name      string
	InAUR     bool
	Exists    bool
	Installed bool
	AURInfo   *AURPackage
	syncPkg   dyalpm.Package
}

func (p *Pac) buildLocalPkgMap() (map[string]dyalpm.Package, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] buildLocalPkgMap: starting...\n")

	localPkgs := make(map[string]dyalpm.Package)

	err := p.localDB.PkgCache().ForEach(func(pkg dyalpm.Package) error {
		localPkgs[pkg.Name()] = pkg
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate local package cache: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[debug] buildLocalPkgMap: done (%.2fs)\n", time.Since(start).Seconds())
	return localPkgs, nil
}

func (p *Pac) checkSyncDBs(pkgNames []string) (map[string]dyalpm.Package, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] checkSyncDBs: starting...\n")

	syncPkgs := make(map[string]dyalpm.Package)
	pkgSet := make(map[string]bool)
	for _, name := range pkgNames {
		pkgSet[name] = true
	}

	for _, db := range p.syncDBs {
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

func (p *Pac) resolvePackages(packages []string) (map[string]*PackageInfo, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] resolvePackages: starting...\n")

	result := make(map[string]*PackageInfo)

	localPkgs, err := p.buildLocalPkgMap()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] resolvePackages: local pkgs built (%.2fs)\n", time.Since(start).Seconds())

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
		syncPkgs, err := p.checkSyncDBs(notInLocal)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[debug] resolvePackages: sync db checked (%.2fs)\n", time.Since(start).Seconds())

		var notInSync []string
		for _, pkg := range notInLocal {
			if syncPkg, ok := syncPkgs[pkg]; ok {
				result[pkg] = &PackageInfo{
					Name:      pkg,
					Exists:    true,
					InAUR:     false,
					Installed: false,
					syncPkg:   syncPkg,
				}
			} else {
				notInSync = append(notInSync, pkg)
			}
		}

		if len(notInSync) > 0 {
			p.ensureAURCache(notInSync)
			fmt.Fprintf(os.Stderr, "[debug] resolvePackages: AUR cache ensured (%.2fs)\n", time.Since(start).Seconds())

			var unfound []string
			for _, pkg := range notInSync {
				if aurInfo, ok := p.aurCache[pkg]; ok {
					result[pkg] = &PackageInfo{
						Name:      pkg,
						Exists:    true,
						InAUR:     true,
						Installed: false,
						AURInfo:   &aurInfo,
					}
				} else {
					unfound = append(unfound, pkg)
				}
			}
			if len(unfound) > 0 {
				return nil, fmt.Errorf("package(s) not found: %s", strings.Join(unfound, ", "))
			}
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] resolvePackages: done (%.2fs)\n", time.Since(start).Seconds())
	return result, nil
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

func (p *Pac) IsDBFresh() (bool, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] IsDBFresh: starting...\n")

	info, err := os.Stat(LockFile)
	if err != nil {
		return false, nil
	}

	age := time.Since(info.ModTime())
	fmt.Fprintf(os.Stderr, "[debug] IsDBFresh: done (%.2fs)\n", time.Since(start).Seconds())
	return age < 24*time.Hour, nil
}

func (p *Pac) SyncDB() error {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] SyncDB: starting...\n")

	cmd := exec.Command("pacman", "-Syy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	fmt.Fprintf(os.Stderr, "[debug] SyncDB: done (%.2fs)\n", time.Since(start).Seconds())
	return err
}

func (p *Pac) MarkAllAsDeps() error {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] MarkAllAsDeps: starting...\n")

	cmd := exec.Command("pacman", "-D", "--asdeps")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	fmt.Fprintf(os.Stderr, "[debug] MarkAllAsDeps: done (%.2fs)\n", time.Since(start).Seconds())
	return err
}

func (p *Pac) MarkAsExplicit(packages []string) error {
	if len(packages) == 0 {
		return nil
	}
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] MarkAsExplicit: starting...\n")

	args := append([]string{"-D", "--asexplicit"}, packages...)
	cmd := exec.Command("pacman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	fmt.Fprintf(os.Stderr, "[debug] MarkAsExplicit: done (%.2fs)\n", time.Since(start).Seconds())
	return err
}

func Sync(packages []string) (*output.Result, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] Sync: starting...\n")

	before, err := getInstalledCount()
	if err != nil {
		return nil, err
	}

	p, err := New()
	if err != nil {
		return nil, err
	}
	defer p.Close()
	fmt.Fprintf(os.Stderr, "[debug] Sync: initialized pacman (%.2fs)\n", time.Since(start).Seconds())

	fresh, err := p.IsDBFresh()
	if err != nil || !fresh {
		fmt.Fprintf(os.Stderr, "[debug] Sync: syncing database...\n")
		if err := p.SyncDB(); err != nil {
			return nil, fmt.Errorf("failed to sync database: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[debug] Sync: database synced (%.2fs)\n", time.Since(start).Seconds())
	}

	fmt.Fprintf(os.Stderr, "[debug] Sync: categorizing packages...\n")
	pacmanPkgs, aurPkgs, err := p.categorizePackages(packages)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: packages categorized (%.2fs)\n", time.Since(start).Seconds())

	if len(pacmanPkgs) > 0 {
		fmt.Fprintf(os.Stderr, "[debug] Sync: syncing %d pacman packages...\n", len(pacmanPkgs))
		_, err = p.SyncPackages(pacmanPkgs)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[debug] Sync: pacman packages synced (%.2fs)\n", time.Since(start).Seconds())
	}

	for _, pkg := range aurPkgs {
		fmt.Fprintf(os.Stderr, "[debug] Sync: installing AUR package %s...\n", pkg)
		if err := p.InstallAUR(pkg); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[debug] Sync: AUR package %s installed (%.2fs)\n", pkg, time.Since(start).Seconds())
	}

	fmt.Fprintf(os.Stderr, "[debug] Sync: marking all as deps...\n")
	if err := p.MarkAllAsDeps(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not mark all as deps: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: all marked as deps (%.2fs)\n", time.Since(start).Seconds())

	fmt.Fprintf(os.Stderr, "[debug] Sync: marking state packages as explicit...\n")
	if err := p.MarkAsExplicit(packages); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not mark state packages as explicit: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: state packages marked as explicit (%.2fs)\n", time.Since(start).Seconds())

	removed, err := p.CleanupOrphans()
	if err != nil {
		return nil, err
	}

	after, _ := getInstalledCount()
	installedCount := max(after-before, 0)

	fmt.Fprintf(os.Stderr, "[debug] Sync: done (%.2fs)\n", time.Since(start).Seconds())
	return &output.Result{
		Installed: installedCount,
		Removed:   removed,
	}, nil
}

func (p *Pac) categorizePackages(packages []string) (pacmanPkgs, aurPkgs []string, err error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] categorizePackages: starting...\n")

	resolved, err := p.resolvePackages(packages)
	if err != nil {
		return nil, nil, err
	}

	for _, pkg := range packages {
		info := resolved[pkg]
		if info == nil || !info.Exists {
			fmt.Fprintf(os.Stderr, "error: package not found: %s\n", pkg)
			continue
		}
		if info.InAUR {
			aurPkgs = append(aurPkgs, pkg)
		} else {
			pacmanPkgs = append(pacmanPkgs, pkg)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] categorizePackages: done (%.2fs)\n", time.Since(start).Seconds())
	return pacmanPkgs, aurPkgs, nil
}

func (p *Pac) ensureAURCache(packages []string) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] ensureAURCache: starting...\n")

	if len(packages) == 0 {
		return
	}

	var uncached []string
	for _, pkg := range packages {
		if _, ok := p.aurCache[pkg]; !ok {
			uncached = append(uncached, pkg)
		}
	}

	if len(uncached) == 0 {
		fmt.Fprintf(os.Stderr, "[debug] ensureAURCache: done (%.2fs)\n", time.Since(start).Seconds())
		return
	}

	p.fetchAURInfo(uncached)
	fmt.Fprintf(os.Stderr, "[debug] ensureAURCache: done (%.2fs)\n", time.Since(start).Seconds())
}

func (p *Pac) fetchAURInfo(packages []string) map[string]AURPackage {
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
		p.aurCache[r.Name] = r
		result[r.Name] = r
	}

	fmt.Fprintf(os.Stderr, "[debug] fetchAURInfo: done (%.2fs)\n", time.Since(start).Seconds())
	return result
}

func (p *Pac) InstallAUR(pkgName string) error {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] InstallAUR: starting...\n")

	aurInfo, ok := p.aurCache[pkgName]
	if !ok {
		return fmt.Errorf("AUR package not found in cache: %s", pkgName)
	}

	tmpDir, err := os.MkdirTemp("", "declpac-aur-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneURL := "https://aur.archlinux.org/" + aurInfo.PackageBase + ".git"
	cloneCmd := exec.Command("git", "clone", cloneURL, tmpDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone AUR repo: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] InstallAUR: cloned (%.2fs)\n", time.Since(start).Seconds())

	makepkgCmd := exec.Command("makepkg", "-si", "--noconfirm")
	makepkgCmd.Stdout = os.Stdout
	makepkgCmd.Stderr = os.Stderr
	makepkgCmd.Dir = tmpDir
	if err := makepkgCmd.Run(); err != nil {
		return fmt.Errorf("makepkg failed to build AUR package: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] InstallAUR: built (%.2fs)\n", time.Since(start).Seconds())

	fmt.Fprintf(os.Stderr, "[debug] InstallAUR: done (%.2fs)\n", time.Since(start).Seconds())
	return nil
}

func getInstalledCount() (int, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] getInstalledCount: starting...\n")

	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		return 0, nil
	}
	count := strings.Count(string(output), "\n") + 1
	if strings.TrimSpace(string(output)) == "" {
		count = 0
	}

	fmt.Fprintf(os.Stderr, "[debug] getInstalledCount: done (%.2fs)\n", time.Since(start).Seconds())
	return count, nil
}

func (p *Pac) SyncPackages(packages []string) (int, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] SyncPackages: starting...\n")

	args := append([]string{"-Syu"}, packages...)
	cmd := exec.Command("pacman", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("pacman sync failed: %s", output)
	}

	re := regexp.MustCompile(`upgrading (\S+)`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	fmt.Fprintf(os.Stderr, "[debug] SyncPackages: done (%.2fs)\n", time.Since(start).Seconds())
	return len(matches), nil
}

func (p *Pac) CleanupOrphans() (int, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] CleanupOrphans: starting...\n")

	listCmd := exec.Command("pacman", "-Qdtq")
	orphans, err := listCmd.Output()
	if err != nil {
		return 0, nil
	}

	orphanList := strings.TrimSpace(string(orphans))
	if orphanList == "" {
		fmt.Fprintf(os.Stderr, "[debug] CleanupOrphans: done (%.2fs)\n", time.Since(start).Seconds())
		return 0, nil
	}

	removeCmd := exec.Command("pacman", "-Rns")
	output, err := removeCmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%s: %s", err, output)
	}

	count := strings.Count(orphanList, "\n") + 1

	fmt.Fprintf(os.Stderr, "[debug] CleanupOrphans: done (%.2fs)\n", time.Since(start).Seconds())
	return count, nil
}

func DryRun(packages []string) (*output.Result, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] DryRun: starting...\n")

	p, err := New()
	if err != nil {
		return nil, err
	}
	defer p.Close()
	fmt.Fprintf(os.Stderr, "[debug] DryRun: initialized pacman (%.2fs)\n", time.Since(start).Seconds())

	resolved, err := p.resolvePackages(packages)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] DryRun: packages resolved (%.2fs)\n", time.Since(start).Seconds())

	localPkgs, err := p.buildLocalPkgMap()
	if err != nil {
		return nil, err
	}

	var toInstall []string
	var aurPkgs []string
	for _, pkg := range packages {
		info := resolved[pkg]
		if info == nil || !info.Exists {
			return nil, fmt.Errorf("package not found: %s", pkg)
		}
		if _, installed := localPkgs[pkg]; !installed {
			if info.InAUR {
				aurPkgs = append(aurPkgs, pkg)
			} else {
				toInstall = append(toInstall, pkg)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "[debug] DryRun: packages categorized (%.2fs)\n", time.Since(start).Seconds())

	orphans, err := p.listOrphans()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] DryRun: orphans listed (%.2fs)\n", time.Since(start).Seconds())

	fmt.Fprintf(os.Stderr, "[debug] DryRun: done (%.2fs)\n", time.Since(start).Seconds())
	return &output.Result{
		Installed: len(toInstall) + len(aurPkgs),
		Removed:   len(orphans),
		ToInstall: append(toInstall, aurPkgs...),
		ToRemove:  orphans,
	}, nil
}

func (p *Pac) listOrphans() ([]string, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] listOrphans: starting...\n")

	cmd := exec.Command("pacman", "-Qdtq")
	orphans, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	list := strings.TrimSpace(string(orphans))
	if list == "" {
		fmt.Fprintf(os.Stderr, "[debug] listOrphans: done (%.2fs)\n", time.Since(start).Seconds())
		return nil, nil
	}

	fmt.Fprintf(os.Stderr, "[debug] listOrphans: done (%.2fs)\n", time.Since(start).Seconds())
	return strings.Split(list, "\n"), nil
}
