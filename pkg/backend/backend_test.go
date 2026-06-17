package backend

import (
	"errors"
	"io"
	"testing"
)

// -----------------------------------------
// Types

type mockBackend struct{}

func (m *mockBackend) Name() string                                  { return "mock" }
func (m *mockBackend) Resolve([]string) ([]PackageInfo, error)       { return nil, nil }
func (m *mockBackend) List() ([]string, error)                      { return nil, nil }
func (m *mockBackend) ListExplicit() ([]string, error)              { return nil, nil }
func (m *mockBackend) ListOrphans() ([]string, error)               { return nil, nil }
func (m *mockBackend) RefreshDB(io.Writer) error                    { return nil }
func (m *mockBackend) InstallPackages([]PackageInfo, io.Writer) error { return nil }
func (m *mockBackend) RemoveOrphans([]string, io.Writer) (int, error) { return 0, nil }
func (m *mockBackend) MarkAs([]string, string, io.Writer) error      { return nil }
func (m *mockBackend) Prune(io.Writer) (int, error)                 { return 0, nil }

// -----------------------------------------
// Private functions

func notFoundLookPath(string) (string, error) {
	return "", errors.New("not found")
}

func saveAndRestore(t *testing.T) func() {
	t.Helper()
	origLookPath := lookPath
	origFactories := factories
	factories = map[string]func() Backend{}
	return func() {
		lookPath = origLookPath
		factories = origFactories
	}
}

// -----------------------------------------
// Tests

func TestNewNeitherFound(t *testing.T) {
	restore := saveAndRestore(t)
	defer restore()

	lookPath = notFoundLookPath

	_, err := New()
	if err == nil {
		t.Fatal("expected error when no package manager on PATH")
	}
}

func TestNewBothFound(t *testing.T) {
	restore := saveAndRestore(t)
	defer restore()

	lookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}

	_, err := New()
	if err == nil {
		t.Fatal("expected error when both package managers on PATH")
	}
}

func TestNewPacmanOnly(t *testing.T) {
	restore := saveAndRestore(t)
	defer restore()

	lookPath = func(name string) (string, error) {
		if name == "pacman" {
			return "/usr/bin/pacman", nil
		}
		return "", errors.New("not found")
	}
	Register("pacman", func() Backend { return &mockBackend{} })

	b, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil backend when pacman on PATH")
	}
}

func TestNewDnfOnly(t *testing.T) {
	restore := saveAndRestore(t)
	defer restore()

	lookPath = func(name string) (string, error) {
		if name == "dnf" {
			return "/usr/bin/dnf", nil
		}
		return "", errors.New("not found")
	}
	Register("dnf", func() Backend { return &mockBackend{} })

	b, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil backend when dnf on PATH")
	}
}

func TestNewNotRegistered(t *testing.T) {
	restore := saveAndRestore(t)
	defer restore()

	lookPath = func(name string) (string, error) {
		if name == "pacman" {
			return "/usr/bin/pacman", nil
		}
		return "", errors.New("not found")
	}
	// No backend registered — factories map is empty.

	_, err := New()
	if err == nil {
		t.Fatal("expected error when backend not registered")
	}
}