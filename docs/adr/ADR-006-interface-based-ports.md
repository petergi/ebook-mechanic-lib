# ADR-006: Interface-Based Ports

## Status
Accepted

## Context
The library needs pluggable validators, reporters, and repair services for both EPUB and PDF formats while preserving testability and isolation from implementations.

## Decision
Define ports as interfaces in `internal/ports/` and implement them in `internal/adapters/`.

## Consequences
### Positive
- Easy to swap implementations and add new adapters
- Enables mocks and test doubles for unit tests
- Clear boundary between domain and infrastructure

### Negative
- Additional interface and wiring boilerplate
