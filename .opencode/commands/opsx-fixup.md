---
description: Fixup diverging archived openspec change artifacts
---

Fixup archived openspec change artifacts in @openspec/changes/archive/ that have diverged from the codebase.

For each archived change directory:
1. Read the .md artifact
2. Review related code to verify consistency
3. Fix any mismatches (status, fields, missing links, etc.)

Only run make-commit if changes were made, with a specific message like "Fixup: add missing X field to Y change".

---

**Steps**

1. **For each archived openspec change**

   Read the change artifact.

   Go through the related code.

   Correct the change artifact where needed.

2. **Run make-commit if changes were made**

   skill [name=make-commit] message: "Fixup: <specific fix>"
