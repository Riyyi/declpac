# declpac

Declarative package manager for Arch Linux that syncs your system with a declared package list using pacman.

## Features

- Declarative state management — define your desired package list in files or stdin
- Smart orphan cleanup — removes packages no longer needed
- Explicit package tracking — marks declared packages as explicit
- AUR support — builds and installs AUR packages automatically
- Machine-readable output — perfect for scripting

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
- pacman
- makepkg (for AUR support)
- git (for AUR support)
- Root privileges

## Usage

```bash
# Single state file
sudo declpac --state packages.txt

# Multiple state files
sudo declpac --state base.txt --state apps.txt

# From stdin
cat packages.txt | sudo declpac

# Preview changes without applying
sudo declpac --dry-run --state packages.txt
```

### State File Format

One package name per line, lines beginning with `#` are comments:

```
bash
vim
git
docker
# this is a comment
```

### Options

| Flag | Alias | Description |
|------|-------|-------------|
| `--state` | `-s` | State file(s) to read package list from (can be used multiple times) |
| `--nocheck` | | Skip safety check (allow significant package count reductions)
| `--dry-run` | | Preview changes without applying them |
| `--verbose` | `-v` | Enable verbose output |
| `--help` | `-h` | Show help message |

## How It Works

1. **Read** — Collect packages from all state files and stdin
2. **Merge** — Combine into single package list
3. **Categorize** — Check if packages are in official repos or AUR
4. **Sync** — Install/update packages via pacman
5. **Build** — Build and install AUR packages via makepkg
6. **Mark** — Mark declared packages as explicit, all others as dependencies
7. **Cleanup** — Remove orphaned packages

### Database Freshness

If the pacman database is older than 24 hours, it is automatically refreshed.

### Logging

Operations are logged to `/var/log/declpac.log`.

## Troubleshooting

### Permission denied

Use sudo:
```bash
sudo declpac --state packages.txt
```

### Package not found

Check if the package exists:
```bash
pacman -Ss <package>
```

### Database sync fails

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
│   ├── fetch/          # Package resolution
│   │   ├── aur/        # AUR support
│   │   └── alpm/       # ALPM support
│   ├── pacman/         # Pacman operations
│   │   ├── read/       # Read packages
│   │   └── sync/       # Sync packages
│   ├── log/            # Logging
│   └── output/         # Output formatting
└── README.md
```

## License

GPL-3.0
