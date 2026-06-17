package backend

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// -----------------------------------------
// Types

type callTracker struct {
	calls []string
}

func (ct *callTracker) record(name string) {
	ct.calls = append(ct.calls, name)
}

func (ct *callTracker) has(name string) bool {
	for _, c := range ct.calls {
		if c == name {
			return true
		}
	}
	return false
}

func (ct *callTracker) indexOf(name string) int {
	for i, c := range ct.calls {
		if c == name {
			return i
		}
	}
	return -1
}

type configurableBackend struct {
	tracker        *callTracker
	resolveFn      func([]string) ([]PackageInfo, error)
	listFn          func() ([]string, error)
	listExplicitFn  func() ([]string, error)
	listOrphansFn   func() ([]string, error)
	refreshDBFn     func(io.Writer) error
	installPkgsFn   func([]PackageInfo, io.Writer) error
	removeOrphansFn func([]string, io.Writer) (int, error)
	markAsFn        func([]string, string, io.Writer) error
	pruneFn         func([]string, io.Writer) (int, error)
}

func (b *configurableBackend) Name() string                                  { return "configurable" }
func (b *configurableBackend) Resolve(pkgs []string) ([]PackageInfo, error) { b.tracker.record("Resolve"); return b.resolveFn(pkgs) }
func (b *configurableBackend) List() ([]string, error)                       { b.tracker.record("List"); return b.listFn() }
func (b *configurableBackend) ListExplicit() ([]string, error)              { b.tracker.record("ListExplicit"); return b.listExplicitFn() }
func (b *configurableBackend) ListOrphans() ([]string, error)               { b.tracker.record("ListOrphans"); return b.listOrphansFn() }
func (b *configurableBackend) RefreshDB(w io.Writer) error                   { b.tracker.record("RefreshDB"); return b.refreshDBFn(w) }
func (b *configurableBackend) InstallPackages(pkgs []PackageInfo, w io.Writer) error {
	b.tracker.record("InstallPackages")
	return b.installPkgsFn(pkgs, w)
}
func (b *configurableBackend) RemoveOrphans(orphs []string, w io.Writer) (int, error) {
	b.tracker.record("RemoveOrphans")
	return b.removeOrphansFn(orphs, w)
}
func (b *configurableBackend) MarkAs(pkgs []string, flag string, w io.Writer) error {
	b.tracker.record("MarkAs")
	return b.markAsFn(pkgs, flag, w)
}
func (b *configurableBackend) Prune(statePkgs []string, w io.Writer) (int, error) {
	b.tracker.record("Prune")
	return b.pruneFn(statePkgs, w)
}

func noOpFn() func(io.Writer) error {
	return func(io.Writer) error { return nil }
}

func noOpIntFn() func([]string, io.Writer) (int, error) {
	return func([]string, io.Writer) (int, error) { return 0, nil }
}

func noOpMarkAsFn() func([]string, string, io.Writer) error {
	return func([]string, string, io.Writer) error { return nil }
}

func noOpPruneFn() func([]string, io.Writer) (int, error) {
	return func([]string, io.Writer) (int, error) { return 0, nil }
}

func noOpInstallFn() func([]PackageInfo, io.Writer) error {
	return func([]PackageInfo, io.Writer) error { return nil }
}

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
// Tests — Auto-detection

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

	_, err := New()
	if err == nil {
		t.Fatal("expected error when backend not registered")
	}
}

// -----------------------------------------
// Tests — DryRun orchestration

