package validation

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

var LockFile = "/var/lib/pacman/db.lock"

func CheckDBFreshness() error {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] CheckDBFreshness: starting...\n")

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

	fmt.Fprintf(os.Stderr, "[debug] CheckDBFreshness: done (%.2fs)\n", time.Since(start).Seconds())
	return nil
}
