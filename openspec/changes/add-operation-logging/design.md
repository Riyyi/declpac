## Implementation

### Log File Location
- Path: `/var/log/declpac.log`
- Single merged log file (stdout + stderr intermingled in order of arrival)

### State-Modifying Functions (need logging)
1. `SyncPackages()` - `pacman -S --needed <packages>`
2. `InstallAUR()` - `git clone` + `makepkg -si --noconfirm`
3. `MarkAllAsDeps()` - `pacman -D --asdeps`
4. `MarkAsExplicit()` - `pacman -D --asexplicit <packages>`
5. `CleanupOrphans()` - `pacman -Rns`

### Functions to Skip (read-only)
- `DryRun()` - queries only
- `getInstalledCount()` - pacman -Qq

### Execution Patterns

| Pattern | Functions | How |
|---------|------------|-----|
| Streaming | MarkAllAsDeps, MarkAsExplicit, InstallAUR | `io.MultiWriter` to tee to both terminal and log |
| Captured | SyncPackages, CleanupOrphans | capture with `CombinedOutput()`, write to log, write to terminal |

### Error Handling
- Write error to log BEFORE returning from function
- Print error to stderr so user sees it

### Dependencies
- Add to imports: `io`, `os`, `path/filepath`

### Structure

```go
// pkg/state/state.go
var logFile *os.File

func OpenLog() error {
    logPath := filepath.Join("/var/log", "declpac.log")
    f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    logFile = f
    return nil
}

func GetLogWriter() io.Writer {
    return logFile
}

func Close() error {
    if logFile == nil {
        return nil
    }
    return logFile.Close()
}
```

### Flow in Sync() or main entrypoint
1. Call `OpenLog()` at program start, defer close
2. Each state-modifying function uses `state.GetLogWriter()` via MultiWriter

### Wire into main.go
- Open log at start of `run()`
- Pass log writer to pacman package (via exported function or global)
