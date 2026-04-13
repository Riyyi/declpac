# declpac

`declpac` is a declarative package manager for Arch Linux that syncs your system
with a declared package list using `pacman`. It ensures your system matches your
desired state, handling package installation, upgrades, and orphan cleanup
automatically.

## Features

- **Declarative state management** — Define your desired package list in files or stdin
- **Automatic dependency resolution** — Pacman handles transitive dependencies
- **Smart orphan cleanup** — Removes packages no longer needed
- **Explicit package tracking** — Marks your declared packages as explicit
- **AUR support** — Falls back to AUR for packages not in official repos
- **Machine-readable output** — Perfect for scripting and automation

## Installation

### Build from Source

```bash
git clone https://github.com/Riyyi/declpac.git
cd declpac
go build -o declpac ./cmd/declpac
sudo mv declpac /usr/local/bin/
```

### Dependencies

- Go 1.21+
- pacman (system package manager)
- aur (AUR helper, optional for AUR support)
- Root privileges (required for pacman operations)

## Usage

### Basic Usage

```bash
# Single state file
sudo declpac --state packages.txt

# Multiple state files
sudo declpac --state base.txt --state apps.txt

# From stdin
cat packages.txt | sudo declpac
```

### State File Format

State files contain one package name per line:

```
bash
vim
git
docker
```

Lines are treated as package names with whitespace trimmed:

```
bash          # bash
  vim        # vim
# comment   # ignored
```

### Command Line Options

| Flag | Alias | Description |
|------|-------|-------------|
| `--state` | `-s` | State file(s) to read package list from (can be used multiple times) |
| `--yes` | `-y` | Skip confirmation prompts (for scripting) |
| `--dry-run` | | Simulate sync without making changes |
| `--help` | `-h` | Show help message |

### Examples

#### Minimal System

```bash
# Create a minimal system package list
echo -e "base\nbase-devel\nlinux-headers\nvim\ngit\ncurl\nwget" > ~/.config/declpac/minimal.txt

# Apply the state
sudo declpac --state ~/.config/declpac/minimal.txt
```

#### Development Environment

```bash
# development.txt
go
nodejs
python
rust
docker
docker-compose
kubectl
helm
terraform

# Apply
sudo declpac --state development.txt
```

#### Full System Sync

```bash
# Combine multiple files
sudo declpac --state ~/.config/declpac/base.txt --state ~/.config/declpac/desktop.txt

# Or use stdin
cat ~/.config/declpac/full-system.txt | sudo declpac
```

#### Dry-Run Preview

```bash
# Preview what would happen without making changes
sudo declpac --dry-run --state packages.txt

# Example output:
# Installed 3 packages, removed 2 packages
# Would install: vim, git, docker
# Would remove: python2, perl-xml-parser
```

## How It Works

1. **Collect packages** — Reads from all `--state` files and stdin
2. **Merge** — Combines all packages additively (duplicates allowed)
3. **Validate** — Checks packages exist in repos or AUR
4. **Mark explicit** — Marks declared packages as explicit dependencies
5. **Sync** — Runs `pacman -Syu` to install/upgrade packages
6. **Cleanup** — Removes orphaned packages with `pacman -Rns`
7. **Report** — Outputs summary: `Installed X packages, removed Y packages`

### Database Freshness

If the pacman database is older than 24 hours, `declpac` automatically refreshes it with `pacman -Syy` before validation.

### Orphan Cleanup

After syncing, `declpac` identifies and removes packages that are:
- Not explicitly installed
- Not required by any other package

This keeps your system clean from dependency artifacts.

## Output Format

```
# Success (packages installed/removed)
Installed 5 packages, removed 2 packages

# Success (no changes)
Installed 0 packages, removed 0 packages

# Error
error: package not found: <package-name>
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (no packages, validation failure, pacman error) |

## Security Considerations

- **Run as root** — `declpac` requires root privileges for pacman operations
- **Review state files** — Only install packages from trusted sources
- **Backup** — Consider backing up your system before major changes

## Troubleshooting

### "Permission denied"

`declpac` requires root privileges. Use `sudo`:

```bash
sudo declpac --state packages.txt
```

### "Package not found"

The package doesn't exist in pacman repos or AUR. Check the package name:

```bash
pacman -Ss <package>
```

### Database sync fails

Refresh manually:

```bash
sudo pacman -Syy
```

## File Structure

```
declpac/
├── cmd/declpac/
│   └── main.go         # Entry point
├── pkg/
│   ├── input/          # State file/stdin reading
│   ├── merge/          # Package merging
│   ├── validation/     # Package validation
│   ├── pacman/         # Pacman integration
│   └── output/         # Output formatting
├── go.mod              # Go module
└── README.md           # This file
```

## License

GPL-3.0
