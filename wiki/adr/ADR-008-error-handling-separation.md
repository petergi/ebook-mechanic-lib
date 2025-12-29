# ADR-008: Error Handling Separation

## Status
Accepted

## Context
Users need to distinguish between operational failures (I/O, parse errors) and validation findings that describe document problems.

## Decision
Return operational failures as Go `error` values and report validation findings inside `ValidationReport.Errors`.

## Consequences
### Positive
- Clear separation between system failures and content defects
- Enables consistent exit code mapping in CLI
- Prevents conflating file access issues with validation results

### Negative
- Requires callers to handle both error paths and report status
