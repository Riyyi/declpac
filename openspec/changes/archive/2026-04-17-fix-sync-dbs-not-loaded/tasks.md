## 1. Implement sync DB registration

- [x] 1.1 Modify Pac struct to include a method for registering sync DBs
- [x] 1.2 In New(), after getting empty syncDBs from handle, loop through ["core", "extra", "multilib"] and call RegisterSyncDB for each
- [x] 1.3 Filter out repos that have no packages (not configured in pacman.conf)

## 2. Test the fix

- [x] 2.1 Run `./declpac --dry-run --state <(echo "cmake")` and verify it resolves cmake from extra repo
- [x] 2.2 Test with other official repo packages (e.g., "git" from extra, "base" from core)