# ADR-004: Go Language Choice

## Status
Accepted

## Context
The project needs fast validation of large EPUB/PDF files, simple deployment as a single binary, strong standard library support, and clear concurrency primitives for parallel workloads.

## Decision
Use Go as the implementation language for the core library and CLI.

## Consequences
### Positive
- Fast compilation and runtime performance for large files
- Simple, portable builds (single binary distribution)
- Strong standard library coverage for I/O and parsing
- Built-in concurrency primitives for parallel validation

### Negative
- More verbose error handling than exception-based languages
- Smaller ecosystem for niche ebook tooling
