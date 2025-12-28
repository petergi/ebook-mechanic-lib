# ADR-003: UniPDF Selection for PDF Parsing

## Status
Accepted

## Context
EPUB files can contain PDF documents as embedded resources or content files. To provide comprehensive content extraction, we need a PDF parsing library capable of extracting text and metadata from PDF files.

Options evaluated:
1. **UniPDF** (`github.com/unidoc/unipdf/v3`): Commercial-friendly PDF library with dual licensing
2. **pdfcpu** (`github.com/pdfcpu/pdfcpu`): Open-source PDF processor focused on manipulation
3. **go-pdf** (`github.com/EndFirstCorp/perigee`): Lightweight PDF text extraction
4. **Standard library only**: No PDF support, skip PDF content

Key evaluation criteria:
- Text extraction quality and reliability
- PDF version support (PDF 1.0 through 2.0)
- Licensing compatibility (permissive for commercial use)
- API design and ease of use
- Active maintenance and community support
- Performance and memory efficiency
- Documentation quality

## Decision
We will use UniPDF (`github.com/unidoc/unipdf/v3`) as our PDF parsing library.

UniPDF will be integrated through the adapter pattern:
- A port interface (`PDFParser`) defines PDF parsing operations
- A UniPDF adapter implements this interface
- Domain logic depends only on the port interface
- Alternative implementations can be swapped without affecting the domain

## Consequences

### Positive
- **Comprehensive Features**: Robust text extraction with support for various PDF encodings
- **PDF Standard Compliance**: Supports PDF versions 1.0 through 2.0
- **Commercial Friendly**: AGPL license with commercial license option available
- **Active Development**: Regular updates and active community
- **Rich API**: Provides access to text, images, metadata, and document structure
- **Good Documentation**: Well-documented with examples and guides
- **Production Ready**: Used in commercial products, battle-tested
- **Metadata Extraction**: Can extract author, title, creation date, and other metadata

### Negative
- **License Considerations**: AGPL requires source code disclosure for networked applications (commercial license available)
- **Dependency Size**: Larger dependency compared to lightweight alternatives
- **Learning Curve**: Rich feature set means more API surface to understand
- **Commercial Cost**: May require commercial license depending on use case

### Neutral
- **Performance**: Good performance for typical EPUB-embedded PDFs (usually small documents)
- **Memory Usage**: Reasonable memory footprint for document processing
- **Error Handling**: Comprehensive error reporting for malformed PDFs

## Alternatives Considered

### pdfcpu
- **Pros**: Apache 2.0 license, good for PDF manipulation (merge, split, watermark)
- **Cons**: Less focused on text extraction, heavier focus on PDF operations rather than parsing
- **Decision**: Better suited for PDF manipulation tools rather than content extraction

### go-pdf
- **Pros**: Lightweight, simple API
- **Cons**: Limited PDF version support, less active maintenance, missing features for complex PDFs
- **Decision**: Too limited for robust content extraction from varied PDF sources

### No PDF Support
- **Pros**: Zero dependencies for PDF handling
- **Cons**: Incomplete EPUB content extraction, poor user experience for EPUB files with PDF content
- **Decision**: PDF support is valuable for comprehensive EPUB processing

## Notes
The adapter pattern allows us to:
- Isolate UniPDF dependency behind an interface
- Write unit tests with mock PDF parsers
- Potentially swap to alternative libraries if licensing or feature requirements change
- Support multiple PDF parsing strategies (e.g., simple extraction vs. complex layout analysis)

If the AGPL license becomes problematic, we can either:
1. Purchase a commercial UniPDF license
2. Implement an adapter for an alternative library (minimal impact due to port abstraction)
3. Disable PDF parsing for specific deployment scenarios
