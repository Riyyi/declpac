package sync

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Riyyi/declpac/pkg/fetch"
	"github.com/Riyyi/declpac/pkg/fetch/aur"
	"github.com/Riyyi/declpac/pkg/log"
)

var sudoUser string
var sudoUserOnce sync.Once

type Result struct {
	Installed int
	Removed   int
}

// -----------------------------------------
// public

func InstallAUR(f *fetch.Fetcher, pkgName string, packageBase string, asDeps bool, logWriter io.Writer) error {
	start := time.Now()
	log.Debug("InstallAUR: starting...")

	if logWriter == nil {
		logWriter = os.Stderr
	}

	aurInfo := getAURInfo(f, pkgName, packageBase)
	if err := resolveAndInstallDeps(f, aurInfo, logWriter); err != nil {
		return err
	}

	sudoUser := getSudoUser()
	tmpDir := "/tmp/declpac/" + pkgName
	if err := createTempDir(sudoUser, tmpDir); err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := cloneRepo(sudoUser, packageBase, tmpDir, logWriter); err != nil {
		return err
	}
	log.Debug("InstallAUR: cloned (%.2fs)", time.Since(start).Seconds())

	if err := buildPackage(sudoUser, tmpDir, asDeps, logWriter); err != nil {
		return err
	}
	log.Debug("InstallAUR: built (%.2fs)", time.Since(start).Seconds())

	pkgFile, err := findPKGFile(pkgName, tmpDir)
	if err != nil {
		return fmt.Errorf("failed to find built package: %w", err)
	}

	if err := installBuiltPackage(pkgFile, logWriter); err != nil {
		return err
	}
	log.Debug("InstallAUR: done (%.2fs)", time.Since(start).Seconds())

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
	cmd := log.Command("pacman", args...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("mark as %s failed: %w", flag, err)
	}

	log.Debug("MarkAs(%s): done (%.2fs)", flag, time.Since(start).Seconds())
	return nil
}

func RefreshDB(logWriter io.Writer) error {
	start := time.Now()
	log.Debug("RefreshDB: starting...")

	if logWriter == nil {
		logWriter = os.Stderr
	}

	cmd := log.Command("pacman", "-Syy")
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to refresh pacman database: %w", err)
	}

	log.Debug("RefreshDB: done (%.2fs)", time.Since(start).Seconds())
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
	removeCmd := log.Command(args[0], args[1:]...)
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

func SyncPackages(packages []string, logWriter io.Writer) error {
	start := time.Now()
	log.Debug("SyncPackages: starting...")

	if logWriter == nil {
		logWriter = os.Stderr
	}

	args := append([]string{"-S", "--needed", "--noconfirm"}, packages...)
	cmd := log.Command("pacman", args...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pacman sync failed: %w", err)
	}

	log.Debug("SyncPackages: done (%.2fs)", time.Since(start).Seconds())
	return nil
}

// -----------------------------------------
// private

func buildPackage(sudoUser string, tmpDir string, asDeps bool, logWriter io.Writer) error {
	makepkgArgs := []string{"makepkg", "-s", "--noconfirm"}
	if asDeps {
		makepkgArgs = append(makepkgArgs, "--asdeps")
	}
	makepkgCmd := log.Command("su", "-", sudoUser, "-c", "cd "+tmpDir+" && "+strings.Join(makepkgArgs, " "))
	makepkgCmd.Stdout = logWriter
	makepkgCmd.Stderr = logWriter
	if err := makepkgCmd.Run(); err != nil {
		return fmt.Errorf("makepkg failed to build AUR package: %w", err)
	}
	return nil
}

func cloneRepo(sudoUser string, packageBase string, tmpDir string, logWriter io.Writer) error {
	cloneURL := "https://aur.archlinux.org/" + packageBase + ".git"
	cloneCmd := log.Command("su", "-", sudoUser, "-c", "git clone "+cloneURL+" "+tmpDir)
	cloneCmd.Stdout = logWriter
	cloneCmd.Stderr = logWriter
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone AUR repo: %w", err)
	}
	return nil
}

func createTempDir(sudoUser string, tmpDir string) error {
	mkdirCmd := log.Command("su", "-", sudoUser, "-c", "rm -rf "+tmpDir+" && mkdir -p "+tmpDir)
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	return nil
}

func findPKGFile(pkgName string, dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".pkg.tar.zst") && !strings.HasSuffix(name, ".pkg.tar.gz") {
			continue
		}
		if strings.HasPrefix(name, pkgName+"-debug") {
			continue
		}
		return strings.Join([]string{dir, name}, "/"), nil
	}
	return "", fmt.Errorf("no package file found in %s", dir)
}

func getAURInfo(f *fetch.Fetcher, pkgName string, packageBase string) *aur.Package {
	if packageBase == "" {
		return nil
	}
	info, ok := f.GetAURPackage(pkgName)
	if !ok {
		return nil
	}
	return &info
}

func getSudoUser() string {
	sudoUserOnce.Do(func() {
		sudoUser = os.Getenv("SUDO_USER")
		if sudoUser == "" {
			sudoUser = os.Getenv("USER")
			if sudoUser == "" {
				sudoUser = "root"
			}
		}
	})
	return sudoUser
}

func installBuiltPackage(pkgFile string, logWriter io.Writer) error {
	installCmd := log.Command("pacman", "-U", "--noconfirm", pkgFile)
	installCmd.Stdout = logWriter
	installCmd.Stderr = logWriter
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}
	return nil
}

func resolveAndInstallDeps(f *fetch.Fetcher, aurInfo *aur.Package, logWriter io.Writer) error {
	if aurInfo == nil {
		return nil
	}

	depends := aurInfo.AllDepends()
	if len(depends) == 0 {
		return nil
	}

	resolved, err := f.Resolve(depends)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	var aurDeps []string
	for _, dep := range depends {
		info := resolved[dep]
		if info.Installed {
			continue
		}
		if info.Exists {
			continue
		}
		if info.InAUR {
			aurDeps = append(aurDeps, dep)
		}
	}

	if len(aurDeps) == 0 {
		return nil
	}

	fetched, err := f.FetchAur(aurDeps)
	if err != nil {
		log.Debug("sync.resolveAndInstallDeps: aur fetch error: %v", err)
	}
	for _, dep := range aurDeps {
		depInfo, ok := fetched[dep]
		if !ok {
			continue
		}
		if err := InstallAUR(f, dep, depInfo.PackageBase, true, logWriter); err != nil {
			return err
		}
	}

	return nil
}
