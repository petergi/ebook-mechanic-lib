# ebm-lib Architecture

**Style:** Hexagonal Architecture (Ports & Adapters)  
**Language:** Go 1.21+  
**Last Updated:** December 2025  
**Status:** Implementation Complete

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Hexagonal Architecture Pattern](#hexagonal-architecture-pattern)
3. [System Components](#system-components)
4. [Data Flow](#data-flow)
5. [Technology Stack](#technology-stack)
6. [Design Decisions](#design-decisions)

---

## Architecture Overview

ebm-lib implements a **hexagonal architecture** (also known as Ports & Adapters or Clean Architecture) that separates business logic from external concerns. This architectural style provides:

- **Independence from frameworks and libraries**
- **Testability** through dependency inversion
- **Flexibility** to change implementations without affecting core logic
- **Clear separation of concerns** between domain, ports, and adapters

### Key Principles

1. **Domain-Centric:** Business logic is isolated in the domain layer
2. **Dependency Inversion:** Dependencies point inward toward the domain
3. **Interface-Based:** Communication through well-defined ports (interfaces)
4. **Pluggable Adapters:** External concerns are implemented as adapters
5. **Framework Independence:** Core logic doesn't depend on external frameworks

---

## Hexagonal Architecture Pattern

### Conceptual Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│                     EXTERNAL WORLD                           │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   CLI App   │  │   Web API   │  │  Mobile App │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         │                 │                 │                │
└─────────┼─────────────────┼─────────────────┼────────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                    PRIMARY ADAPTERS                         │
│                  (Drive the Application)                    │
│                                                             │
│     ┌─────────────────────────────────────────┐           │
│     │      pkg/ebmlib (Public API)            │           │
│     │  - Simple facade over internal ports     │           │
│     │  - Type aliases for domain entities      │           │
│     │  - Convenience functions                 │           │
│     └──────────────────┬──────────────────────┘           │
│                        │                                    │
└────────────────────────┼────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                         PORTS                               │
│              (Interface Definitions)                        │
│                                                             │
│     ┌──────────────────────────────────────┐              │
│     │  internal/ports/                     │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  Validator Ports               │ │              │
│     │  │  - EPUBValidator               │ │              │
│     │  │  - PDFValidator                │ │              │
│     │  └────────────────────────────────┘ │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  Repair Ports                  │ │              │
│     │  │  - RepairService               │ │              │
│     │  │  - EPUBRepairService           │ │              │
│     │  │  - PDFRepairService            │ │              │
│     │  └────────────────────────────────┘ │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  Reporter Ports                │ │              │
│     │  │  - Reporter                    │ │              │
│     │  │  - MultiReporter               │ │              │
│     │  └────────────────────────────────┘ │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  Repository Ports              │ │              │
│     │  │  - Repository                  │ │              │
│     │  └────────────────────────────────┘ │              │
│     └──────────────────────────────────────┘              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                      DOMAIN LAYER                           │
│                   (Business Logic)                          │
│                                                             │
│     ┌──────────────────────────────────────┐              │
│     │  internal/domain/                    │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  Core Entities                 │ │              │
│     │  │  - ValidationReport            │ │              │
│     │  │  - ValidationError             │ │              │
│     │  │  - Severity                    │ │              │
│     │  │  - ErrorLocation               │ │              │
│     │  └────────────────────────────────┘ │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  Business Rules                │ │              │
│     │  │  - Error counting              │ │              │
│     │  │  - Validation logic            │ │              │
│     │  │  - Report aggregation          │ │              │
│     │  └────────────────────────────────┘ │              │
│     └──────────────────────────────────────┘              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                   SECONDARY ADAPTERS                        │
│                (Implement Port Interfaces)                  │
│                                                             │
│     ┌──────────────────────────────────────┐              │
│     │  internal/adapters/                  │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  epub/                         │ │              │
│     │  │  - EPUBValidator (impl)        │ │              │
│     │  │  - ContainerValidator          │ │              │
│     │  │  - NavValidator                │ │              │
│     │  │  - RepairService (impl)        │ │              │
│     │  └────────────────────────────────┘ │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  pdf/                          │ │              │
│     │  │  - StructureValidator          │ │              │
│     │  │  - RepairService (impl)        │ │              │
│     │  └────────────────────────────────┘ │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  reporter/                     │ │              │
│     │  │  - JSONReporter                │ │              │
│     │  │  - TextReporter                │ │              │
│     │  │  - MarkdownReporter            │ │              │
│     │  └────────────────────────────────┘ │              │
│     │                                       │              │
│     │  ┌────────────────────────────────┐ │              │
│     │  │  repository_impl.go            │ │              │
│     │  │  - Repository implementation   │ │              │
│     │  └────────────────────────────────┘ │              │
│     └──────────────────────────────────────┘              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                   EXTERNAL DEPENDENCIES                     │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │
│  │ File System │  │ archive/zip │  │unipdf/pdfcpu│       │
│  └─────────────┘  └─────────────┘  └─────────────┘       │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐                         │
│  │golang.org/x/│  │  Database   │                         │
│  │  net/html   │  │  (future)   │                         │
│  └─────────────┘  └─────────────┘                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

#### Domain Layer (`internal/domain/`)

**Purpose:** Contains core business entities and logic

**Responsibilities:**
- Define core data structures (ValidationReport, ValidationError)
- Implement business rules (error counting, validation status)
- Remain independent of external frameworks
- Be the most stable layer (least likely to change)

**Key Files:**
- `validation.go` - Validation report structures and methods
- `entity.go` - Core domain entities

**Dependencies:** None (pure Go standard library only)

#### Ports Layer (`internal/ports/`)

**Purpose:** Define interfaces that adapters must implement

**Responsibilities:**
- Define validator interfaces (EPUBValidator, PDFValidator)
- Define repair service interfaces (RepairService, EPUBRepairService, PDFRepairService)
- Define reporter interfaces (Reporter, MultiReporter)
- Define repository interfaces (Repository)
- Establish contracts between layers

**Key Files:**
- `validator.go` - Validator port definitions
- `repair.go` - Repair service port definitions
- `reporter.go` - Reporter port definitions
- `repository.go` - Repository port definitions

**Dependencies:** Domain layer only

#### Adapters Layer (`internal/adapters/`)

**Purpose:** Implement port interfaces with concrete functionality

**Responsibilities:**
- Implement validation logic for EPUB and PDF
- Implement repair strategies
- Implement report formatting
- Handle external dependencies (file I/O, parsing libraries)
- Transform between external formats and domain models

**Structure:**

```
internal/adapters/
├── epub/
│   ├── validator.go          # EPUBValidator implementation
│   ├── container_validator.go # ZIP and container validation
│   ├── nav_validator.go       # Navigation document validation
│   ├── repair_service.go      # EPUB repair implementation
│   └── ...
├── pdf/
│   ├── structure_validator.go # PDF structure validation
│   ├── repair_service.go      # PDF repair implementation
│   └── ...
├── reporter/
│   ├── json_reporter.go       # JSON format implementation
│   ├── text_reporter.go       # Text format implementation
│   └── markdown_reporter.go   # Markdown format implementation
└── repository_impl.go         # Repository implementation
```

**Dependencies:** Ports, Domain, External libraries

#### Public API Layer (`pkg/ebmlib/`)

**Purpose:** Provide simple, user-friendly API

**Responsibilities:**
- Expose clean, intuitive functions
- Hide internal complexity
- Provide type aliases for domain entities
- Offer convenience wrappers with sensible defaults
- Support both simple and context-aware operations

**Key Files:**
- `doc.go` - Package documentation
- `client.go` - Public API functions

**Dependencies:** Internal adapters (instantiates concrete implementations)

---

## System Components

### Validation System

```
┌──────────────────────────────────────────────────────────┐
│                   Validation Flow                        │
└──────────────────────────────────────────────────────────┘

User Application
      │
      ▼
ebmlib.ValidateEPUB("book.epub")
      │
      ▼
EPUBValidator Port
      │
      ▼
epub.EPUBValidator (Adapter)
      │
      ├─► ContainerValidator ───► ZIP structure check
      │                       ├─► Mimetype validation
      │                       └─► Container.xml validation
      │
      ├─► NavValidator ──────────► Navigation document check
      │                       ├─► Well-formedness
      │                       ├─► TOC structure
      │                       └─► Link validation
      │
      └─► (Additional validators as needed)
      
      │
      ▼
ValidationReport (Domain Entity)
      │
      ├─► Errors: []ValidationError
      ├─► Warnings: []ValidationError
      ├─► Info: []ValidationError
      └─► Metadata: map[string]interface{}
      
      │
      ▼
Return to User
```

### Repair System

```
┌──────────────────────────────────────────────────────────┐
│                     Repair Flow                          │
└──────────────────────────────────────────────────────────┘

User Application
      │
      ▼
ebmlib.RepairEPUB("broken.epub")
      │
      ▼
1. Validate File
      │
      ▼
2. RepairService.Preview(ctx, report)
      │
      ├─► Analyze each error
      ├─► Determine repairability
      ├─► Plan repair actions
      └─► Check safety levels
      │
      ▼
RepairPreview
      │
      ├─► Actions: []RepairAction
      ├─► CanAutoRepair: bool
      ├─► Warnings: []string
      └─► EstimatedTime: int64
      │
      ▼
3. RepairService.Apply(ctx, filePath, preview)
      │
      ├─► Create backup
      ├─► Execute repairs
      │   ├─► Mimetype fix
      │   ├─► Container.xml fix
      │   ├─► Navigation fixes
      │   └─► Content fixes
      │
      ▼
RepairResult
      │
      ├─► Success: bool
      ├─► ActionsApplied: []RepairAction
      ├─► BackupPath: string
      └─► Error: error
      │
      ▼
Return to User
```

### Reporting System

```
┌──────────────────────────────────────────────────────────┐
│                   Reporting Flow                         │
└──────────────────────────────────────────────────────────┘

ValidationReport (Domain)
      │
      ▼
ebmlib.FormatReport(report, format)
      │
      ▼
Reporter Port
      │
      ├─► FormatJSON    ──► JSONReporter
      ├─► FormatText    ──► TextReporter
      ├─► FormatMarkdown──► MarkdownReporter
      ├─► FormatHTML    ──► HTMLReporter (future)
      └─► FormatXML     ──► XMLReporter (future)
      │
      ▼
Formatted Output (string or file)
      │
      ├─► Console display
      ├─► API response
      ├─► Log file
      └─► Documentation
```

---

## Data Flow

### EPUB Validation Data Flow

```
Input: "book.epub"
      │
      ▼
┌─────────────────────────────────────┐
│  1. Read File                        │
│     - Open as ZIP                    │
│     - Read file list                 │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  2. Container Validation             │
│     - Check ZIP structure            │
│     - Validate mimetype              │
│     - Parse container.xml            │
│     - Extract OPF path               │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  3. Package Document Validation      │
│     - Parse OPF                      │
│     - Validate metadata              │
│     - Validate manifest              │
│     - Validate spine                 │
│     - Extract navigation path        │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  4. Navigation Validation            │
│     - Parse nav.xhtml                │
│     - Validate TOC structure         │
│     - Validate links                 │
│     - Check landmarks                │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  5. Content Validation (optional)    │
│     - Validate XHTML documents       │
│     - Check well-formedness          │
│     - Verify DOCTYPE                 │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  6. Aggregate Results                │
│     - Collect all errors             │
│     - Collect warnings               │
│     - Calculate statistics           │
│     - Set validation status          │
└──────────────┬──────────────────────┘
               │
               ▼
      ValidationReport
```

### PDF Validation Data Flow

```
Input: "document.pdf"
      │
      ▼
┌─────────────────────────────────────┐
│  1. Read File                        │
│     - Open file stream               │
│     - Read header                    │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  2. Header Validation                │
│     - Check %PDF- signature          │
│     - Validate version (1.0-1.7)     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  3. Parse with unipdf                │
│     - Load PDF document              │
│     - Access object catalog          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  4. Trailer Validation               │
│     - Check startxref                │
│     - Validate trailer dict          │
│     - Check %%EOF marker             │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  5. Cross-Reference Validation       │
│     - Parse xref table/stream        │
│     - Check for overlaps             │
│     - Validate object references     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  6. Catalog Validation               │
│     - Check /Type /Catalog           │
│     - Validate /Pages entry          │
│     - Check document structure       │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  7. Aggregate Results                │
│     - Collect all errors             │
│     - Map to standard error codes    │
│     - Set validation status          │
└──────────────┬──────────────────────┘
               │
               ▼
      ValidationReport
```

### Repair Data Flow

```
Input: ValidationReport
      │
      ▼
┌─────────────────────────────────────┐
│  1. Analyze Errors                   │
│     - Check each error code          │
│     - Determine repairability        │
│     - Assess safety level            │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  2. Plan Repairs                     │
│     - Create RepairAction list       │
│     - Order by dependency            │
│     - Calculate time estimate        │
│     - Generate warnings              │
└──────────────┬──────────────────────┘
               │
               ▼
        RepairPreview
               │
               ▼ (if user confirms)
┌─────────────────────────────────────┐
│  3. Backup Original                  │
│     - Copy to .backup                │
│     - Verify backup integrity        │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  4. Apply Repairs                    │
│     - Execute each action            │
│     - Validate each step             │
│     - Roll back on failure           │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  5. Verify Repairs                   │
│     - Re-validate file               │
│     - Check for new errors           │
│     - Confirm improvements           │
└──────────────┬──────────────────────┘
               │
               ▼
        RepairResult
```

---

## Technology Stack

### Core Language
- **Go 1.21+**
  - Chosen for: Performance, concurrency, standard library
  - Compilation target: Native binaries for multiple platforms
  - CGO: Disabled for maximum portability

### Standard Library Dependencies
- `archive/zip` - EPUB (ZIP) file handling
- `encoding/xml` - XML parsing for EPUB and PDF metadata
- `encoding/json` - JSON report formatting
- `io`, `io/fs` - File I/O operations
- `context` - Cancellation and timeout support
- `time` - Timestamps and duration tracking

### External Dependencies

#### Required
- **golang.org/x/net/html**
  - Purpose: HTML/XHTML parsing and validation
  - Used for: EPUB content document validation

- **github.com/unidoc/unipdf/v3**
  - Purpose: PDF parsing and structure access
  - Used for: PDF validation and repair
  - License: Commercial-friendly open source

#### Optional (Future)
- **github.com/pdfcpu/pdfcpu**
  - Alternative PDF processing library
  - May be used for specific repair operations

### Build & Development Tools
- **Make** - Build automation
- **golangci-lint** - Code quality and linting
- **go test** - Unit and integration testing
- **go mod** - Dependency management

### Testing
- Standard `testing` package
- Table-driven tests
- Test fixtures in `testdata/`
- Integration tests in `tests/integration/`

---

## Design Decisions

### 1. Why Hexagonal Architecture?

**Decision:** Use hexagonal (ports & adapters) architecture

**Rationale:**
- **Testability:** Easy to mock external dependencies via ports
- **Flexibility:** Swap implementations without changing core logic
- **Maintainability:** Clear separation of concerns
- **Framework Independence:** Core logic doesn't depend on external frameworks
- **Evolution:** Easy to add new validators, reporters, or repair strategies

**Trade-offs:**
- More files and interfaces than simpler architectures
- Initial learning curve for contributors
- Some boilerplate for adapter implementations

**Benefits Realized:**
- Can test domain logic without file I/O
- Easy to add new output formats (JSON, Text, Markdown)
- Simple to extend with new validators
- Clear boundaries for unit testing

### 2. Why Go?

**Decision:** Implement in Go rather than other languages

**Rationale:**
- **Performance:** Fast compilation and execution
- **Simplicity:** Easy to learn, maintain, and deploy
- **Standard Library:** Excellent built-in support for our needs
- **Concurrency:** Built-in goroutines for parallel validation
- **Portability:** Single binary deployment, cross-platform
- **Tooling:** Excellent development tools (go fmt, go vet, golangci-lint)

**Trade-offs:**
- Less expressive than some languages (no generics in early versions)
- Smaller ecosystem for ebook-specific libraries
- Manual error handling (no exceptions)

**Benefits Realized:**
- Fast validation even for large files
- Easy distribution (single binary)
- Excellent testing infrastructure
- Strong community support

### 3. Domain-First Design

**Decision:** Start with domain entities, then define ports, then implement adapters

**Rationale:**
- **Focus on business logic** before implementation details
- **Stable foundation:** Domain changes less frequently than adapters
- **Clear dependencies:** Always point inward toward domain
- **Testability:** Can test domain logic independently

**Benefits:**
- ValidationReport and ValidationError are simple, stable structs
- Business rules (error counting, severity) live in domain
- Adapters can change without affecting domain

### 4. Interface-Based Ports

**Decision:** Define behavior through interfaces rather than concrete types

**Rationale:**
- **Dependency Inversion:** High-level modules don't depend on low-level modules
- **Substitutability:** Easy to swap implementations
- **Mocking:** Simple to create test doubles
- **Extensibility:** New implementations can be added without changes to existing code

**Benefits:**
- Easy to test without file I/O
- Can add new validators without changing existing code
- Multiple reporter formats without changing validation logic

### 5. Context Support

**Decision:** Provide context-aware functions for all long-running operations

**Rationale:**
- **Cancellation:** Allow users to cancel long operations
- **Timeouts:** Enforce time limits on validation/repair
- **Tracing:** Support for distributed tracing (future)
- **Go Best Practice:** Idiomatic Go for long-running operations

**Implementation:**
- All validation and repair functions have `*WithContext` variants
- Non-context versions use `context.Background()`
- Context passed through all layers

### 6. Error Handling Strategy

**Decision:** Distinguish between operational errors and validation errors

**Rationale:**
- **Clarity:** Different types of problems require different handling
- **User Experience:** Users need to know if problem is file content or system issue
- **Recovery:** Operational errors may be transient; validation errors require fixes

**Implementation:**
- **Operational errors:** Returned as Go `error` (file not found, I/O failure)
- **Validation errors:** Contained in `ValidationReport.Errors`
- Never conflate the two types

**Example:**
```go
// Operational error (can't read file)
report, err := ValidateEPUB("missing.epub")
if err != nil {
    return err // Handle system error
}

// Validation error (file content invalid)
if !report.IsValid {
    // Handle validation errors
    for _, valErr := range report.Errors {
        log.Printf("%s: %s", valErr.Code, valErr.Message)
    }
}
```

### 7. Repair Safety Levels

**Decision:** Classify repairs by safety level and require different approval levels

**Rationale:**
- **User Trust:** Users need confidence repairs won't corrupt files
- **Automation:** Some repairs safe to automate, others require review
- **Risk Management:** Clear communication about potential impacts

**Levels:**
- **Very High:** Purely additive, no content changes (auto-approve)
- **High:** Structure changes only, content preserved (auto-approve)
- **Medium:** Heuristic-based, may need adjustment (require confirmation)
- **Low:** Potentially lossy or complex (require explicit approval)

**Benefits:**
- Users can set automation policies
- Clear documentation of repair impacts
- Builds user confidence

### 8. Structured Error Codes

**Decision:** Use structured, hierarchical error codes

**Format:** `<FORMAT>-<CATEGORY>-<NUMBER>`

**Rationale:**
- **Machine-Readable:** Easy to filter and categorize programmatically
- **Human-Friendly:** Codes are self-documenting
- **Extensibility:** New categories and codes can be added systematically
- **Specification Alignment:** Maps directly to EPUB and PDF spec sections

**Benefits:**
- Easy to filter errors by category
- Clear documentation mapping
- Consistent across EPUB and PDF
- Enables targeted error handling

### 9. Multiple Output Formats

**Decision:** Support JSON, Text, and Markdown report formats

**Rationale:**
- **JSON:** Machine-readable for APIs and automation
- **Text:** Human-readable for console and logs
- **Markdown:** Documentation and reports
- **Extensibility:** Architecture makes adding formats easy

**Implementation:**
- Each format is a separate adapter implementing `Reporter` port
- Factory pattern in public API selects appropriate reporter
- Consistent report options across formats

### 10. Test Data Organization

**Decision:** Separate test fixtures by validity and error type

**Structure:**
```
testdata/
├── epub/
│   ├── valid/
│   ├── invalid-container/
│   ├── invalid-nav/
│   └── invalid-metadata/
└── pdf/
    ├── valid/
    ├── invalid-header/
    ├── invalid-trailer/
    └── invalid-xref/
```

**Rationale:**
- **Clarity:** Easy to find relevant test files
- **Completeness:** Ensures coverage of all error types
- **Maintainability:** Simple to add new test cases
- **Documentation:** Test files serve as examples

---

## Appendix: File Organization

### Complete Directory Structure

```
.
├── cmd/
│   └── main.go                    # CLI entry point
│
├── internal/
│   ├── domain/                    # Core business entities
│   │   ├── entity.go
│   │   └── validation.go
│   │
│   ├── ports/                     # Interface definitions
│   │   ├── validator.go
│   │   ├── repair.go
│   │   ├── reporter.go
│   │   └── repository.go
│   │
│   └── adapters/                  # Implementations
│       ├── epub/
│       │   ├── validator.go
│       │   ├── container_validator.go
│       │   ├── nav_validator.go
│       │   └── repair_service.go
│       ├── pdf/
│       │   ├── structure_validator.go
│       │   └── repair_service.go
│       ├── reporter/
│       │   ├── json_reporter.go
│       │   ├── text_reporter.go
│       │   └── markdown_reporter.go
│       └── repository_impl.go
│
├── pkg/
│   └── ebmlib/                    # Public API
│       ├── doc.go
│       ├── client.go
│       └── README.md
│
├── testdata/                      # Test fixtures
│   ├── epub/
│   └── pdf/
│
├── tests/                         # Test suites
│   ├── integration/
│   └── unit/
│
├── examples/                      # Example usage
│   ├── basic_validation.go
│   ├── repair_example.go
│   └── custom_reporting.go
│
├── docs/                          # Documentation
│   ├── ARCHITECTURE.md            # This file
│   ├── USER_GUIDE.md
│   ├── ERROR_CODES.md
│   ├── SPEC.md
│   ├── specs/
│   │   ├── ebm-lib-EPUB-SPEC.md
│   │   └── ebm-lib-PDF-SPEC.md
│   └── adr/                       # Architecture Decision Records
│
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
├── Makefile                       # Build automation
├── .gitignore
└── README.md
```

### Key File Purposes

| File/Directory | Purpose | Dependencies |
|----------------|---------|--------------|
| `internal/domain/` | Core entities, business logic | None |
| `internal/ports/` | Interface definitions | Domain only |
| `internal/adapters/` | Port implementations | Ports, Domain, External libs |
| `pkg/ebmlib/` | Public API | Adapters (instantiates) |
| `cmd/` | Application entry points | pkg/ebmlib |
| `testdata/` | Test fixtures | None (data files) |
| `examples/` | Usage examples | pkg/ebmlib |

---

**For more information:**
- User Guide: `docs/USER_GUIDE.md`
- Error Codes: `docs/ERROR_CODES.md`
- API Documentation: `pkg/ebmlib/doc.go`
- Specifications: `docs/SPEC.md`
