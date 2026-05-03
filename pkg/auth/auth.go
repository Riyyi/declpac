package auth

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/Riyyi/declpac/pkg/log"
)

var tool string
var timeout time.Duration = 5 * time.Minute
var refreshCommand []string = []string{"-n", "true"}

// -----------------------------------------
// public

func Command(name string, args ...string) *exec.Cmd {
	if tool == "" {
		return log.Command(name, args...)
	}
	args = append([]string{name}, args...)
	return log.Command(tool, args...)
}

func Run() {
	exec.Command(tool, refreshCommand...).Run()
}

func Start() error {
	err := detect()
	if err != nil {
		return err
	}

	// Automatically refresh privilege elevation to prevent user prompts
	go func() {
		for {
			Run()
			time.Sleep(timeout)
		}
	}()

	return nil
}

// -----------------------------------------
// private

func detect() error {
	tool = getTool()
	if tool == "" {
		return fmt.Errorf("no privilege elevation tool detected in PATH")
	}

	parseTimeout()
	// We have to be a little faster than the actual timeout
	timeout -= 30 * time.Second

	return nil
}

func execLookPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func getTool() string {
	sudo := execLookPath("sudo")
	doas := execLookPath("doas")

	if sudo != "" {
		return "sudo"
	}
	if doas != "" {
		return "doas"
	}

	return ""
}

func parseTimeout() {
	switch tool {
	case "sudo":
		out, err := exec.Command("sudo", "sudo", "-V").CombinedOutput()
		if err != nil {
			return
		}
		re := regexp.MustCompile(`Authentication timestamp timeout: (\d+)\..*`)
		matches := re.FindStringSubmatch(string(out))
		if len(matches) == 2 {
			if minutes, err := strconv.Atoi(matches[1]); err == nil {
				timeout = time.Duration(minutes) * time.Minute
			}
		}
	case "doas":
		exec.Command("doas", "true").Run()
	}
}
