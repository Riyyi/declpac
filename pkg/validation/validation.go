package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

var LockFile = "/var/lib/pacman/db.lock"

const AURInfoURL = "https://aur.archlinux.org/rpc?v=5&type=info"

type AURResponse struct {
	Results []AURResult `json:"results"`
}

type AURResult struct {
	Name string `json:"Name"`
}

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
	var pacmanPkgs []string
	var aurPkgs []string

	for _, pkg := range packages {
		if inPacman(pkg) {
			pacmanPkgs = append(pacmanPkgs, pkg)
		} else {
			aurPkgs = append(aurPkgs, pkg)
		}
	}

	if len(aurPkgs) > 0 {
		foundAUR := batchSearchAUR(aurPkgs)
		for _, pkg := range aurPkgs {
			if !foundAUR[pkg] {
				return fmt.Errorf("package not found: %s", pkg)
			}
		}
	}

	return nil
}

func inPacman(name string) bool {
	cmd := exec.Command("pacman", "-Qip", name)
	if err := cmd.Run(); err == nil {
		return true
	}

	cmd = exec.Command("pacman", "-Sip", name)
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

func batchSearchAUR(packages []string) map[string]bool {
	result := make(map[string]bool)

	if len(packages) == 0 {
		return result
	}

	v := url.Values{}
	v.Set("type", "info")
	for _, pkg := range packages {
		v.Add("arg[]", pkg)
	}

	resp, err := http.Get(AURInfoURL + "&" + v.Encode())
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result
	}

	var aurResp AURResponse
	if err := json.Unmarshal(body, &aurResp); err != nil {
		return result
	}

	for _, r := range aurResp.Results {
		result[r.Name] = true
	}

	return result
}
