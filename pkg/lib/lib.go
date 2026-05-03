package lib

import (
	"os"
	"path/filepath"
	"strings"
)

// -----------------------------------------
// public

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
