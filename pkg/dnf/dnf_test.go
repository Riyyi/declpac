package dnf

import (
	"testing"

	"github.com/Riyyi/declpac/pkg/backend"
)

// -----------------------------------------
// Tests

func TestDnfSatisfiesBackend(t *testing.T) {
	var _ backend.Backend = (*Dnf)(nil)
}

func TestDnfName(t *testing.T) {
	d := &Dnf{}
	if d.Name() != "dnf" {
		t.Errorf("Name() = %q, want %q", d.Name(), "dnf")
	}
}

func TestDnfUnimplementedMethods(t *testing.T) {
	d := &Dnf{}
	pkgs := []backend.PackageInfo{{Name: "test"}}

	if _, err := d.Resolve([]string{"test"}); err == nil {
		t.Error("Resolve should return error")
	}

	if _, err := d.List(); err == nil {
		t.Error("List should return error")
	}

	if _, err := d.ListExplicit(); err == nil {
		t.Error("ListExplicit should return error")
	}

	if _, err := d.ListOrphans(); err == nil {
		t.Error("ListOrphans should return error")
	}

	if err := d.RefreshDB(nil); err == nil {
		t.Error("RefreshDB should return error")
	}

	if err := d.InstallPackages(pkgs, nil); err == nil {
		t.Error("InstallPackages should return error")
	}

	if _, err := d.RemoveOrphans([]string{"test"}, nil); err == nil {
		t.Error("RemoveOrphans should return error")
	}

	if err := d.MarkAs([]string{"test"}, "explicit", nil); err == nil {
		t.Error("MarkAs should return error")
	}

	if _, err := d.Prune([]string{"test"}, nil); err == nil {
		t.Error("Prune should return error")
	}
}