package validation

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var LockFile = "/var/lib/pacman/db.lock"

func Validate(packages []string) error {
	if len(packages) == 0 {
		return errors.New("no packages found")
	}

	if err := checkDBFreshness(); err != nil {
		return err
	}

	if err := validatePackages(packages); err != nil {
		return err
	}

	return nil
}

func checkDBFreshness() error {
	info, err := os.Stat(LockFile)
	if err != nil {
		return nil
	}

	age := time.Since(info.ModTime())
	if age > 24*time.Hour {
		cmd := exec.Command("pacman", "-Syy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to refresh pacman database: %w", err)
		}
	}

	return nil
}

func validatePackages(packages []string) error {
	for _, pkg := range packages {
		if err := validatePackage(pkg); err != nil {
			return err
		}
	}
	return nil
}

func validatePackage(name string) error {
	cmd := exec.Command("pacman", "-Qip", name)
	if err := cmd.Run(); err == nil {
		return nil
	}

	cmd = exec.Command("pacman", "-Sip", name)
	if err := cmd.Run(); err == nil {
		return nil
	}

	cmd = exec.Command("aur", "search", name)
	if out, err := cmd.Output(); err == nil && len(out) > 0 {
		return nil
	}

	return fmt.Errorf("package not found: %s", name)
}
