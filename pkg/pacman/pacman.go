package pacman

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Riyyi/declpac/pkg/output"
)

var (
	Root     = "/"
	LockFile = "/var/lib/pacman/db.lock"
)

type Pac struct{}

func New() (*Pac, error) {
	return &Pac{}, nil
}

func (p *Pac) Close() error {
	return nil
}

type PackageInfo struct {
	Name   string
	InAUR  bool
	Exists bool
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

	cmd = exec.Command("aur", "search", name)
	if out, err := cmd.Output(); err == nil && len(out) > 0 {
		return &PackageInfo{Name: name, Exists: true, InAUR: true}, nil
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

	for _, pkg := range packages {
		if err := p.MarkExplicit(pkg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not mark %s as explicit: %v\n", pkg, err)
		}
	}

	_, err = p.SyncPackages(packages)
	if err != nil {
		return nil, err
	}

	removed, err := p.CleanupOrphans()
	if err != nil {
		return nil, err
	}

	after, _ := getInstalledCount()
	installedCount := max(after - before, 0)

	return &output.Result{
		Installed: installedCount,
		Removed:   removed,
	}, nil
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
