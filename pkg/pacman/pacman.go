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

	"github.com/Riyyi/declpac/pkg/output"
)

var (
	Root       = "/"
	LockFile   = "/var/lib/pacman/db.lock"
	AURInfoURL = "https://aur.archlinux.org/rpc?v=5&type=info"
)

type Pac struct {
	aurCache map[string]AURPackage
}

func New() (*Pac, error) {
	return &Pac{aurCache: make(map[string]AURPackage)}, nil
}

func (p *Pac) Close() error {
	return nil
}

type PackageInfo struct {
	Name    string
	InAUR   bool
	Exists  bool
	AURInfo *AURPackage
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

func (p *Pac) ValidatePackage(name string) (*PackageInfo, error) {
	cmd := exec.Command("pacman", "-Qip", name)
	if err := cmd.Run(); err == nil {
		return &PackageInfo{Name: name, Exists: true, InAUR: false}, nil
	}

	cmd = exec.Command("pacman", "-Sip", name)
	if err := cmd.Run(); err == nil {
		return &PackageInfo{Name: name, Exists: true, InAUR: false}, nil
	}

	p.ensureAURCache([]string{name})
	if aurInfo, ok := p.aurCache[name]; ok {
		return &PackageInfo{Name: name, Exists: true, InAUR: true, AURInfo: &aurInfo}, nil
	}

	return &PackageInfo{Name: name, Exists: false, InAUR: false}, nil
}

func (p *Pac) IsDBFresh() (bool, error) {
	info, err := os.Stat(LockFile)
	if err != nil {
		return false, nil
	}

	age := time.Since(info.ModTime())
	return age < 24*time.Hour, nil
}

func (p *Pac) SyncDB() error {
	cmd := exec.Command("pacman", "-Syy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *Pac) GetInstalledPackages() ([]string, error) {
	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	packages := strings.Split(strings.TrimSpace(string(output)), "\n")
	return packages, nil
}

func (p *Pac) MarkExplicit(pkgName string) error {
	cmd := exec.Command("pacman", "-D", "--explicit", pkgName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Sync(packages []string) (*output.Result, error) {
	before, err := getInstalledCount()
	if err != nil {
		return nil, err
	}

	p, err := New()
	if err != nil {
		return nil, err
	}
	defer p.Close()

	fresh, err := p.IsDBFresh()
	if err != nil || !fresh {
		if err := p.SyncDB(); err != nil {
			return nil, fmt.Errorf("failed to sync database: %w", err)
		}
	}

	pacmanPkgs, aurPkgs := p.categorizePackages(packages)

	for _, pkg := range packages {
		if err := p.MarkExplicit(pkg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not mark %s as explicit: %v\n", pkg, err)
		}
	}

	if len(pacmanPkgs) > 0 {
		_, err = p.SyncPackages(pacmanPkgs)
		if err != nil {
			return nil, err
		}
	}

	for _, pkg := range aurPkgs {
		if err := p.InstallAUR(pkg); err != nil {
			return nil, err
		}
	}

	removed, err := p.CleanupOrphans()
	if err != nil {
		return nil, err
	}

	after, _ := getInstalledCount()
	installedCount := max(after-before, 0)

	return &output.Result{
		Installed: installedCount,
		Removed:   removed,
	}, nil
}

func (p *Pac) categorizePackages(packages []string) (pacmanPkgs, aurPkgs []string) {
	var notInPacman []string

	for _, pkg := range packages {
		info, err := p.ValidatePackage(pkg)
		if err != nil || !info.Exists {
			notInPacman = append(notInPacman, pkg)
		} else if !info.InAUR {
			pacmanPkgs = append(pacmanPkgs, pkg)
		}
	}

	if len(notInPacman) > 0 {
		p.ensureAURCache(notInPacman)
		for _, pkg := range notInPacman {
			if _, ok := p.aurCache[pkg]; ok {
				aurPkgs = append(aurPkgs, pkg)
			} else {
				fmt.Fprintf(os.Stderr, "error: package not found: %s\n", pkg)
			}
		}
	}

	return pacmanPkgs, aurPkgs
}

func (p *Pac) ensureAURCache(packages []string) {
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
		return
	}

	p.fetchAURInfo(uncached)
}

func (p *Pac) fetchAURInfo(packages []string) map[string]AURPackage {
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

	return result
}

func (p *Pac) InstallAUR(pkgName string) error {
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

	makepkgCmd := exec.Command("makepkg", "-si", "--noconfirm")
	makepkgCmd.Stdout = os.Stdout
	makepkgCmd.Stderr = os.Stderr
	makepkgCmd.Dir = tmpDir
	if err := makepkgCmd.Run(); err != nil {
		return fmt.Errorf("makepkg failed to build AUR package: %w", err)
	}

	return nil
}

func getInstalledCount() (int, error) {
	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		return 0, nil
	}
	count := strings.Count(string(output), "\n") + 1
	if strings.TrimSpace(string(output)) == "" {
		count = 0
	}
	return count, nil
}

func (p *Pac) SyncPackages(packages []string) (int, error) {
	args := append([]string{"-Syu"}, packages...)
	cmd := exec.Command("pacman", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("pacman sync failed: %s", output)
	}

	re := regexp.MustCompile(`upgrading (\S+)`)
	matches := re.FindAllStringSubmatch(string(output), -1)
	return len(matches), nil
}

func (p *Pac) CleanupOrphans() (int, error) {
	listCmd := exec.Command("pacman", "-Qdtq")
	orphans, err := listCmd.Output()
	if err != nil {
		return 0, nil
	}

	orphanList := strings.TrimSpace(string(orphans))
	if orphanList == "" {
		return 0, nil
	}

	removeCmd := exec.Command("pacman", "-Rns")
	output, err := removeCmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%s: %s", err, output)
	}

	count := strings.Count(orphanList, "\n") + 1
	return count, nil
}

func DryRun(packages []string) (*output.Result, error) {
	p, err := New()
	if err != nil {
		return nil, err
	}
	defer p.Close()

	current, err := p.GetInstalledPackages()
	if err != nil {
		return nil, err
	}
	currentSet := make(map[string]bool)
	for _, pkg := range current {
		currentSet[pkg] = true
	}

	var toInstall []string
	var aurPkgs []string
	for _, pkg := range packages {
		if !currentSet[pkg] {
			info, err := p.ValidatePackage(pkg)
			if err != nil || !info.Exists {
				return nil, fmt.Errorf("package not found: %s", pkg)
			}
			if info.InAUR {
				aurPkgs = append(aurPkgs, pkg)
			} else {
				toInstall = append(toInstall, pkg)
			}
		}
	}

	orphans, err := p.listOrphans()
	if err != nil {
		return nil, err
	}

	return &output.Result{
		Installed: len(toInstall) + len(aurPkgs),
		Removed:   len(orphans),
		ToInstall: append(toInstall, aurPkgs...),
		ToRemove:  orphans,
	}, nil
}

func (p *Pac) listOrphans() ([]string, error) {
	cmd := exec.Command("pacman", "-Qdtq")
	orphans, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	list := strings.TrimSpace(string(orphans))
	if list == "" {
		return nil, nil
	}

	return strings.Split(list, "\n"), nil
}
