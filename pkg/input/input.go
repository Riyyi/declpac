package input

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Riyyi/declpac/pkg/lib"
)

var ErrEmptyList = errors.New("package list is empty")

// -----------------------------------------
// public

func Merge(packages map[string]bool) ([]string, error) {
	result := make([]string, 0, len(packages))
	for name := range packages {
		result = append(result, name)
	}
	if len(result) == 0 {
		return nil, ErrEmptyList
	}
	return result, nil
}

func ReadPackages(stateFiles []string) (map[string]bool, error) {
	packages := make(map[string]bool)

	for _, file := range stateFiles {
		expanded := lib.ExpandPath(file)
		if err := readStateFile(expanded, packages); err != nil {
			return nil, err
		}
	}

	implicitStateFile := getImplicitStateFile()
	if fileExists(implicitStateFile) {
		if err := readStateFile(implicitStateFile, packages); err != nil {
			return nil, err
		}
	}

	if err := readStdin(packages); err != nil {
		return nil, err
	}

	return packages, nil
}

// -----------------------------------------
// private

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getImplicitStateFile() string {
	cfgDir := os.Getenv("XDG_CONFIG_HOME")
	if cfgDir == "" {
		cfgDir = "~/.config"
	}
	cfgDir = lib.ExpandPath(cfgDir)
	return filepath.Join(cfgDir, "declpac")
}

func normalizePackageName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || strings.HasPrefix(name, "#") {
		return ""
	}
	return name
}

func readStateFile(path string, packages map[string]bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := normalizePackageName(scanner.Text())
		if name != "" {
			packages[name] = true
		}
	}

	return scanner.Err()
}

func readStdin(packages map[string]bool) error {
	info, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	if (info.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := normalizePackageName(scanner.Text())
		if name != "" {
			packages[name] = true
		}
	}

	return scanner.Err()
}
