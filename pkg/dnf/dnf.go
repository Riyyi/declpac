package dnf

import (
	"fmt"
	"io"

	"github.com/Riyyi/declpac/pkg/backend"
)

// -----------------------------------------
// Variables

func init() {
	backend.Register("dnf", func() backend.Backend {
		return New()
	})
}

// -----------------------------------------
// Types

type Dnf struct{}

// -----------------------------------------
// Constructors

func New() *Dnf {
	return &Dnf{}
}

// -----------------------------------------
// Public methods

func (d *Dnf) InstallPackages(_ []backend.PackageInfo, _ io.Writer) error {
	return fmt.Errorf("dnf: InstallPackages not implemented")
}

func (d *Dnf) List() ([]string, error) {
	return nil, fmt.Errorf("dnf: List not implemented")
}

func (d *Dnf) ListExplicit() ([]string, error) {
	return nil, fmt.Errorf("dnf: ListExplicit not implemented")
}

func (d *Dnf) ListOrphans() ([]string, error) {
	return nil, fmt.Errorf("dnf: ListOrphans not implemented")
}

func (d *Dnf) MarkAs(_ []string, _ string, _ io.Writer) error {
	return fmt.Errorf("dnf: MarkAs not implemented")
}

func (d *Dnf) Name() string {
	return "dnf"
}

func (d *Dnf) Prune(_ []string, _ io.Writer) (int, error) {
	return 0, fmt.Errorf("dnf: Prune not implemented")
}

func (d *Dnf) RefreshDB(_ io.Writer) error {
	return fmt.Errorf("dnf: RefreshDB not implemented")
}

func (d *Dnf) RemoveOrphans(_ []string, _ io.Writer) (int, error) {
	return 0, fmt.Errorf("dnf: RemoveOrphans not implemented")
}

func (d *Dnf) Resolve(_ []string) ([]backend.PackageInfo, error) {
	return nil, fmt.Errorf("dnf: Resolve not implemented")
}
