package pacman

import (
	"fmt"
	"os"
	"time"

	"github.com/Riyyi/declpac/pkg/fetch"
	"github.com/Riyyi/declpac/pkg/log"
	"github.com/Riyyi/declpac/pkg/output"
	"github.com/Riyyi/declpac/pkg/pacman/read"
	"github.com/Riyyi/declpac/pkg/pacman/sync"
)

func Sync(packages []string) (*output.Result, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] Sync: starting...\n")

	list, err := read.List()
	if err != nil {
		return nil, err
	}
	before := len(list)

	fresh, err := read.DBFreshness()
	if err != nil {
		return nil, err
	}
	if !fresh {
		if err := sync.RefreshDB(log.GetLogWriter()); err != nil {
			return nil, err
		}
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
		if err := sync.SyncPackages(pacmanPkgs, log.GetLogWriter()); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[debug] Sync: pacman packages synced (%.2fs)\n", time.Since(start).Seconds())
	}

	for _, pkg := range aurPkgs {
		fmt.Fprintf(os.Stderr, "[debug] Sync: installing AUR package %s...\n", pkg)
		aurInfo, ok := f.GetAURPackage(pkg)
		if !ok {
			return nil, fmt.Errorf("AUR package not found in cache: %s", pkg)
		}
		if err := sync.InstallAUR(pkg, aurInfo.PackageBase, log.GetLogWriter()); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[debug] Sync: AUR package %s installed (%.2fs)\n", pkg, time.Since(start).Seconds())
	}

	fmt.Fprintf(os.Stderr, "[debug] Sync: marking all as deps...\n")
	markAllAsDeps()
	fmt.Fprintf(os.Stderr, "[debug] Sync: all marked as deps (%.2fs)\n", time.Since(start).Seconds())

	fmt.Fprintf(os.Stderr, "[debug] Sync: marking state packages as explicit...\n")
	if err := sync.MarkAs(packages, "explicit", log.GetLogWriter()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not mark state packages as explicit: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "[debug] Sync: state packages marked as explicit (%.2fs)\n", time.Since(start).Seconds())

	removed, err := cleanupOrphans()
	if err != nil {
		return nil, err
	}

	list, _ = read.List()
	if err != nil {
		return nil, err
	}
	after := len(list)

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
		if info == nil || (!info.Exists && !info.InAUR) {
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

func markAllAsDeps() error {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] markAllAsDeps: starting...\n")

	packages, err := read.List()
	if err != nil || len(packages) == 0 {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if err := sync.MarkAs(packages, "deps", log.GetLogWriter()); err != nil {
		log.Write([]byte(fmt.Sprintf("error: %v\n", err)))
		return err
	}

	fmt.Fprintf(os.Stderr, "[debug] markAllAsDeps: done (%.2fs)\n", time.Since(start).Seconds())
	return nil
}

func cleanupOrphans() (int, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] cleanupOrphans: starting...\n")

	orphans, err := read.ListOrphans()
	if err != nil {
		log.Write([]byte(fmt.Sprintf("error: %v\n", err)))
		return 0, err
	}

	removed, err := sync.RemoveOrphans(orphans, log.GetLogWriter())
	if err != nil {
		log.Write([]byte(fmt.Sprintf("error: %v\n", err)))
		return 0, err
	}

	fmt.Fprintf(os.Stderr, "[debug] cleanupOrphans: done (%.2fs)\n", time.Since(start).Seconds())
	return removed, nil
}
