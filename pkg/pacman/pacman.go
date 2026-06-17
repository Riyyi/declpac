package pacman

import (
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/Riyyi/declpac/pkg/backend"
	"github.com/Riyyi/declpac/pkg/log"
	"github.com/Riyyi/declpac/pkg/output"
	"github.com/Riyyi/declpac/pkg/pacman/fetch"
	"github.com/Riyyi/declpac/pkg/pacman/read"
	"github.com/Riyyi/declpac/pkg/pacman/sync"
)

// -----------------------------------------
// Variables

func init() {
	backend.Register("pacman", func() backend.Backend {
		return New()
	})
}

// -----------------------------------------
// Types

type Pacman struct {
	resolved map[string]*fetch.PackageInfo
}

// -----------------------------------------
// Constructors

func New() *Pacman {
	return &Pacman{}
}

// -----------------------------------------
// Public functions

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
// Public methods

func (p *Pacman) InstallPackages(pkgs []backend.PackageInfo, logWriter io.Writer) error {
	start := time.Now()
	log.Debug("Pacman.InstallPackages: starting...")

	var repoPkgs []string
	var aurPkgs []string

	for _, pkg := range pkgs {
		if pkg.Installed {
			continue
		}
		switch pkg.Source {
		case "repo":
			name := pkg.Name
			if pkg.Provided != "" {
				name = pkg.Provided
			}
			repoPkgs = append(repoPkgs, name)
		case "aur":
			aurPkgs = append(aurPkgs, pkg.Name)
		}
	}

	if len(repoPkgs) > 0 {
		log.Debug("Pacman.InstallPackages: syncing %d repo packages...", len(repoPkgs))
		if err := sync.SyncPackages(repoPkgs, logWriter); err != nil {
			return err
		}
		log.Debug("Pacman.InstallPackages: repo packages synced (%.2fs)", time.Since(start).Seconds())
	}

	if len(aurPkgs) > 0 {
		f, err := fetch.New()
		if err != nil {
			return err
		}
		defer f.Close()

		for _, pkg := range aurPkgs {
			log.Debug("Pacman.InstallPackages: installing AUR package %s...", pkg)
			aurInfo, ok := f.GetAURPackage(pkg)
			if !ok {
				return fmt.Errorf("AUR package not found in cache: %s", pkg)
			}
			if err := sync.InstallAUR(f, pkg, aurInfo.PackageBase, false, logWriter); err != nil {
				return err
			}
			log.Debug("Pacman.InstallPackages: AUR package %s installed (%.2fs)", pkg, time.Since(start).Seconds())
		}
	}

	log.Debug("Pacman.InstallPackages: done (%.2fs)", time.Since(start).Seconds())
	return nil
}

func (p *Pacman) List() ([]string, error) {
	return read.List()
}

func (p *Pacman) ListExplicit() ([]string, error) {
	return read.ExplicitList()
}

func (p *Pacman) ListOrphans() ([]string, error) {
	return read.ListOrphans()
}

func (p *Pacman) MarkAs(packages []string, flag string, logWriter io.Writer) error {
	return sync.MarkAs(packages, flag, logWriter)
}

func (p *Pacman) Name() string {
	return "pacman"
}

func (p *Pacman) Prune(statePackages []string, logWriter io.Writer) (int, error) {
	start := time.Now()
	log.Debug("Pacman.Prune: starting...")

	allPkgs, err := read.List()
	if err != nil || len(allPkgs) == 0 {
		return 0, fmt.Errorf("failed to list packages: %w", err)
	}

	if err := sync.MarkAs(allPkgs, "deps", logWriter); err != nil {
		return 0, err
	}
	log.Debug("Pacman.Prune: all marked as deps (%.2fs)", time.Since(start).Seconds())

	if err := sync.MarkAs(statePackages, "explicit", logWriter); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not mark state packages as explicit: %v\n", err)
	}
	log.Debug("Pacman.Prune: state packages marked as explicit (%.2fs)", time.Since(start).Seconds())

	orphans, err := read.ListOrphans()
	if err != nil {
		log.Write([]byte(fmt.Sprintf("error: %v\n", err)))
		return 0, err
	}

	removed, err := sync.RemoveOrphans(orphans, logWriter)
	if err != nil {
		log.Write([]byte(fmt.Sprintf("error: %v\n", err)))
		return 0, err
	}

	log.Debug("Pacman.Prune: done (%.2fs)", time.Since(start).Seconds())
	return removed, nil
}

func (p *Pacman) RefreshDB(logWriter io.Writer) error {
	fresh, err := read.DBFreshness()
	if err != nil {
		return err
	}
	if fresh {
		return nil
	}
	return sync.RefreshDB(logWriter)
}

func (p *Pacman) RemoveOrphans(orphans []string, logWriter io.Writer) (int, error) {
	return sync.RemoveOrphans(orphans, logWriter)
}

func (p *Pacman) Resolve(packages []string) ([]backend.PackageInfo, error) {
	start := time.Now()
	log.Debug("Pacman.Resolve: starting...")

	f, err := fetch.New()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	resolved, err := f.Resolve(packages)
	if err != nil {
		return nil, err
	}

	p.resolved = resolved

	result := make([]backend.PackageInfo, 0, len(packages))
	for _, pkg := range packages {
		info := resolved[pkg]
		if info == nil {
			return nil, fmt.Errorf("package not found: %s", pkg)
		}

		bi := backend.PackageInfo{
			Name:      info.Name,
			Installed: info.Installed,
			Provided:  info.Provided,
		}

		if info.InAUR {
			bi.Source = "aur"
			bi.Available = info.InAUR
		} else {
			bi.Source = "repo"
			bi.Available = info.Exists
		}

		result = append(result, bi)
	}

	log.Debug("Pacman.Resolve: done (%.2fs)", time.Since(start).Seconds())
	return result, nil
}

// -----------------------------------------
// Private functions

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