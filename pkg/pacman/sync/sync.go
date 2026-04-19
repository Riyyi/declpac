package sync

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Riyyi/declpac/pkg/log"
)

type Result struct {
	Installed int
	Removed   int
}

func SyncPackages(packages []string, logWriter io.Writer) error {
	start := time.Now()
	log.Debug("SyncPackages: starting...")

	if logWriter == nil {
		logWriter = os.Stderr
	}

	args := append([]string{"-S", "--needed", "--noconfirm"}, packages...)
	cmdStr := "pacman " + strings.Join(args, " ")
	fmt.Fprintf(logWriter, "[cmd] %s\n", cmdStr)
	cmd := exec.Command("pacman", args...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pacman sync failed: %w", err)
	}

	log.Debug("SyncPackages: done (%.2fs)", time.Since(start).Seconds())
	return nil
}

func RefreshDB(logWriter io.Writer) error {
	start := time.Now()
	log.Debug("RefreshDB: starting...")

	if logWriter == nil {
		logWriter = os.Stderr
	}

	fmt.Fprintf(logWriter, "[cmd] pacman -Syy\n")
	cmd := exec.Command("pacman", "-Syy")
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to refresh pacman database: %w", err)
	}

	log.Debug("RefreshDB: done (%.2fs)", time.Since(start).Seconds())
	return nil
}

func MarkAs(packages []string, flag string, logWriter io.Writer) error {
	if len(packages) == 0 {
		return nil
	}
	start := time.Now()
	flagName := map[string]string{"deps": "asdeps", "explicit": "asexplicit"}[flag]
	log.Debug("MarkAs(%s): starting...", flag)

	if logWriter == nil {
		logWriter = os.Stderr
	}

	args := append([]string{"-D", "--" + flagName}, packages...)
	cmdStr := "pacman " + strings.Join(args, " ")
	fmt.Fprintf(logWriter, "[cmd] %s\n", cmdStr)
	cmd := exec.Command("pacman", args...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("mark as %s failed: %w", flag, err)
	}

	log.Debug("MarkAs(%s): done (%.2fs)", flag, time.Since(start).Seconds())
	return nil
}

func RemoveOrphans(orphans []string, logWriter io.Writer) (int, error) {
	start := time.Now()
	log.Debug("RemoveOrphans: starting...")

	if logWriter == nil {
		logWriter = os.Stderr
	}

	if len(orphans) == 0 {
		log.Debug("RemoveOrphans: done (no orphans) (%.2fs)", time.Since(start).Seconds())
		return 0, nil
	}

	args := make([]string, 0, 3+len(orphans))
	args = append(args, "pacman", "-Rns", "--noconfirm")
	args = append(args, orphans...)
	cmdStr := strings.Join(args, " ")
	fmt.Fprintf(logWriter, "[cmd] %s\n", cmdStr)
	removeCmd := exec.Command(args[0], args[1:]...)
	removeCmd.Stdout = logWriter
	removeCmd.Stderr = logWriter
	err := removeCmd.Run()
	if err != nil {
		return 0, fmt.Errorf("remove orphans failed: %w", err)
	}

	count := len(orphans)

	log.Debug("RemoveOrphans: done (%d) (%.2fs)", count, time.Since(start).Seconds())
	return count, nil
}

func InstallAUR(pkgName string, packageBase string, logWriter io.Writer) error {
	start := time.Now()
	log.Debug("InstallAUR: starting...")

	if logWriter == nil {
		logWriter = os.Stderr
	}

	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		sudoUser = os.Getenv("USER")
		if sudoUser == "" {
			sudoUser = "root"
		}
	}

	tmpDir := "/tmp/declpac-aur-" + pkgName
	mkdirCmdStr := "su - " + sudoUser + " -c 'rm -rf " + tmpDir + " && mkdir -p " + tmpDir + "'"
	fmt.Fprintf(logWriter, "[cmd] %s\n", mkdirCmdStr)
	mkdirCmd := exec.Command("su", "-", sudoUser, "-c", "rm -rf "+tmpDir+" && mkdir -p "+tmpDir)
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneURL := "https://aur.archlinux.org/" + packageBase + ".git"
	cloneCmdStr := "su - " + sudoUser + " -c 'git clone " + cloneURL + " " + tmpDir + "'"
	fmt.Fprintf(logWriter, "[cmd] %s\n", cloneCmdStr)
	cloneCmd := exec.Command("su", "-", sudoUser, "-c", "git clone "+cloneURL+" "+tmpDir)
	cloneCmd.Stdout = logWriter
	cloneCmd.Stderr = logWriter
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone AUR repo: %w", err)
	}
	log.Debug("InstallAUR: cloned (%.2fs)", time.Since(start).Seconds())

	makepkgCmdStr := "su - " + sudoUser + " -c 'cd " + tmpDir + " && makepkg -s --noconfirm'"
	fmt.Fprintf(logWriter, "[cmd] %s\n", makepkgCmdStr)
	makepkgCmd := exec.Command("su", "-", sudoUser, "-c", "cd "+tmpDir+" && makepkg -s --noconfirm")
	makepkgCmd.Stdout = logWriter
	makepkgCmd.Stderr = logWriter
	if err := makepkgCmd.Run(); err != nil {
		return fmt.Errorf("makepkg failed to build AUR package: %w", err)
	}
	log.Debug("InstallAUR: built (%.2fs)", time.Since(start).Seconds())

	pkgFile, err := findPKGFile(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to find built package: %w", err)
	}

	installCmdStr := "pacman -U --noconfirm " + pkgFile
	fmt.Fprintf(logWriter, "[cmd] %s\n", installCmdStr)
	installCmd := exec.Command("pacman", "-U", "--noconfirm", pkgFile)
	installCmd.Stdout = logWriter
	installCmd.Stderr = logWriter
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}
	log.Debug("InstallAUR: done (%.2fs)", time.Since(start).Seconds())

	return nil
}

func findPKGFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".pkg.tar.zst") || strings.HasSuffix(name, ".pkg.tar.gz") {
			return strings.Join([]string{dir, name}, "/"), nil
		}
	}
	return "", fmt.Errorf("no package file found in %s", dir)
}
