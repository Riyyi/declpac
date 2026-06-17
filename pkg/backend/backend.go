package backend

import (
	"fmt"
	"io"
	"os/exec"
)

// -----------------------------------------
// Variables

var lookPath = exec.LookPath

// -----------------------------------------
// Types

type Backend interface {
	Name() string
	Resolve(packages []string) ([]PackageInfo, error)
	List() ([]string, error)
	ListExplicit() ([]string, error)
	ListOrphans() ([]string, error)
	RefreshDB(logWriter io.Writer) error
	InstallPackages(pkgs []PackageInfo, logWriter io.Writer) error
	RemoveOrphans(orphans []string, logWriter io.Writer) (int, error)
	MarkAs(packages []string, flag string, logWriter io.Writer) error
	Prune(statePackages []string, logWriter io.Writer) (int, error)
}

type PackageInfo struct {
	Name      string
	Source    string
	Available bool
	Installed bool
	Provided  string
}

// -----------------------------------------
// Constructors

var factories = map[string]func() Backend{}

func Register(name string, factory func() Backend) {
	factories[name] = factory
}

// -----------------------------------------
// Public functions

func New() (Backend, error) {
	hasPacman := lookPathFound("pacman")
	hasDnf := lookPathFound("dnf")

	switch {
	case hasPacman && hasDnf:
		return nil, fmt.Errorf(
			"both pacman and dnf found on PATH; " +
				"only one supported package manager expected",
		)
	case hasPacman:
		return newBackend("pacman")
	case hasDnf:
		return newBackend("dnf")
	default:
		return nil, fmt.Errorf(
			"no supported package manager found on PATH " +
				"(expected pacman or dnf)",
		)
	}
}

// -----------------------------------------
// Private functions

func lookPathFound(name string) bool {
	_, err := lookPath(name)
	return err == nil
}

func newBackend(name string) (Backend, error) {
	fn, ok := factories[name]
	if !ok {
		return nil, fmt.Errorf("%s backend not registered", name)
	}
	return fn(), nil
}