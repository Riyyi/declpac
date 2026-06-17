package backend

import (
	"time"

	"github.com/Riyyi/declpac/pkg/log"
	"github.com/Riyyi/declpac/pkg/output"
)

// -----------------------------------------
// Public functions

func DryRun(b Backend, packages []string) (*output.Result, error) {
	start := time.Now()
	log.Debug("DryRun: starting...")

	resolved, err := b.Resolve(packages)
	if err != nil {
		return nil, err
	}
	log.Debug("DryRun: packages resolved (%.2fs)", time.Since(start).Seconds())

	var toInstall []string
	var aurPkgs []string
	for _, info := range resolved {
		if info.Source == "aur" && !info.Installed {
			aurPkgs = append(aurPkgs, info.Name)
		} else if !info.Installed {
			toInstall = append(toInstall, info.Name)
		}
	}
	log.Debug("DryRun: packages categorized (%.2fs)", time.Since(start).Seconds())

	orphans, err := b.ListOrphans()
	if err != nil {
		return nil, err
	}
	log.Debug("DryRun: orphans listed (%.2fs)", time.Since(start).Seconds())

	pkgSet := make(map[string]bool)
	for _, p := range packages {
		pkgSet[p] = true
	}
	var toRemove []string
	for _, o := range orphans {
		if !pkgSet[o] {
			toRemove = append(toRemove, o)
		}
	}

	log.Debug("DryRun: done (%.2fs)", time.Since(start).Seconds())
	return &output.Result{
		Installed: len(toInstall) + len(aurPkgs),
		Removed:   len(toRemove),
		ToInstall: append(toInstall, aurPkgs...),
		ToRemove:  toRemove,
	}, nil
}
