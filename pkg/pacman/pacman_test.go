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
	// Arrange
	p := New()

	// Act
	name := p.Name()

	// Verify
	if name != "pacman" {
		t.Errorf("Name() = %q, want %q", name, "pacman")
	}
}
