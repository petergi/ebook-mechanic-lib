# ADR-012: Test Data Organization

## Status
Accepted

## Context
Test fixtures need to cover each validation error category while remaining easy to locate and extend.

## Decision
Organize fixtures under `testdata/` by format and validity/error category.

## Consequences
### Positive
- Predictable fixture discovery
- Easier coverage tracking per error category
- Encourages deterministic fixture generation

### Negative
- More directories to maintain as categories grow
