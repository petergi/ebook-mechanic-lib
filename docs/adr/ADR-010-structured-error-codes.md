# ADR-010: Structured Error Codes

## Status
Accepted

## Context
Validation findings need to be machine-readable, consistently documented, and easy to filter across EPUB and PDF validators.

## Decision
Adopt a structured error code format: `<FORMAT>-<CATEGORY>-<NUMBER>` (e.g., `EPUB-OPF-004`).

## Consequences
### Positive
- Predictable filtering and grouping
- Simple mapping to documentation and specs
- Consistent across formats

### Negative
- Requires maintaining code registries and documentation
