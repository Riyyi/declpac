package read

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Riyyi/declpac/pkg/fetch"
	"github.com/Riyyi/declpac/pkg/log"
	"github.com/Riyyi/declpac/pkg/output"
)

var LockFile = "/var/lib/pacman/db.lock"

func List() ([]string, error) {
	start := time.Now()
	log.Debug("List: starting...")

	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	list := strings.Split(strings.TrimSpace(string(output)), "\n")
	if list[0] == "" {
		list = nil
	}

	log.Debug("List: done (%.2fs)", time.Since(start).Seconds())
	return list, nil
}

func ListOrphans() ([]string, error) {
	start := time.Now()
	log.Debug("ListOrphans: starting...")

	cmd := exec.Command("pacman", "-Qdtq")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	orphans := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(orphans) > 0 && orphans[0] == "" {
		orphans = orphans[1:]
	}

	log.Debug("ListOrphans: done (%.2fs)", time.Since(start).Seconds())
	return orphans, nil
}

func DBFreshness() (bool, error) {
	info, err := os.Stat(LockFile)
	if err != nil {
		return false, nil
	}

	age := time.Since(info.ModTime())
	return age <= 24*time.Hour, nil
}

func DryRun(packages []string) (*output.Result, error) {
	start := time.Now()
	log.Debug("DryRun: starting...")

	f, err := fetch.New()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Debug("DryRun: initialized fetcher (%.2fs)", time.Since(start).Seconds())

	resolved, err := f.Resolve(packages)
	if err != nil {
		return nil, err
	}
	log.Debug("DryRun: packages resolved (%.2fs)", time.Since(start).Seconds())

	var toInstall []string
	var aurPkgs []string
	for _, pkg := range packages {
		info := resolved[pkg]
		if info == nil || (!info.Exists && !info.InAUR) {
			return nil, fmt.Errorf("package not found: %s", pkg)
		}
		if info.InAUR && !info.Installed {
			aurPkgs = append(aurPkgs, pkg)
		} else if !info.Installed {
			toInstall = append(toInstall, pkg)
		}
	}
	log.Debug("DryRun: packages categorized (%.2fs)", time.Since(start).Seconds())

	orphans, err := ListOrphans()
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
