# ADR-011: Reporting Output Formats

## Status
Accepted

## Context
Consumers need both human-readable and machine-readable outputs for validation and repair results.

## Decision
Provide JSON, Text, and Markdown report formats through reporter adapters.

## Consequences
### Positive
- JSON for automation and tooling
- Text for console usage
- Markdown for documentation and reports

### Negative
- Requires maintaining format parity and consistent options
