# AGENTS.md Guide for Declpac

## Skills: Always Active

At the start of every conversation, load the following skills using the `skill` tool before responding to the user:

1. **caveman** — Always use caveman mode (full intensity) for all responses

## Code Organization

### Go file structure
- Public (exported) functions must appear at the top of each file.
- Private (unexported) functions must appear at the bottom of each file.
- Within each section, functions must be sorted alphabetically by function name.
- The two sections must be separated by exactly this divider:

	// -----------------------------------------
	// private
