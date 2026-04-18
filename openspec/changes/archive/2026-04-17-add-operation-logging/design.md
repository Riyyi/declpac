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
| Captured | All state-modifying functions | capture output, write to log with single timestamp at start, write to terminal |

### One Timestamp Per Tool Call
Instead of streaming with MultiWriter (multiple timestamps), each state-modifying function:
1. Writes timestamp + operation name to log
2. Runs command, captures output
3. Writes captured output to log
4. Writes output to terminal

This ensures exactly 1 timestamp print per tool call.

### Error Handling
- Write error to log BEFORE returning from function
- Print error to stderr so user sees it

### Dependencies
- Add to imports: `os`, `path/filepath`

### Flow in Sync() or main entrypoint
1. Call `OpenLog()` at program start, defer close
2. Each state-modifying function calls `state.Write()` with timestamp prefix

### Wire into main.go
- Open log at start of `run()`
- Pass log writer to pacman package (via exported function or global)