func TestDryRunInstallAndRemove(t *testing.T) {
	tracker := &callTracker{}
	b := &configurableBackend{
		tracker: tracker,
		resolveFn: func(pkgs []string) ([]PackageInfo, error) {
			return []PackageInfo{
				{Name: "vim", Source: "repo", Available: true, Installed: false},
				{Name: "neovim", Source: "aur", Available: true, Installed: false},
				{Name: "git", Source: "repo", Available: true, Installed: true},
			}, nil
		},
		listOrphansFn: func() ([]string, error) {
			return []string{"orphan1", "orphan2"}, nil
		},
	}

	result, err := DryRun(b, []string{"vim", "neovim", "git"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Installed != 2 {
		t.Errorf("Installed = %d, want 2", result.Installed)
	}
	if result.Removed != 2 {
		t.Errorf("Removed = %d, want 2", result.Removed)
	}
	if len(result.ToInstall) != 2 {
		t.Errorf("ToInstall = %v, want 2 items", result.ToInstall)
	}
	if len(result.ToRemove) != 2 {
		t.Errorf("ToRemove = %v, want 2 items", result.ToRemove)
	}

	if !tracker.has("Resolve") {
		t.Error("expected Resolve to be called")
	}
	if !tracker.has("ListOrphans") {
		t.Error("expected ListOrphans to be called")
	}
}

func TestDryRunEmptyPackages(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		resolveFn: func(pkgs []string) ([]PackageInfo, error) {
			return nil, nil
		},
		listOrphansFn: func() ([]string, error) {
			return nil, nil
		},
	}

	result, err := DryRun(b, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Installed != 0 {
		t.Errorf("Installed = %d, want 0", result.Installed)
	}
	if result.Removed != 0 {
		t.Errorf("Removed = %d, want 0", result.Removed)
	}
}

func TestDryRunResolveError(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		resolveFn: func(pkgs []string) ([]PackageInfo, error) {
			return nil, errors.New("resolve failed")
		},
	}

	_, err := DryRun(b, []string{"vim"})
	if err == nil {
		t.Fatal("expected error from Resolve failure")
	}
}

func TestDryRunListOrphansError(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		resolveFn: func(pkgs []string) ([]PackageInfo, error) {
			return []PackageInfo{{Name: "vim", Installed: true}}, nil
		},
		listOrphansFn: func() ([]string, error) {
			return nil, errors.New("orphan list failed")
		},
	}

	_, err := DryRun(b, []string{"vim"})
	if err == nil {
		t.Fatal("expected error from ListOrphans failure")
	}
}

func TestDryRunOrphansNotInStatePackages(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		resolveFn: func(pkgs []string) ([]PackageInfo, error) {
			return []PackageInfo{
				{Name: "vim", Source: "repo", Available: true, Installed: true},
			}, nil
		},
		listOrphansFn: func() ([]string, error) {
			return []string{"vim", "oldpkg"}, nil
		},
	}

	result, err := DryRun(b, []string{"vim"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ToRemove) != 1 {
		t.Errorf("ToRemove = %v, want 1 item (oldpkg)", result.ToRemove)
	}
	if result.ToRemove[0] != "oldpkg" {
		t.Errorf("ToRemove[0] = %q, want %q", result.ToRemove[0], "oldpack")
	}
}

// -----------------------------------------
// Tests — Sync orchestration

func TestSyncSafetyCheckFails(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		listExplicitFn: func() ([]string, error) {
			return []string{"a", "b", "c", "d", "e", "f"}, nil
		},
	}

	_, err := Sync(b, []string{"vim"}, false, false)
	if err == nil {
		t.Fatal("expected safety check error")
	}
	if !strings.Contains(err.Error(), "safety check") {
		t.Errorf("error = %v, want safety check error", err)
	}
}

func TestSyncSafetyCheckSkippedWithNoCheck(t *testing.T) {
	tracker := &callTracker{}
	b := &configurableBackend{
		tracker: tracker,
		listExplicitFn: func() ([]string, error) {
			return []string{"a", "b", "c", "d", "e", "f"}, nil
		},
		listFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		refreshDBFn:     noOpFn(),
		resolveFn:       func(pkgs []string) ([]PackageInfo, error) { return nil, nil },
		installPkgsFn:   noOpInstallFn(),
	}

	_, err := Sync(b, []string{"vim"}, true, false)
	if err != nil {
		t.Fatalf("unexpected error with noCheck: %v", err)
	}
}

func TestSyncPruneSanityCheckFails(t *testing.T) {
	tracker := &callTracker{}
	b := &configurableBackend{
		tracker: tracker,
		listExplicitFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		listFn: func() ([]string, error) {
			if tracker.has("List") && len(tracker.calls) > 1 {
				return []string{"otherpkg"}, nil
			}
			return []string{"otherpkg"}, nil
		},
		refreshDBFn:     noOpFn(),
		resolveFn:       func(pkgs []string) ([]PackageInfo, error) { return nil, nil },
		installPkgsFn:   noOpInstallFn(),
		pruneFn:         noOpPruneFn(),
	}

	_, err := Sync(b, []string{"missingpkg"}, true, true)
	if err == nil {
		t.Fatal("expected prune sanity check error")
	}
	if !strings.Contains(err.Error(), "missing state packages") {
		t.Errorf("error = %v, want missing state packages error", err)
	}
}

