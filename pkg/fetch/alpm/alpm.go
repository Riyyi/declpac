package alpm

import (
	"fmt"
	"os"
	"time"

	"github.com/Jguer/dyalpm"
)

var (
	Root        = "/"
	PacmanState = "/var/lib/pacman"
)

type Handle struct {
	handle  dyalpm.Handle
	localDB dyalpm.Database
	syncDBs []dyalpm.Database
}

func New() (*Handle, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] alpm.New: starting...\n")

	handle, err := dyalpm.Initialize(Root, PacmanState)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize alpm: %w", err)
	}

	localDB, err := handle.LocalDB()
	if err != nil {
		handle.Release()
		return nil, fmt.Errorf("failed to get local database: %w", err)
	}

	syncDBs, err := handle.SyncDBs()
	if err != nil {
		handle.Release()
		return nil, fmt.Errorf("failed to get sync databases: %w", err)
	}

	if len(syncDBs) == 0 {
		syncDBs, err = registerSyncDBs(handle)
		if err != nil {
			handle.Release()
			return nil, fmt.Errorf("failed to register sync databases: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] alpm.New: done (%.2fs)\n", time.Since(start).Seconds())
	return &Handle{
		handle:  handle,
		localDB: localDB,
		syncDBs: syncDBs,
	}, nil
}

func (h *Handle) Release() error {
	if h.handle != nil {
		h.handle.Release()
	}
	return nil
}

func registerSyncDBs(handle dyalpm.Handle) ([]dyalpm.Database, error) {
	fmt.Fprintf(os.Stderr, "[debug] registerSyncDBs: starting...\n")

	repos := []string{"core", "extra", "multilib"}
	var dbs []dyalpm.Database

	for _, repo := range repos {
		db, err := handle.RegisterSyncDB(repo, 0)
		if err != nil {
			continue
		}

		count := 0
		db.PkgCache().ForEach(func(pkg dyalpm.Package) error {
			count++
			return nil
		})

		if count > 0 {
			dbs = append(dbs, db)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] registerSyncDBs: done (%d dbs)\n", len(dbs))
	return dbs, nil
}

func (h *Handle) LocalPackages() (map[string]dyalpm.Package, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] LocalPackages: starting...\n")

	localPkgs := make(map[string]dyalpm.Package)

	err := h.localDB.PkgCache().ForEach(func(pkg dyalpm.Package) error {
		localPkgs[pkg.Name()] = pkg
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate local package cache: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[debug] LocalPackages: done (%.2fs)\n", time.Since(start).Seconds())
	return localPkgs, nil
}

func (h *Handle) SyncPackages(pkgNames []string) (map[string]dyalpm.Package, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] SyncPackages: starting...\n")

	syncPkgs := make(map[string]dyalpm.Package)
	pkgSet := make(map[string]bool)
	for _, name := range pkgNames {
		pkgSet[name] = true
	}

	for _, db := range h.syncDBs {
		err := db.PkgCache().ForEach(func(pkg dyalpm.Package) error {
			if pkgSet[pkg.Name()] {
				if _, exists := syncPkgs[pkg.Name()]; !exists {
					syncPkgs[pkg.Name()] = pkg
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to iterate sync database %s: %w", db.Name(), err)
		}
	}

	fmt.Fprintf(os.Stderr, "[debug] SyncPackages: done (%.2fs)\n", time.Since(start).Seconds())
	return syncPkgs, nil
}
