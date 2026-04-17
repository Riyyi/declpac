## Context

Current `Resolve` in `pkg/fetch/fetch.go`:
1. Initialize all packages with `Exists: true`
2. Check local DB, set `Installed: true` if found
3. Check sync DBs, set `Installed: false`
4. Check AUR, set `InAUR: true`
5. Error if not found anywhere

Bug: wrong default + wrong order + conflates two concerns.

## Goals / Non-Goals

**Goals:**
- Fix Resolve algorithm to correctly classify packages
- Separate "available" from "installed" concerns

**Non-Goals:**
- No new APIs or features
- No refactoring outside Resolve function

## Decisions

**Decision 1: Start with Exists: false**
- Default `Exists: false` (unknown) instead of `true`
- Only set true when confirmed in sync DBs

**Decision 2: Check sync DBs first**
- Order: sync → local → AUR
- First determine if package exists (anywhere)
- Then determine if installed (local only)

**Decision 3: Two-phase classification**

| Phase | Check | Sets |
|------|-------|-----|
| 1. Availability | sync DBs | Exists: true/false |
| 2. Installation | local DB | Installed: true/false |
| 3. Fallback | AUR | InAUR: true if not in sync |

## Risks / Trade-offs

- Minimal risk: localized fix only
- Trade-off: slight performance cost (check sync before local) - acceptable for correctness