package read

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Riyyi/declpac/pkg/log"
)

var LockFile = "/var/lib/pacman/db.lock"

func DBFreshness() (bool, error) {
	info, err := os.Stat(LockFile)
	if err != nil {
		return false, nil
	}

	age := time.Since(info.ModTime())
	return age <= 24*time.Hour, nil
}

func ExplicitList() ([]string, error) {
	start := time.Now()
	log.Debug("ExplicitList: starting...")

	cmd := exec.Command("pacman", "-Qqe")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	list := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(list) > 0 && list[0] == "" {
		list = nil
	}

	log.Debug("ExplicitList: done (%.2fs)", time.Since(start).Seconds())
	return list, nil
}

func List() ([]string, error) {
	start := time.Now()
	log.Debug("List: starting...")

	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	list := strings.Split(strings.TrimSpace(string(output)), "\n")
	if list[0] == "" {
		list = nil
	}

	log.Debug("List: done (%.2fs)", time.Since(start).Seconds())
	return list, nil
}

func ListOrphans() ([]string, error) {
	start := time.Now()
	log.Debug("ListOrphans: starting...")

	cmd := exec.Command("pacman", "-Qdtq")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && stderr.Len() == 0 {
			return nil, nil
		}
		return nil, err
	}

	orphans := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(orphans) > 0 && orphans[0] == "" {
		orphans = orphans[1:]
	}

	log.Debug("ListOrphans: done (%.2fs)", time.Since(start).Seconds())
	return orphans, nil
}
