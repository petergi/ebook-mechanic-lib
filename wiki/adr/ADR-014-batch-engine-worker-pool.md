# ADR-014: Batch Engine Worker Pool

## Status
Accepted

## Context
Batch validation and repair must scale to large directories while supporting cancellation, backpressure, and predictable resource usage.

## Decision
Implement a worker pool using buffered channels and context-aware cancellation with structured result aggregation.

## Consequences
### Positive
- Scales across cores with configurable worker counts
- Supports graceful shutdown and partial results
- Provides backpressure for large file sets

### Negative
- Additional concurrency complexity and coordination logic
