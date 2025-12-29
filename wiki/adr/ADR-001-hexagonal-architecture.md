# ADR-001: Hexagonal Architecture Adoption

## Status
Accepted

## Context
We need to design a system for processing EPUB files that extracts and analyzes their content. The system must be maintainable, testable, and flexible enough to accommodate changes in file formats, storage mechanisms, and business logic without requiring extensive rewrites.

Key requirements include:
- Processing EPUB files (which are ZIP archives containing HTML and other resources)
- Extracting and parsing content from various formats (HTML, PDF)
- Supporting multiple storage backends (file system, in-memory, etc.)
- Enabling easy testing through dependency injection
- Separating business logic from infrastructure concerns

## Decision
We will adopt the Hexagonal Architecture pattern (also known as Ports and Adapters) to structure our application.

The architecture will be organized into three main layers:

1. **Domain Layer** (`internal/domain/`): Contains core business entities, value objects, and business rules. This layer has no dependencies on external frameworks or infrastructure.

2. **Ports Layer** (`internal/ports/`): Defines interfaces that describe how the application interacts with external systems. This includes:
   - Input ports: Interfaces for use cases and application services
   - Output ports: Interfaces for repositories, file systems, parsers, and other external dependencies

3. **Adapters Layer** (`internal/adapters/`): Implements the port interfaces, connecting the domain to concrete implementations:
   - Input adapters: CLI handlers, HTTP handlers (if needed)
   - Output adapters: File system implementations, ZIP readers, HTML parsers, PDF processors

The `pkg/` directory will contain public, reusable packages that can be imported by external projects.

## Consequences

### Positive
- **Testability**: Business logic can be tested in isolation by mocking port interfaces
- **Flexibility**: Easy to swap implementations (e.g., different storage backends or parsers) without affecting business logic
- **Maintainability**: Clear separation of concerns makes the codebase easier to understand and modify
- **Independence**: Domain logic remains independent of frameworks, databases, and external libraries
- **Adaptability**: New adapters can be added without modifying existing code

### Negative
- **Initial Complexity**: More upfront design work required to define proper boundaries
- **Indirection**: Additional interfaces and abstractions may seem like overhead for simple operations
- **Learning Curve**: Team members unfamiliar with hexagonal architecture need time to understand the pattern

### Neutral
- **File Organization**: Clear directory structure (`domain/`, `ports/`, `adapters/`) makes navigation predictable
- **Dependency Rules**: Dependencies flow inward (adapters → ports → domain), which must be enforced through code reviews
