package input

import (
	"bufio"
	"os"
	"strings"
)

func ReadPackages(stateFiles []string) (map[string]bool, error) {
	packages := make(map[string]bool)

	for _, file := range stateFiles {
		if err := readStateFile(file, packages); err != nil {
			return nil, err
		}
	}

	if err := readStdin(packages); err != nil {
		return nil, err
	}

	return packages, nil
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

func normalizePackageName(name string) string {
	return strings.TrimSpace(name)
}
