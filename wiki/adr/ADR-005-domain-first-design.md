# ADR-005: Domain-First Design

## Status
Accepted

## Context
Validation and repair rules are the most stable part of the system. The team needs to avoid coupling these rules to file I/O, CLI concerns, or reporter formats.

## Decision
Model domain entities and rules first, then define ports, then implement adapters.

## Consequences
### Positive
- Stable core entities (`ValidationReport`, `ValidationError`, `Severity`)
- Clear dependency direction (adapters depend on ports/domain)
- Easier unit testing without external dependencies

### Negative
- Requires more upfront modeling before implementing adapters
