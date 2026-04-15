---
name: make-commit
description: >
  Make a git commit, asking if it was by the user or AI. User commits use git
  config for both committer and author. AI commits use git config for committer
  but set author to AI Bot <ai@local>.
license: GPL-3.0
metadata:
  author: riyyi
  version: "1.0"
---

Make a git commit, distinguishing between user and AI contributions.

---

**Steps**

1. **[REQUIRED] Ask user if this commit is by them or by AI**

   Use the **question tool** to ask:
   > "Was this commit made by you or by AI?"

   Options:
   - "By me" - User made the commit
   - "By AI" - AI made the commit

2. **Check for commit message**

   **Capitalization rule**: Commit messages should start with a capital letter,
   unless it refers to a tool or project that explicitly uses lowercase as its
   name (e.g., "go", "npm", "rustc").

   If the user did NOT provide a commit message, generate one from staged changes:
   ```bash
   git diff --staged --stat
   ```
   Create a reasonable commit message based on the changes.

   If thed user DID provide a message, format it into a proper commit message.

3. **Show commit message and confirm**

   Use the **question tool** to ask:
   > "Is this commit message okay, or would you like to make tweaks?"
   > <message>

   Options:
   - "Looks good" - Proceed with this message
   - "Make tweaks" - User will provide a new message

   **If user wants tweaks**: Ask them for the new commit message.

4. **Make the commit**

   Use the commit message provided by the user.

   **If by user:**
   ```bash
   git commit -m "<message>"
   ```
   (Uses git config user as both committer and author)

   **If by AI:**
   ```bash
   git commit -m "<message>" --author="AI Bot <ai@local>"
   ```
   (Uses git config for committer, but sets author to AI Bot)

**Output**

- Tell user the commit was made
- If AI commit, mention that the author was set to "AI Bot <ai@local>"
