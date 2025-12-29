
# ebm-lib – System Requirements Specification

**Project Name**: ebm-lib  
**Description**: Core validation library for EPUB files (EPUB 3.x) and basic well-formed PDF 1.7 files  
**Language**: Go (Golang)  
**Architecture**: Hexagonal (Ports & Adapters / Clean Architecture)  
**Status**: Implementation Phase  
**Last Updated**: December 28, 2025

## 1. Product Vision

Develop a robust, maintainable, and standards-compliant **core validation library** that:

- Provides comprehensive EPUB validation with feature parity for the essential checks performed by **epubcheck** (targeting EPUB 3.0 and above, with maximum compatibility toward EPUB 3.3)
- Offers basic validation for well-formed **PDF 1.7** files that can be correctly opened by standard PDF readers
- Is implemented following **hexagonal architecture** principles in **Go**
- Serves as a reliable foundation for future CLI, GUI, or integration use cases

## 2. Target Standards

| Format | Target Version                          | Goal / Notes                                                                 |
|--------|-----------------------------------------|------------------------------------------------------------------------------|
| EPUB   | EPUB 3.0 or later (aim for EPUB 3.3 compatibility) | Full validation of structure, metadata, content documents, and container rules |
| PDF    | Well-formed PDF 1.7                         | Syntactically correct files that open correctly in standard readers (no PDF/A, PDF/UA, etc. required at this stage) |

## 3. Required Validation Features

### 3.1 EPUB Validation (Core Requirements)

The library **must** detect and report issues in the following areas, with at least the coverage of **epubcheck**'s essential/common checks:

#### 3.1.1 Metadata (content.opf)
- Required Dublin Core elements: title, identifier, language
- Correct presence, format, and value of `dcterms:modified` date
- Valid `unique-identifier` attribute referencing an existing identifier

#### 3.1.2 HTML Content Documents
- Must be **well-formed XML** (properly nested tags, closed elements, valid entities)
- Must include a valid **HTML5 DOCTYPE**
- Must contain the required structural elements: `<html>`, `<head>`, `<body>`

#### 3.1.3 Container Format (OCF)
- Must be a **valid ZIP archive**
- First file in the ZIP archive **must** be named `mimetype` with **exact content** (no whitespace, no line breaks):  
  `application/epub+zip`
- Required file: `META-INF/container.xml`
  - Must exist and be non-empty
  - Must be valid XML
  - Must correctly reference at least one package document (`.opf`) via `full-path` attribute

### 3.2 PDF Validation (Basic)
- Must be a syntactically correct PDF 1.7 file
- Must open and render correctly in standard PDF readers
- Basic structural checks (header, cross-reference table, trailer, EOF marker, etc.)

## 4. Authoritative Reference Specifications

All validation logic **must** be derived exclusively from the following local documents (single source of truth):

- EPUB rules:  
  `/Users/petergiannopoulos/Documents/Projects/Personal/Active/ebm-lib/docs/specs/ebm-lib-EPUB-SPEC.md`
- PDF rules:  
  `/Users/petergiannopoulos/Documents/Projects/Personal/Active/ebm-lib/docs/specs/ebm-lib-PDF-SPEC.md`
- Architecture & design guidelines:  
  `/Users/petergiannopoulos/Documents/Projects/Personal/Active/ebm-lib/docs/ARCHITECTURE.md`

## 5. Architectural & Documentation Requirements

- **Architecture style**: Hexagonal / Ports & Adapters (Clean Architecture)
- **Primary language**: Go (Golang)
- **Decision documentation**:
  - All significant architectural, design, and technology decisions **must** be recorded as **Architectural Decision Records (ADRs)**
  - ADRs stored in:  
    `/Users/petergiannopoulos/Documents/Projects/Personal/Active/ebm-lib/docs/adr/`
  - One ADR file per meaningful decision

## 6. Version Control

**Version Control System**  
- Primary VCS: **Git**  
- Hosting: GitHub (preferred) or GitLab (as decided in ADR)  
- Repository URL: `https://github.com/[your-username]/ebmlib` (to be confirmed)  

**Branching Strategy**  
- Main branch: `main` (protected, only accepts merge requests/PRs)  
- Development branch: `develop` (integration branch for features and fixes)  
- Feature branches: `feature/<short-description>` (e.g. `feature/epub-container-validation`)  
- Bugfix branches: `fix/<issue-id>-<short-description>` (e.g. `fix/zip-mimetype-validation`)  
- Release branches: `release/vX.Y.Z` (for stabilization before tagging)  
- Hotfix branches: `hotfix/<issue-id>-<short-description>`  

**Commit Message Convention**  
Follow the **Conventional Commits** specification:  
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```
Common types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`, `ci`, `revert`

**Tagging & Releases**  
- Semantic Versioning: `vMAJOR.MINOR.PATCH` (e.g. `v0.1.0`, `v1.0.0`)  
- Tags created on `main` after successful release stabilization  
- Release notes generated from Conventional Commits (using tools like `git-chglog`, `auto-changelog`, or GitHub Releases)  

**Protected Branches & Workflows**  
- `main` and `develop` protected against direct pushes  
- Require PR reviews (minimum 1 approver)  
- Require passing CI checks before merge  
- Squash or rebase merges preferred (to be decided in ADR)

**Initial Setup**  
- `.gitignore` for Go projects (including IDE files, binaries, vendor/, coverage.out, etc.)  
- Initial commit with project skeleton, Makefile, README, and this SPEC.md

## 7. Implementation Status & Milestones (High-Level)

| Category                       | Status     | Notes / Target                                      |
|--------------------------------|------------|-----------------------------------------------------|
| Core EPUB container validation | 🟡         | ZIP + mimetype + container.xml                      |
| EPUB metadata validation       | 📋         | Planned                                             |
| EPUB content document checks   | 📋         | Planned                                             |
| Basic PDF 1.7 well-formedness  | 📋         | Planned                                             |
| CLI interface                  | 🟡         | Scaffolded / Makefile automation present            |
| Test coverage target           | 📋         | ≥80% (enforced in CI)                               |

## 8. Key Technology & Process Decisions

- **Language & runtime**: Go (current stable version)
- **Build & automation**: Makefile (setup, lint, test, coverage, build, docker)
- **Linting**: golangci-lint + custom Datadog rules
- **CI/CD**: Automated build, lint, test, coverage gating, Docker image generation & publishing
- **Testing**: Unit + integration tests with high coverage target

## 9. Assumptions & Open Questions

- Custom ZIP handling is mentioned for Swift components → clarify if Go core needs custom ZIP reader or can use standard `archive/zip`
- Exact subset of epubcheck rules to implement in phase 1 (full vs. critical subset)
- Future support for EPUB 3.3-specific features (e.g. new manifest properties, SSML, etc.)
- Whether to support legacy EPUB 2.0.1 (NCX) in the future

## 10. Next Steps (Recommended)

1. Finalize first ADR: choice of ZIP library & error handling strategy
2. Implement container-level EPUB checks (mimetype + container.xml)
3. Set up baseline test suite and CI pipeline with coverage gating
4. Begin implementing metadata parser & basic validation

---
*This specification serves as the primary requirements document for the ebm-lib project. All implementation must align with this document and the referenced local spec files.*
