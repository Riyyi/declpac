# 04 - Add Dnf Backend Stub

Status: ready-for-agent

## Description

Create `pkg/dnf/dnf.go` with a `Dnf` struct that implements the `Backend` interface. All methods return errors indicating they are not yet implemented, except:
- `Name()` returns `"dnf"`

This stub allows the project to compile and the auto-detection to work, while making it clear the dnf backend is not functional.

## Acceptance Criteria

- `*dnf.Dnf` satisfies `backend.Backend` interface
- `backend.New()` returns `*dnf.Dnf` when only dnf is on PATH
- All methods except `Name()` return an error