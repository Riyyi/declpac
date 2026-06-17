# 06 - Unit Tests for Backend Orchestration and Auto-Detection

Status: ready-for-agent

## Description

Add unit tests for the new `pkg/backend/` package using a mock Backend implementation.

### Mock Backend

Create a mock that implements `backend.Backend` with configurable responses and call tracking.

### DryRun orchestration tests

- Given packages that resolve to install/remove/orphan, returns correct `output.Result`
- Given empty packages, handles correctly
- Given backend that returns errors, propagates errors

### Sync orchestration tests

- Safety check: config.NoCheck=false and state packages < half of explicit → error
- Safety check: config.NoCheck=true → skips check
- Prune sanity check: missing state packages → error
- Normal sync flow: calls Resolve, InstallPackages, Prune in order

### Auto-detection tests

- PATH with only pacman → returns pacman backend
- PATH with only dnf → returns dnf backend
- PATH with both → error
- PATH with neither → error

## Acceptance Criteria

- `pkg/backend/backend_test.go` exists and passes
- Mock Backend implementation exists
- All test scenarios above covered
- `go test ./pkg/backend/` passes