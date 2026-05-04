package sync

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/Riyyi/declpac/pkg/auth"
	"github.com/Riyyi/declpac/pkg/fetch"
	"github.com/Riyyi/declpac/pkg/fetch/aur"
	"github.com/Riyyi/declpac/pkg/log"
)

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

	tmpDir := getTempDirName() + "/" + pkgName
	if err := createTempDir(tmpDir); err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := cloneRepo(packageBase, tmpDir, logWriter); err != nil {
		return err
	}
	log.Debug("InstallAUR: cloned (%.2fs)", time.Since(start).Seconds())

	if err := buildPackage(tmpDir, asDeps, logWriter); err != nil {
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
	cmd := auth.Command("pacman", args...)
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

	cmd := auth.Command("pacman", "-Syy")
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
	removeCmd := auth.Command(args[0], args[1:]...)
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
	cmd := auth.Command("pacman", args...)
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

func buildPackage(tmpDir string, asDeps bool, logWriter io.Writer) error {
	makepkgArgs := []string{"-D", tmpDir, "-s", "--noconfirm"}
	if asDeps {
		makepkgArgs = append(makepkgArgs, "--asdeps")
	}
	makepkgCmd := log.Command("makepkg", makepkgArgs...)
	makepkgCmd.Stdout = logWriter
	makepkgCmd.Stderr = logWriter
	if err := makepkgCmd.Run(); err != nil {
		return fmt.Errorf("makepkg failed to build AUR package: %w", err)
	}
	return nil
}

func cloneRepo(packageBase string, tmpDir string, logWriter io.Writer) error {
	cloneURL := "https://aur.archlinux.org/" + packageBase + ".git"
	cloneCmd := log.Command("git", "clone", cloneURL, tmpDir)
	cloneCmd.Stdout = logWriter
	cloneCmd.Stderr = logWriter
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone AUR repo: %w", err)
	}
	return nil
}

func createTempDir(tmpDir string) error {
	if tmpDir == "" || tmpDir == "/" || !strings.HasPrefix(tmpDir, "/tmp") {
		return fmt.Errorf("safety check: prevented malformed rm -rf call")
	}

	rmdirCmd := log.Command("rm", "-rf", tmpDir)
	if err := rmdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove temp directory: %w", err)
	}

	mkdirCmd := log.Command("mkdir", "-p", tmpDir)
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
		// Skip packages that do not start with the exact package name, ex: sunshine-bin -> sunshine
		if !strings.HasPrefix(name, pkgName) {
			continue
		}
		// Skip packages that provide a debug package, ex: sunshine-bin -> sunshine-debug
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

func getTempDirName() string {
	user, err := user.Current()
	if err != nil {
		return "/tmp/declpac"
	}

	return "/tmp/declpac-" + user.Username
}

func installBuiltPackage(pkgFile string, logWriter io.Writer) error {
	installCmd := auth.Command("pacman", "-U", "--noconfirm", pkgFile)
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

	var repoDeps, aurDeps []string
	for _, dep := range depends {
		info := resolved[dep]
		if info.Installed {
			continue
		}
		pkg := dep
		if info.Provided != "" {
			pkg = info.Provided
		}
		if info.Exists {
			repoDeps = append(repoDeps, pkg)
		} else if info.InAUR {
			aurDeps = append(aurDeps, pkg)
		}
	}

	if len(repoDeps) > 0 {
		if err := SyncPackages(repoDeps, logWriter); err != nil {
			return fmt.Errorf("failed to install repo dependencies: %w", err)
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
