# ADR-007: Context-Aware Operations

## Status
Accepted

## Context
Validation and repair can be long-running and need to support cancellation, timeouts, and future tracing integration.

## Decision
Expose context-aware entry points for validation and repair, and propagate context through ports and adapters.

## Consequences
### Positive
- Supports cancellation and timeouts (e.g., Ctrl+C in CLI)
- Aligns with Go best practices for long-running operations
- Ready for future tracing or telemetry

### Negative
- Additional plumbing in public APIs and internal calls
