package fetch

import (
	"fmt"
	"time"

	"github.com/Riyyi/declpac/pkg/fetch/alpm"
	"github.com/Riyyi/declpac/pkg/fetch/aur"
	"github.com/Riyyi/declpac/pkg/log"
)

type PackageInfo struct {
	Name      string
	InAUR     bool
	Exists    bool
	Installed bool
	Provided  string
	AURInfo   *aur.Package
}

type Fetcher struct {
	alpmHandle *alpm.Handle
	aurClient  *aur.Client
}

// -----------------------------------------
// constructor

func New() (*Fetcher, error) {
	start := time.Now()
	log.Debug("fetch.Fetcher New: starting...")

	alpmHandle, err := alpm.New()
	if err != nil {
		return nil, err
	}

	aurClient := aur.New()

	log.Debug("fetch.Fetcher New: done (%.2fs)", time.Since(start).Seconds())
	return &Fetcher{
		alpmHandle: alpmHandle,
		aurClient:  aurClient,
	}, nil
}

// -----------------------------------------
// public

func (f *Fetcher) BuildLocalPkgMap() (map[string]interface{}, error) {
	localPkgs, err := f.alpmHandle.LocalPackages()
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for k, v := range localPkgs {
		result[k] = v
	}
	return result, nil
}

func (f *Fetcher) Close() error {
	return f.alpmHandle.Release()
}

func (f *Fetcher) FetchAur(packages []string) (map[string]aur.Package, error) {
	return f.aurClient.Fetch(packages)
}

func (f *Fetcher) FindProvidingPackage(depName string) (string, bool) {
	return f.alpmHandle.FindProvidingPackage(depName)
}

func (f *Fetcher) GetAURPackage(name string) (aur.Package, bool) {
	return f.aurClient.Get(name)
}

func (f *Fetcher) Resolve(packages []string) (map[string]*PackageInfo, error) {
	start := time.Now()
	log.Debug("fetch.Resolve: starting...")

	result := make(map[string]*PackageInfo)
	for _, pkg := range packages {
		result[pkg] = &PackageInfo{Name: pkg, Exists: false}
	}

	syncPkgs, err := f.alpmHandle.SyncPackages(packages)
	if err != nil {
		return nil, err
	}
	log.Debug("fetch.Resolve: sync db check done (%.2fs)", time.Since(start).Seconds())

	for pkg := range syncPkgs {
		result[pkg].Exists = true
		result[pkg].InAUR = false
	}

	localPkgs, err := f.alpmHandle.LocalPackages()
	if err != nil {
		return nil, err
	}
	log.Debug("fetch.Resolve: local pkgs built (%.2fs)", time.Since(start).Seconds())

	for pkg := range localPkgs {
		if info, ok := result[pkg]; ok {
			info.Installed = true
		}
	}

	var notInSync []string
	for _, pkg := range packages {
		if !result[pkg].Exists {
			notInSync = append(notInSync, pkg)
		}
	}

	if len(notInSync) > 0 {
		if _, err := f.aurClient.Fetch(notInSync); err != nil {
			log.Debug("fetch.Resolve: aur fetch error: %v", err)
		}

		for _, pkg := range packages {
			info := result[pkg]
			if info.Exists {
				continue
			}

			if aurInfo, ok := f.aurClient.Get(pkg); ok {
				info.InAUR = true
				info.AURInfo = &aurInfo
				continue
			}

			if providedBy, ok := f.FindProvidingPackage(pkg); ok {
				log.Debug("fetch.Resolve: %s provided by %s", pkg, providedBy)
				info.Provided = providedBy
				info.Exists = true
				continue
			}

			return nil, fmt.Errorf("package not found: %s", pkg)
		}
	}

	for _, pkg := range packages {
		info := result[pkg]
		if !info.Exists && !info.InAUR {
			return nil, fmt.Errorf("package not validated: %s", pkg)
		}
	}

	log.Debug("fetch.Resolve: done (%.2fs)", time.Since(start).Seconds())
	return result, nil
}
