# Go Project

A Go application using hexagonal architecture with clean separation of concerns.

## Project Structure

```
.
├── cmd/                    # Application entrypoints
├── internal/
│   ├── domain/            # Domain entities and business logic
│   ├── ports/             # Interface definitions (ports)
│   └── adapters/          # Implementation of ports (adapters)
├── pkg/                   # Public reusable packages
├── testdata/              # Test fixtures and sample data
└── examples/              # Example usage code
```

## Getting Started

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
make install
```

### Building

```bash
make build
```

### Running

```bash
make run
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Generate coverage report
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run vet
make vet

# Run linter
make lint
```

## Make Targets

Run `make help` to see all available targets with descriptions.

## Dependencies

- `archive/zip` - Standard library for ZIP archive handling
- `golang.org/x/net/html` - HTML parsing
- `github.com/unidoc/unipdf/v3` - PDF processing

## License

TBD
