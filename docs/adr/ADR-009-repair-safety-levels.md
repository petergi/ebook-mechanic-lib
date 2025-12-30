# ADR-009: Repair Safety Levels

## Status
Accepted

## Context
Automatic repairs can alter content or structure. Users need predictable behavior and explicit control over higher-risk fixes.

## Decision
Classify repair actions by safety level (Very High, High, Medium, Low) and require explicit approval for higher-risk actions.

## Consequences
### Positive
- Clear user expectations about repair risk
- Enables policy-based automation for safe repairs
- Reduces accidental destructive changes

### Negative
- Additional configuration and UI surface for approvals
