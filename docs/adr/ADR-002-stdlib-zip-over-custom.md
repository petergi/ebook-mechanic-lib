# ADR-002: Go Standard Library archive/zip Usage

## Status
Accepted

## Context
EPUB files are fundamentally ZIP archives with a specific internal structure (mimetype file, META-INF directory, content files). To read and process EPUB files, we need reliable ZIP archive handling capabilities.

Options considered:
1. **Go standard library `archive/zip`**: Built-in package for reading and writing ZIP archives
2. **Third-party ZIP libraries**: External packages like `github.com/alexmullins/zip` (supports encryption)
3. **Custom ZIP implementation**: Build our own ZIP reader tailored to EPUB needs

Key considerations:
- EPUB files use standard ZIP compression without encryption
- Need to read ZIP central directory and extract individual files
- Must handle various compression methods (stored, deflated)
- Reliability and long-term maintenance
- Performance for reading potentially large archives
- Dependency management and security

## Decision
We will use the Go standard library `archive/zip` package for all ZIP archive operations.

The standard library provides:
- `zip.Reader` for reading ZIP archives from `io.ReaderAt`
- `zip.ReadCloser` for reading ZIP files from filesystem
- Access to individual file entries with metadata
- Built-in support for common compression methods
- No external dependencies required

## Consequences

### Positive
- **Zero Dependencies**: No external packages to manage or update for core ZIP functionality
- **Stability**: Standard library has extensive testing and is backed by the Go team
- **Performance**: Well-optimized implementation suitable for most use cases
- **Security**: Security vulnerabilities are addressed through Go releases
- **Compatibility**: Guaranteed to work across all platforms Go supports
- **Simplicity**: Well-documented API with extensive examples
- **Future-proof**: Will continue to be maintained as part of the Go standard library

### Negative
- **Limited Features**: No built-in support for ZIP encryption (not needed for EPUB)
- **No Streaming Write**: Writing large ZIP files requires keeping entries in memory (not a concern for reading)
- **Basic Error Messages**: Error messages may be less descriptive than specialized libraries

### Neutral
- **EPUB Compliance**: Standard library handles all ZIP features required by EPUB specification
- **Performance Trade-offs**: Adequate for typical EPUB file sizes (< 100MB)
- **Memory Usage**: Requires random access, so very large files need to be seekable

## Notes
If future requirements demand ZIP encryption support (e.g., for DRM-protected EPUB files), we can introduce a specialized library through the adapter pattern without affecting the domain layer. The port interface abstractions allow for such extensions without breaking changes.
