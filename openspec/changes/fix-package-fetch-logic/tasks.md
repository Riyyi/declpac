## 1. Fix Resolve Function Logic

- [ ] 1.1 Change initialization: start packages with Exists: false instead of true
- [ ] 1.2 Reorder to check sync DBs BEFORE local DB
- [ ] 1.3 Separate availability check (sync DBs) from installation check (local DB)
- [ ] 1.4 Add AUR fallback for packages not in sync DBs
- [ ] 1.5 Validate all packages have either Exists or InAUR before returning
