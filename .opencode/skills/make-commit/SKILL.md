---
name: make-commit
description: >
  Make a git commit, asking if it was by the user or AI. User commits use git
  config for both committer and author. AI commits use git config for committer
  but set author to AI Bot <ai@local>.
license: GPL-3.0
metadata:
  author: riyyi
  version: "1.1"
---

Make a git commit, distinguishing between user and AI contributions.

---

**Commit message format**

A valid commit message consists of a required **title** and an optional **body**:

Rules:
- **Title** (required): max 72 characters, starts with a capital letter —
  unless referring to a tool/project that explicitly uses lowercase (e.g.,
  "go", "npm", "rustc"). No trailing period.
- **Body** (optional): any further elaboration. Each line max 72 characters.
  Wrap manually — do not rely on the terminal to wrap.

---

**Steps**

1. **[REQUIRED] Ask user if this commit is by them or by AI**

   Use the **question tool** to ask:
   > "Was this commit made by you or by AI?"

   Options:
   - "By me" - User made the commit
   - "By AI" - AI made the commit

2. **Compose the commit message**

   If the user did NOT provide a commit message, generate one from staged
   changes:
   ```bash
   git --no-pager diff --staged
   ```
   Write a commit message following the format above.

   If the user **DID** provide a message, treat it as raw input and apply the
   format rules to it.

3. **Validate the commit message**

   Before presenting the message to the user, check it against every rule:

   - [ ] Title is present and non-empty
   - [ ] Title is at most 72 characters
   - [ ] Title starts with a capital letter (or an intentionally lowercase name)
   - [ ] Title has no trailing period
   - [ ] Every line in the body is at most 72 characters

   Fix any violations silently before showing the message to the user.

4. **Show commit message and confirm**

   Use the **question tool** to ask:
   > "Is this commit message okay, or would you like to make tweaks?"
   > ```
   > <message>
   > ```

   Options:
   - "Looks good" - Proceed with this message
   - "Make tweaks" - User will provide a new message or describe changes

   **If user wants tweaks**: apply the same validation (step 3) to the revised
   message before committing.

5. **Make the commit**

   For a title-only message:
   ```bash
   git commit -m "<title>"
   ```

   For a message with a body, pass `-m` twice (git inserts the blank line):
   ```bash
   git commit -m "<title>" -m "<body>"
   ```

   Append `--author="AI Bot <ai@local>"` when the commit is by AI:
   ```bash
   git commit -m "<title>" [-m "<body>"] --author="AI Bot <ai@local>"
   ```

**Output**

- Tell user the commit was made.
- If AI commit, mention that the author was set to "AI Bot <ai@local>".