func TestSyncCallOrder(t *testing.T) {
	tracker := &callTracker{}
	b := &configurableBackend{
		tracker: tracker,
		listExplicitFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		listFn: func() ([]string, error) {
			return []string{"vim", "git"}, nil
		},
		refreshDBFn:     noOpFn(),
		resolveFn:       func(pkgs []string) ([]PackageInfo, error) { return nil, nil },
		installPkgsFn:   noOpInstallFn(),
	}

	_, err := Sync(b, []string{"vim"}, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOrder := []string{"ListExplicit", "List", "RefreshDB", "Resolve", "InstallPackages"}
	if len(tracker.calls) < len(expectedOrder) {
		t.Fatalf("expected at least %d calls, got %d", len(expectedOrder), len(tracker.calls))
	}
	for i, expected := range expectedOrder {
		if tracker.calls[i] != expected {
			t.Errorf("call %d = %q, want %q", i, tracker.calls[i], expected)
		}
	}
}

func TestSyncWithPrune(t *testing.T) {
	tracker := &callTracker{}
	b := &configurableBackend{
		tracker: tracker,
		listExplicitFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		listFn: func() ([]string, error) {
			return []string{"vim", "git"}, nil
		},
		refreshDBFn:   noOpFn(),
		resolveFn:     func(pkgs []string) ([]PackageInfo, error) { return nil, nil },
		installPkgsFn: noOpInstallFn(),
		pruneFn:       noOpPruneFn(),
	}

	result, err := Sync(b, []string{"vim"}, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !tracker.has("Prune") {
		t.Error("expected Prune to be called")
	}
	if result.Removed != 0 {
		t.Errorf("Removed = %d, want 0", result.Removed)
	}
}

func TestSyncResolveError(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		listExplicitFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		listFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		refreshDBFn: noOpFn(),
		resolveFn: func(pkgs []string) ([]PackageInfo, error) {
			return nil, errors.New("resolve failed")
		},
	}

	_, err := Sync(b, []string{"vim"}, true, false)
	if err == nil {
		t.Fatal("expected error from Resolve failure")
	}
}

func TestSyncInstallError(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		listExplicitFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		listFn: func() ([]string, error) {
			return []string{"vim"}, nil
		},
		refreshDBFn: noOpFn(),
		resolveFn:   func(pkgs []string) ([]PackageInfo, error) { return nil, nil },
		installPkgsFn: func([]PackageInfo, io.Writer) error {
			return errors.New("install failed")
		},
	}

	_, err := Sync(b, []string{"vim"}, true, false)
	if err == nil {
		t.Fatal("expected error from InstallPackages failure")
	}
}

// -----------------------------------------
// Tests — Prune sanity check

func TestPruneSanityCheckMissingPackages(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		listFn: func() ([]string, error) {
			return []string{"git", "curl"}, nil
		},
	}

	err := pruneSanityCheck(b, []string{"vim", "git"})
	if err == nil {
		t.Fatal("expected error for missing packages")
	}
	if !strings.Contains(err.Error(), "vim") {
		t.Errorf("error should mention vim: %v", err)
	}
}

func TestPruneSanityCheckAllPresent(t *testing.T) {
	b := &configurableBackend{
		tracker: &callTracker{},
		listFn: func() ([]string, error) {
			return []string{"vim", "git", "curl"}, nil
		},
	}

	err := pruneSanityCheck(b, []string{"vim", "git"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Keep mockBackend for auto-detection tests

type mockBackend struct{}

func (m *mockBackend) Name() string                                        { return "mock" }
func (m *mockBackend) Resolve([]string) ([]PackageInfo, error)             { return nil, nil }
func (m *mockBackend) List() ([]string, error)                             { return nil, nil }
func (m *mockBackend) ListExplicit() ([]string, error)                     { return nil, nil }
func (m *mockBackend) ListOrphans() ([]string, error)                      { return nil, nil }
func (m *mockBackend) RefreshDB(io.Writer) error                           { return nil }
func (m *mockBackend) InstallPackages([]PackageInfo, io.Writer) error      { return nil }
func (m *mockBackend) RemoveOrphans([]string, io.Writer) (int, error)      { return 0, nil }
func (m *mockBackend) MarkAs([]string, string, io.Writer) error            { return nil }
func (m *mockBackend) Prune([]string, io.Writer) (int, error)             { return 0, nil }