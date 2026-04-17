# Dry-Run Mode

## Summary

Add `--dry-run` flag to simulate the sync operation without making any changes
to the system. Shows what packages would be installed and what would be removed.

## Motivation

Users want to preview the effects of a sync operation before committing changes.
This is useful for:
- Verifying the intended changes are correct
- Avoiding unintended package installations
- Understanding what orphan cleanup will remove

## Interface

```
declpac --dry-run --state packages.txt
```

## Behavior

1. Read state files and stdin (same as normal mode)
2. Validate packages exist (same as normal mode)
3. Query current installed packages via `pacman -Qq`
4. Compare declared packages to current state
5. Identify packages that would be installed (not currently installed)
6. Identify orphans that would be removed via `pacman -Qdtq`
7. Output results with "Would install:" and "Would remove:" sections

## Output Format

```
Installed 3 packages, removed 2 packages
Would install: vim, git, docker
Would remove: python2, perl-xml-parser
```

## Non-Goals

- Actual package operations (no pacman -Syu, no pacman -Rns execution)
- Package version comparison
- Detailed dependency analysis

## Trade-offs

- Doesn't predict transitive dependencies that pacman might install
- Orphan list may change after packages are installed
