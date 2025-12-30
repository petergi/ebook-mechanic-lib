# ADR-013: Cobra CLI

## Status
Accepted

## Context
The project requires a production-ready CLI with subcommands, consistent help text, structured flags, and shell-friendly behavior across platforms.

## Decision
Use Cobra for the CLI framework, with command implementations under `cmd/` and shared logic in `internal/cli/`.

## Consequences
### Positive
- Standardized command/flag handling and help generation
- Consistent UX across validate/repair/batch commands
- Easier future extension of CLI surface

### Negative
- Adds an external dependency and learning curve for contributors
