package pacman

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Riyyi/declpac/pkg/fetch"
	"github.com/Riyyi/declpac/pkg/output"
	"github.com/Riyyi/declpac/pkg/state"
	"github.com/Riyyi/declpac/pkg/validation"
)

func MarkAllAsDeps() error {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] MarkAllAsDeps: starting...\n")

	cmd := exec.Command("pacman", "-D", "--asdeps")
	state.Write([]byte("MarkAllAsDeps...\n"))
	cmd.Stdout = io.MultiWriter(os.Stdout, state.GetLogWriter())
	cmd.Stderr = io.MultiWriter(os.Stderr, state.GetLogWriter())
	err := cmd.Run()
	if err != nil {
		state.Write([]byte(fmt.Sprintf("error: %v\n", err)))
	}

	fmt.Fprintf(os.Stderr, "[debug] MarkAllAsDeps: done (%.2fs)\n", time.Since(start).Seconds())
	return err
}

func MarkAsExplicit(packages []string) error {
	if len(packages) == 0 {
		return nil
	}
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] MarkAsExplicit: starting...\n")

	args := append([]string{"-D", "--asexplicit"}, packages...)
	cmd := exec.Command("pacman", args...)
	state.Write([]byte("MarkAsExplicit...\n"))
	cmd.Stdout = io.MultiWriter(os.Stdout, state.GetLogWriter())
	cmd.Stderr = io.MultiWriter(os.Stderr, state.GetLogWriter())
	err := cmd.Run()
	if err != nil {
		state.Write([]byte(fmt.Sprintf("error: %v\n", err)))
	}

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

	if err := validation.CheckDBFreshness(); err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: database fresh (%.2fs)\n", time.Since(start).Seconds())

	f, err := fetch.New()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fmt.Fprintf(os.Stderr, "[debug] Sync: initialized fetcher (%.2fs)\n", time.Since(start).Seconds())

	fmt.Fprintf(os.Stderr, "[debug] Sync: categorizing packages...\n")
	pacmanPkgs, aurPkgs, err := categorizePackages(f, packages)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: packages categorized (%.2fs)\n", time.Since(start).Seconds())

	if len(pacmanPkgs) > 0 {
		fmt.Fprintf(os.Stderr, "[debug] Sync: syncing %d pacman packages...\n", len(pacmanPkgs))
		_, err = SyncPackages(pacmanPkgs)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[debug] Sync: pacman packages synced (%.2fs)\n", time.Since(start).Seconds())
	}

	for _, pkg := range aurPkgs {
		fmt.Fprintf(os.Stderr, "[debug] Sync: installing AUR package %s...\n", pkg)
		if err := InstallAUR(f, pkg); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[debug] Sync: AUR package %s installed (%.2fs)\n", pkg, time.Since(start).Seconds())
	}

	fmt.Fprintf(os.Stderr, "[debug] Sync: marking all as deps...\n")
	if err := MarkAllAsDeps(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not mark all as deps: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: all marked as deps (%.2fs)\n", time.Since(start).Seconds())

	fmt.Fprintf(os.Stderr, "[debug] Sync: marking state packages as explicit...\n")
	if err := MarkAsExplicit(packages); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not mark state packages as explicit: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: state packages marked as explicit (%.2fs)\n", time.Since(start).Seconds())

	removed, err := CleanupOrphans()
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

func categorizePackages(f *fetch.Fetcher, packages []string) (pacmanPkgs, aurPkgs []string, err error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] categorizePackages: starting...\n")

	resolved, err := f.Resolve(packages)
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

func InstallAUR(f *fetch.Fetcher, pkgName string) error {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] InstallAUR: starting...\n")

	aurInfo, ok := f.GetAURPackage(pkgName)
	if !ok {
		return fmt.Errorf("AUR package not found in cache: %s", pkgName)
	}

	tmpDir, err := os.MkdirTemp("", "declpac-aur-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneURL := "https://aur.archlinux.org/" + aurInfo.PackageBase + ".git"
	state.Write([]byte("Cloning " + cloneURL + "\n"))
	cloneCmd := exec.Command("git", "clone", cloneURL, tmpDir)
	cloneCmd.Stdout = io.MultiWriter(os.Stdout, state.GetLogWriter())
	cloneCmd.Stderr = io.MultiWriter(os.Stderr, state.GetLogWriter())
	if err := cloneCmd.Run(); err != nil {
		errMsg := fmt.Sprintf("failed to clone AUR repo: %v\n", err)
		state.Write([]byte("error: " + errMsg))
		return fmt.Errorf("failed to clone AUR repo: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] InstallAUR: cloned (%.2fs)\n", time.Since(start).Seconds())

	state.Write([]byte("Building package...\n"))
	makepkgCmd := exec.Command("makepkg", "-si", "--noconfirm")
	makepkgCmd.Stdout = io.MultiWriter(os.Stdout, state.GetLogWriter())
	makepkgCmd.Stderr = io.MultiWriter(os.Stderr, state.GetLogWriter())
	makepkgCmd.Dir = tmpDir
	if err := makepkgCmd.Run(); err != nil {
		errMsg := fmt.Sprintf("makepkg failed to build AUR package: %v\n", err)
		state.Write([]byte("error: " + errMsg))
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

func SyncPackages(packages []string) (int, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] SyncPackages: starting...\n")

	args := append([]string{"-S", "--needed"}, packages...)
	cmd := exec.Command("pacman", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("pacman sync failed: %s", output)
		state.Write([]byte(errMsg))
		return 0, fmt.Errorf("pacman sync failed: %s", output)
	}

	if len(output) > 0 {
		state.Write(output)
	}

	re := regexp.MustCompile(`upgrading (\S+)`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	fmt.Fprintf(os.Stderr, "[debug] SyncPackages: done (%.2fs)\n", time.Since(start).Seconds())
	return len(matches), nil
}

func CleanupOrphans() (int, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] CleanupOrphans: starting...\n")

	f, err := fetch.New()
	if err != nil {
		return 0, err
	}
	defer f.Close()

	orphans, err := f.ListOrphans()
	if err != nil || len(orphans) == 0 {
		fmt.Fprintf(os.Stderr, "[debug] CleanupOrphans: done (%.2fs)\n", time.Since(start).Seconds())
		return 0, nil
	}

	removeCmd := exec.Command("pacman", "-Rns")
	output, err := removeCmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("%s: %s", err, output)
		state.Write([]byte(errMsg))
		return 0, fmt.Errorf("%s: %s", err, output)
	}

	if len(output) > 0 {
		state.Write(output)
	}

	count := len(orphans)

	fmt.Fprintf(os.Stderr, "[debug] CleanupOrphans: done (%.2fs)\n", time.Since(start).Seconds())
	return count, nil
}

func DryRun(packages []string) (*output.Result, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] DryRun: starting...\n")

	f, err := fetch.New()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fmt.Fprintf(os.Stderr, "[debug] DryRun: initialized fetcher (%.2fs)\n", time.Since(start).Seconds())

	resolved, err := f.Resolve(packages)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[debug] DryRun: packages resolved (%.2fs)\n", time.Since(start).Seconds())

	localPkgs, err := f.BuildLocalPkgMap()
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
		if info.InAUR {
			aurPkgs = append(aurPkgs, pkg)
		} else if _, installed := localPkgs[pkg]; !installed {
			toInstall = append(toInstall, pkg)
		}
	}
	fmt.Fprintf(os.Stderr, "[debug] DryRun: packages categorized (%.2fs)\n", time.Since(start).Seconds())

	orphans, err := f.ListOrphans()
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
