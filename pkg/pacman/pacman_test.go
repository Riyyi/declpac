package pacman

import (
	"testing"

	"github.com/Riyyi/declpac/pkg/backend"
)

// -----------------------------------------
// Tests

func TestPacmanSatisfiesBackend(t *testing.T) {
	var _ backend.Backend = (*Pacman)(nil)
}

func TestPacmanName(t *testing.T) {
	p := New()
	if p.Name() != "pacman" {
		t.Errorf("Name() = %q, want %q", p.Name(), "pacman")
	}
}