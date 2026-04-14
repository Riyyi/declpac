## Context

The dyalpm library's `SyncDBs()` returns an empty slice because it doesn't automatically load sync databases from pacman.conf. This causes `resolvePackages()` to skip the sync DB search entirely and fall through to checking the AUR, which fails for official repo packages.

## Goals / Non-Goals

**Goals:**
- Register sync databases so `resolvePackages()` can find official repo packages

**Non-Goals:**
- Modify package installation logic
- Add support for custom repositories

## Decisions

1. **Register each repo manually** - After `handle.SyncDBs()` returns empty, loop through known repos (core, extra, multilib) and call `handle.RegisterSyncDB()` for each.

2. **Use hardcoded repo list** - Arch Linux standard repos are core, extra, multilib. This matches pacman.conf.

3. **Silent failure for missing repos** - If a repo isn't configured in pacman.conf, `RegisterSyncDB` will return a valid but empty DB. Filter by checking if `PkgCache()` has any packages.

## Risks / Trade-offs

- Hardcoded repo names may need updating if Arch adds/removes standard repos → Low risk, rare change
- Repo registration could fail silently → Mitigated by checking PkgCache count