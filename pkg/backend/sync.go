package backend

import (
	"fmt"
	"slices"
	"time"

	"github.com/Riyyi/declpac/pkg/log"
	"github.com/Riyyi/declpac/pkg/output"
)

// -----------------------------------------
// Public functions

func Sync(b Backend, packages []string, noCheck bool, prune bool) (*output.Result, error) {
	start := time.Now()
	log.Debug("Sync: starting...")

	explicitList, err := b.ListExplicit()
	if err != nil {
		return nil, err
	}
	explicitCount := len(explicitList)

	if !noCheck && len(packages) < explicitCount/2 {
		errMsg := "safety check: state packages (%d) less than half of explicitly installed (%d), override with --nocheck"
		return nil, fmt.Errorf(errMsg, len(packages), explicitCount)
	}

	list, err := b.List()
	if err != nil {
		return nil, err
	}
	before := len(list)

	if err := b.RefreshDB(log.GetLogWriter()); err != nil {
		return nil, err
	}
	log.Debug("Sync: database refreshed (%.2fs)", time.Since(start).Seconds())

	resolved, err := b.Resolve(packages)
	if err != nil {
		return nil, err
	}
	log.Debug("Sync: packages resolved (%.2fs)", time.Since(start).Seconds())

	if err := b.InstallPackages(resolved, log.GetLogWriter()); err != nil {
		return nil, err
	}
	log.Debug("Sync: packages installed (%.2fs)", time.Since(start).Seconds())

	var removed int
	if prune {
		log.Debug("Sync: running prune sanity check...")
		if err := pruneSanityCheck(b, packages); err != nil {
			return nil, err
		}
		log.Debug("Sync: prune sanity check passed (%.2fs)", time.Since(start).Seconds())

		removed, err = b.Prune(packages, log.GetLogWriter())
		if err != nil {
			return nil, err
		}
		log.Debug("Sync: prune done (%.2fs)", time.Since(start).Seconds())
	}

	list, err = b.List()
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
// Private functions

func pruneSanityCheck(b Backend, statePackages []string) error {
	start := time.Now()
	log.Debug("pruneSanityCheck: starting...")

	localPackages, err := b.List()
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