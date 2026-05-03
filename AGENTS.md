# AGENTS.md Guide for Declpac

## Skills: Always Active

At the start of every conversation, load the following skills using the `skill`
tool before responding to the user:

1. **caveman** — Always use caveman mode (full intensity) for all responses

## Code Organization

### Go: File Structure
- Public (exported) functions must appear at the top of each file.
- Private (unexported) functions must appear at the bottom of each file.
- Within each section, functions must be sorted alphabetically by function name.
- The two sections must be separated by exactly this divider:

	// -----------------------------------------
	// private

### Go: Line Length

Keep Go lines as short as reasonably possible.

- **Hard limit:** 120 characters — never exceed this.
- **Preferred:** 100 characters or fewer in normal cases.
- **Target:** 80 characters if breaking the line produces cleaner, more readable
  code (e.g. chained calls, long argument lists, multi-condition `if` statements).
