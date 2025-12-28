# Agent Development Guide

## Commands

**Setup:**
```bash
make install
```

**Build:** `make build`  
**Lint:** `make lint`  
**Test:** `make test`  
**Dev Server:** `make run`

## Tech Stack
- Language: Go 1.21+
- Framework: Standard library with hexagonal architecture
- Dependencies:
  - archive/zip (standard library)
  - golang.org/x/net/html
  - github.com/unidoc/unipdf/v3

## Repository Structure
```
.
├── cmd/                    # Application entrypoints
├── internal/
│   ├── domain/            # Domain entities and business logic
│   ├── ports/             # Interface definitions (ports)
│   └── adapters/          # Implementation of ports (adapters)
├── pkg/                   # Public reusable packages
├── testdata/              # Test fixtures and sample data
├── examples/              # Example usage code
├── build/                 # Build artifacts (generated)
├── go.mod                 # Go module definition
├── Makefile              # Build automation
└── .gitignore            # Git ignore rules
```

## Code Style
- Follow existing patterns in the codebase
- No comments unless code is complex
- Match indentation and naming conventions
- Use hexagonal architecture (ports and adapters pattern)
