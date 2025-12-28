# EPUB Test Fixtures

This directory contains EPUB test fixtures for validating the EPUB container validator.

## Test Files

The test files are created programmatically in the unit tests (`internal/adapters/epub/container_validator_test.go`) to ensure they are correct and up-to-date.

## Test Scenarios Covered

1. **Valid EPUB**: Properly formatted EPUB with all required components
2. **Invalid ZIP**: Not a valid ZIP archive
3. **Wrong Mimetype Content**: Mimetype file contains incorrect content
4. **Compressed Mimetype**: Mimetype file is compressed (should be stored uncompressed)
5. **Mimetype Not First**: Mimetype file is not the first entry in the ZIP
6. **Missing Container XML**: META-INF/container.xml is missing
7. **Invalid Container XML**: META-INF/container.xml is not valid XML
8. **No Rootfiles**: container.xml has no rootfile entries
9. **Empty Rootfile Path**: A rootfile has an empty full-path attribute
10. **Multiple Rootfiles**: Valid EPUB with multiple rootfile entries

## Error Codes

- `EPUB-CONTAINER-001`: ZIP Invalid - File is not a valid ZIP archive
- `EPUB-CONTAINER-002`: Mimetype Invalid - Mimetype file has incorrect content or compression
- `EPUB-CONTAINER-003`: Mimetype Not First - Mimetype file must be first in ZIP archive
- `EPUB-CONTAINER-004`: Container XML Missing - META-INF/container.xml is missing
- `EPUB-CONTAINER-005`: Container XML Invalid - META-INF/container.xml is malformed or invalid
