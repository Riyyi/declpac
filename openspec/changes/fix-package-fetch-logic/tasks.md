## 1. Fix Resolve Function Logic

- [x] 1.1 Change initialization: start packages with Exists: false instead of true
- [x] 1.2 Reorder to check sync DBs BEFORE local DB
- [x] 1.3 Separate availability check (sync DBs) from installation check (local DB)
- [x] 1.4 Add AUR fallback for packages not in sync DBs
- [x] 1.5 Validate all packages have either Exists or InAUR before returning
