package pacman

import (
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/Riyyi/declpac/pkg/fetch"
	"github.com/Riyyi/declpac/pkg/log"
	"github.com/Riyyi/declpac/pkg/output"
	"github.com/Riyyi/declpac/pkg/pacman/read"
	"github.com/Riyyi/declpac/pkg/pacman/sync"
)

func Sync(packages []string, noCheck bool, prune bool) (*output.Result, error) {
	start := time.Now()
	log.Debug("Sync: starting...")

	explicitList, err := read.ExplicitList()
	if err != nil {
		return nil, err
	}
	explicitCount := len(explicitList)

	if !noCheck && len(packages) < explicitCount/2 {
		errMsg := "safety check: state packages (%d) less than half of explicitly installed (%d), override with --nocheck"
		return nil, fmt.Errorf(errMsg, len(packages), explicitCount)
	}

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
	log.Debug("Sync: database fresh (%.2fs)", time.Since(start).Seconds())

	f, err := fetch.New()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Debug("Sync: initialized fetcher (%.2fs)", time.Since(start).Seconds())

	log.Debug("Sync: categorizing packages...")
	pacmanPkgs, aurPkgs, err := categorizePackages(f, packages)
	if err != nil {
		return nil, err
	}
	log.Debug("Sync: packages categorized (%.2fs)", time.Since(start).Seconds())

	if len(pacmanPkgs) > 0 {
		log.Debug("Sync: syncing %d pacman packages...", len(pacmanPkgs))
		if err := sync.SyncPackages(pacmanPkgs, log.GetLogWriter()); err != nil {
			return nil, err
		}
		log.Debug("Sync: pacman packages synced (%.2fs)", time.Since(start).Seconds())
	}

	for _, pkg := range aurPkgs {
		if slices.Contains(list, pkg) {
			continue
		}
		log.Debug("Sync: installing AUR package %s...", pkg)
		aurInfo, ok := f.GetAURPackage(pkg)
		if !ok {
			return nil, fmt.Errorf("AUR package not found in cache: %s", pkg)
		}
		if err := sync.InstallAUR(f, pkg, aurInfo.PackageBase, false, log.GetLogWriter()); err != nil {
			return nil, err
		}
		log.Debug("Sync: AUR package %s installed (%.2fs)", pkg, time.Since(start).Seconds())
	}

	var removed int
	if prune {
		log.Debug("Sync: running prune sanity check...")
		if err := pruneSanityCheck(packages); err != nil {
			return nil, err
		}
		log.Debug("Sync: prune sanity check passed (%.2fs)", time.Since(start).Seconds())

		log.Debug("Sync: marking all as deps...")
		if err := markAllAsDeps(); err != nil {
			return nil, err
		}
		log.Debug("Sync: all marked as deps (%.2fs)", time.Since(start).Seconds())

		log.Debug("Sync: marking state packages as explicit...")
		if err := sync.MarkAs(packages, "explicit", log.GetLogWriter()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not mark state packages as explicit: %v\n", err)
		}
		log.Debug("Sync: state packages marked as explicit (%.2fs)", time.Since(start).Seconds())

		removed, err = cleanupOrphans()
		if err != nil {
			return nil, err
		}
	}

	list, _ = read.List()
	if err != nil {
		return nil, err
	}
	after := len(list)

	installedCount := max(after-before, 0)

	log.Debug("Sync: done (%.2fs)", time.Since(start).Seconds())
	return &output.Result{
		Installed: installedCount,
		Removed:   removed,
	}, nil
}

// -----------------------------------------
// private

func categorizePackages(f *fetch.Fetcher, packages []string) (pacmanPkgs, aurPkgs []string, err error) {
	start := time.Now()
	log.Debug("categorizePackages: starting...")

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

	log.Debug("categorizePackages: done (%.2fs)", time.Since(start).Seconds())
	return pacmanPkgs, aurPkgs, nil
}

func cleanupOrphans() (int, error) {
	start := time.Now()
	log.Debug("cleanupOrphans: starting...")

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

	log.Debug("cleanupOrphans: done (%.2fs)", time.Since(start).Seconds())
	return removed, nil
}

func markAllAsDeps() error {
	start := time.Now()
	log.Debug("markAllAsDeps: starting...")

	packages, err := read.List()
	if err != nil || len(packages) == 0 {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if err := sync.MarkAs(packages, "deps", log.GetLogWriter()); err != nil {
		log.Write([]byte(fmt.Sprintf("error: %v\n", err)))
		return err
	}

	log.Debug("markAllAsDeps: done (%.2fs)", time.Since(start).Seconds())
	return nil
}

// pruneSanityCheck checks if the installation of all state packages succeeded,
// before attempting to do package marking and orphan cleanup.
func pruneSanityCheck(statePackages []string) error {
	start := time.Now()
	log.Debug("pruneSanityCheck: starting...")

	localPackages, err := read.List()
	if err != nil {
		return fmt.Errorf("failed to list local packages: %w", err)
	}

	var missing []string
	for _, pkg := range statePackages {
		if !slices.Contains(localPackages, pkg) {
			missing = append(missing, pkg)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("safety check: missing state packages: %v", missing)
	}

	log.Debug("pruneSanityCheck: done (%.2fs)", time.Since(start).Seconds())
	return nil
}
