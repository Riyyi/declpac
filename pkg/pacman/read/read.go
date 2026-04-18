package read

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Riyyi/declpac/pkg/fetch"
	"github.com/Riyyi/declpac/pkg/output"
)

var LockFile = "/var/lib/pacman/db.lock"

func List() ([]string, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] List: starting...\n")

	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	list := strings.Split(strings.TrimSpace(string(output)), "\n")
	if list[0] == "" {
		list = nil
	}

	fmt.Fprintf(os.Stderr, "[debug] List: done (%.2fs)\n", time.Since(start).Seconds())
	return list, nil
}

func ListOrphans() ([]string, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] ListOrphans: starting...\n")

	cmd := exec.Command("pacman", "-Qdtq")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	orphans := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(orphans) > 0 && orphans[0] == "" {
		orphans = orphans[1:]
	}

	fmt.Fprintf(os.Stderr, "[debug] ListOrphans: done (%.2fs)\n", time.Since(start).Seconds())
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

	var toInstall []string
	var aurPkgs []string
	for _, pkg := range packages {
		info := resolved[pkg]
		if info == nil || (!info.Exists && !info.InAUR) {
			return nil, fmt.Errorf("package not found: %s", pkg)
		}
		if info.InAUR {
			aurPkgs = append(aurPkgs, pkg)
		} else if !info.Installed {
			toInstall = append(toInstall, pkg)
		}
	}
	fmt.Fprintf(os.Stderr, "[debug] DryRun: packages categorized (%.2fs)\n", time.Since(start).Seconds())

	orphans, err := ListOrphans()
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
