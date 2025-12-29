# Performance Benchmarking Guide

This guide explains how to use the performance benchmarking suite to measure, monitor, and optimize validation library performance.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Benchmark Categories](#benchmark-categories)
- [Running Benchmarks](#running-benchmarks)
- [Analyzing Results](#analyzing-results)
- [CI Integration](#ci-integration)
- [Performance Optimization](#performance-optimization)
- [Best Practices](#best-practices)

## Overview

The benchmark suite measures performance across three key areas:

1. **Validation Throughput**: EPUB and PDF validation performance across various file sizes
2. **Reporter Formatting**: Error report generation with different error counts and formats
3. **Repair Operations**: Preview and apply performance for automated file repairs

All benchmarks are integrated into the CI pipeline to automatically detect performance regressions.

## Quick Start

```bash
# Run all benchmarks once
make test-bench

# Run benchmarks with result capture (recommended)
make bench

# Create performance baseline for future comparisons
make bench-baseline

# Compare current performance with baseline
make bench-compare
```

## Benchmark Categories

### 1. EPUB Validation Benchmarks

Measures validation throughput for EPUB files of various sizes:

```bash
# Small files (<1MB) - Interactive use cases
go test -bench=BenchmarkEPUBValidation_Small -benchmem ./tests/integration/...

# Medium files (1-10MB) - Typical ebooks
go test -bench=BenchmarkEPUBValidation_Medium -benchmem ./tests/integration/...

# Large files (>10MB) - Complex publications
go test -bench=BenchmarkEPUBValidation_Large -benchmem ./tests/integration/...
```

**Available Benchmarks:**
- `BenchmarkEPUBValidation_Small_Minimal` - Minimal valid EPUB
- `BenchmarkEPUBValidation_Medium_100Chapters` - 100-chapter book
- `BenchmarkEPUBValidation_Large_500Chapters` - 500-chapter book
- `BenchmarkEPUBValidation_Structure_Small` - Structure-only validation
- `BenchmarkEPUBValidation_Metadata_Small` - Metadata-only validation
- `BenchmarkEPUBValidation_Content_Medium` - Content validation

### 2. PDF Validation Benchmarks

Measures PDF structural validation performance:

```bash
# Run all PDF benchmarks
go test -bench=BenchmarkPDFValidation -benchmem ./tests/integration/...
```

**Available Benchmarks:**
- `BenchmarkPDFValidation_Small_Minimal` - Minimal valid PDF
- `BenchmarkPDFValidation_Medium_100Pages` - 100-page document
- `BenchmarkPDFValidation_Large_500Pages` - 500-page document
- `BenchmarkPDFValidation_Reader_Small` - In-memory validation

### 3. Reporter Formatting Benchmarks

Measures report generation performance with varying error counts:

```bash
# Test JSON reporter performance
go test -bench=BenchmarkReporter_JSON -benchmem ./tests/integration/...

# Test Markdown reporter performance
go test -bench=BenchmarkReporter_Markdown -benchmem ./tests/integration/...

# Test Text reporter performance
go test -bench=BenchmarkReporter_Text -benchmem ./tests/integration/...
```

**Error Set Sizes:**
- **Small**: 10 errors
- **Medium**: 100 errors
- **Large**: 1,000 errors
- **Very Large**: 10,000 errors

### 4. Repair Service Benchmarks

Measures repair operation performance:

```bash
# Test repair preview (analysis only)
go test -bench=BenchmarkRepairService.*Preview -benchmem ./tests/integration/...

# Test repair apply (with I/O)
go test -bench=BenchmarkRepairService.*Apply -benchmem ./tests/integration/...

# Test backup operations
go test -bench=BenchmarkRepairService.*Backup -benchmem ./tests/integration/...
```

## Running Benchmarks

### Basic Usage

```bash
# Run all benchmarks once
go test -bench=. -benchmem ./tests/integration/...

# Run specific benchmark
go test -bench=BenchmarkEPUBValidation_Large -benchmem ./tests/integration/...

# Run with pattern matching
go test -bench='EPUB.*Large' -benchmem ./tests/integration/...
```

### Advanced Options

```bash
# Run for longer duration (more accurate results)
go test -bench=. -benchmem -benchtime=5s ./tests/integration/...

# Run multiple iterations
go test -bench=. -benchmem -count=10 ./tests/integration/...

# Save results to file
go test -bench=. -benchmem > benchmarks.txt

# Run with timeout
go test -bench=. -benchmem -timeout=30m ./tests/integration/...
```

### Make Targets

```bash
# Quick benchmark run
make test-bench

# Run with result capture
make bench

# Create baseline (10 iterations for stability)
make bench-baseline

# Compare with baseline
make bench-compare
```

## Analyzing Results

### Understanding Benchmark Output

```
BenchmarkEPUBValidation_Small-8    500    2.5 ms/op    450 KB/op    2500 allocs/op
```

- `BenchmarkEPUBValidation_Small` - Benchmark name
- `-8` - Number of CPUs used (GOMAXPROCS)
- `500` - Number of iterations run
- `2.5 ms/op` - Time per operation (lower is better)
- `450 KB/op` - Memory allocated per operation (lower is better)
- `2500 allocs/op` - Number of allocations per operation (lower is better)

### Using benchstat

Install benchstat for statistical comparison:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

Compare two benchmark runs:

```bash
# Run baseline
go test -bench=. -benchmem -count=10 > old.txt

# Make changes...

# Run new benchmarks
go test -bench=. -benchmem -count=10 > new.txt

# Compare with statistical analysis
benchstat old.txt new.txt
```

Example output:

```
name                                old time/op    new time/op    delta
EPUBValidation_Large_500Chapters    95.2ms ± 2%    72.1ms ± 1%  -24.26%  (p=0.000 n=10+10)

name                                old alloc/op   new alloc/op   delta
EPUBValidation_Large_500Chapters    18.0MB ± 0%    14.2MB ± 0%  -21.11%  (p=0.000 n=10+10)
```

### Profiling

#### CPU Profiling

```bash
# Generate CPU profile
go test -bench=BenchmarkEPUBValidation_Large -cpuprofile=cpu.prof ./tests/integration/...

# Analyze with pprof
go tool pprof cpu.prof

# Common pprof commands:
# - top: Show top CPU consumers
# - list FunctionName: Show source code with CPU usage
# - web: Generate visual graph (requires graphviz)
```

#### Memory Profiling

```bash
# Generate memory profile
go test -bench=BenchmarkReporter_JSON_LargeErrorSet -memprofile=mem.prof ./tests/integration/...

# Analyze with pprof
go tool pprof mem.prof

# Look for memory leaks:
# - top: Show top memory allocators
# - list FunctionName: See allocation sites
```

#### Execution Trace

```bash
# Generate trace
go test -bench=BenchmarkRepairService -trace=trace.out ./tests/integration/...

# View trace
go tool trace trace.out
```

## CI Integration

### Automatic Performance Regression Detection

The CI pipeline automatically:

1. Runs benchmarks on every PR and push
2. Compares results with baseline from target branch
3. Detects regressions exceeding thresholds:
   - **Time/op > 20%**: Performance degradation
   - **Memory/op > 30%**: Memory usage increase
   - **Allocations/op > 40%**: Allocation increase

### Viewing Results in CI

1. Check the "Performance Benchmarks" job in GitHub Actions
2. Download benchmark artifacts from the workflow run
3. Review PR comments for automated benchmark comparisons

### Updating Baseline

To update the baseline after intentional changes:

```bash
# Run benchmarks
make bench-baseline

# Commit the baseline (if tracked) or push to update CI cache
```

## Performance Optimization

### Identifying Bottlenecks

1. **Run benchmarks** to establish current performance
2. **Profile** the slowest operations:
   ```bash
   go test -bench=BenchmarkSlow -cpuprofile=cpu.prof
   go tool pprof -http=:8080 cpu.prof
   ```
3. **Identify hot paths** in the flame graph
4. **Focus optimization** on the most expensive operations

### Common Optimization Techniques

#### 1. Reduce Allocations

```go
// Before: Multiple small allocations
func process() []string {
    result := []string{}
    for _, item := range items {
        result = append(result, item)
    }
    return result
}

// After: Pre-allocated capacity
func process() []string {
    result := make([]string, 0, len(items))
    for _, item := range items {
        result = append(result, item)
    }
    return result
}
```

#### 2. Use String Builder

```go
// Before: String concatenation
func buildMessage(parts []string) string {
    result := ""
    for _, part := range parts {
        result += part
    }
    return result
}

// After: strings.Builder
func buildMessage(parts []string) string {
    var builder strings.Builder
    for _, part := range parts {
        builder.WriteString(part)
    }
    return builder.String()
}
```

#### 3. Reuse Buffers

```go
// Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()
    // Use buffer...
}
```

#### 4. Optimize Loops

```go
// Before: Function call in condition
for i := 0; i < len(items); i++ {
    process(items[i])
}

// After: Cache length
n := len(items)
for i := 0; i < n; i++ {
    process(items[i])
}

// Or use range (often faster)
for _, item := range items {
    process(item)
}
```

### Optimization Workflow

1. **Measure baseline**: `make bench-baseline`
2. **Implement optimization**
3. **Run benchmarks**: `make bench`
4. **Compare results**: `make bench-compare`
5. **Verify no regressions**: Run full test suite
6. **Document changes**: Update BENCHMARKS.md if targets change

## Best Practices

### Writing Benchmarks

1. **Name descriptively**: `Benchmark<Component>_<Scenario>_<Size>`
2. **Use ResetTimer**: Call `b.ResetTimer()` after setup
3. **Report allocations**: Use `b.ReportAllocs()` to track memory
4. **Avoid optimizations**: Don't let compiler optimize away code
5. **Test realistic data**: Use production-like inputs

Example:

```go
func BenchmarkMyFeature_LargeInput(b *testing.B) {
    // Setup (not measured)
    input := createLargeTestData()
    
    // Reset timer before benchmark loop
    b.ResetTimer()
    b.ReportAllocs()
    
    // Benchmark loop
    for i := 0; i < b.N; i++ {
        result := MyFeature(input)
        
        // Prevent compiler optimization
        if result == nil {
            b.Fatal("unexpected nil")
        }
    }
}
```

### Running Benchmarks Reliably

1. **Close other applications** to reduce noise
2. **Run multiple iterations**: `-count=5` or more
3. **Use consistent hardware** for comparisons
4. **Disable CPU scaling** if possible:
   ```bash
   # Linux
   sudo cpupower frequency-set -g performance
   ```
5. **Check system load** before running:
   ```bash
   uptime  # Load should be < 1.0
   ```

### Interpreting Results

1. **Look for patterns**: Is growth linear, quadratic, or constant?
2. **Compare proportionally**: Is memory growing faster than input size?
3. **Check variance**: High variance (±%) indicates unstable benchmarks
4. **Validate improvements**: Run benchstat with n=10 for statistical significance
5. **Profile outliers**: Investigate unexpectedly slow operations

### Documenting Performance

When making performance-related changes:

1. **Include benchmark comparison** in PR:
   ```markdown
   ## Performance Impact
   
   ```
   benchstat old.txt new.txt
   name              old time/op    new time/op    delta
   Validation-8      10.2ms ± 2%     8.1ms ± 1%  -20.59%
   ```
   
   Optimization: Pre-allocated slice capacity for manifest items.
   ```

2. **Update baseline metrics** in `tests/integration/BENCHMARKS.md`
3. **Explain trade-offs** if any (e.g., memory vs. speed)
4. **Verify no regressions** in other benchmarks

## Troubleshooting

### Benchmarks Skipped

**Problem**: "Test file not found" message

**Solution**: Generate test fixtures:
```bash
make generate-fixtures
```

### Inconsistent Results

**Problem**: Large variance in benchmark results

**Solutions**:
- Close resource-intensive applications
- Increase iteration count: `-count=10`
- Use longer benchmark time: `-benchtime=5s`
- Run on consistent hardware
- Check system load with `uptime`

### Out of Memory

**Problem**: Benchmark crashes with OOM

**Solutions**:
- Reduce test data size
- Check for memory leaks with `-memprofile`
- Increase available memory
- Run subset of benchmarks

### Timeout

**Problem**: Benchmarks exceed timeout

**Solutions**:
- Increase timeout: `-timeout=60m`
- Skip slow benchmarks: `-bench='(?!Slow)'`
- Reduce iteration count
- Run in parallel if independent

## Additional Resources

- [Go Benchmark Documentation](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [benchstat Documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Dave Cheney's High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)

## Contributing

When contributing benchmarks or optimizations:

1. Follow benchmark naming conventions
2. Include baseline comparison in PR
3. Document optimization techniques used
4. Update performance targets if changed
5. Verify all existing benchmarks pass
6. Run benchstat for statistical validation

Questions? See [README.md](../README.md) or open an issue.
