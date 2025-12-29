# Performance Benchmarks and Baseline Metrics

This document tracks baseline performance metrics and optimization targets for the validation library.

## Benchmark Overview

The benchmark suite measures performance across three key areas:
1. **Validation Throughput**: EPUB and PDF validation across various file sizes
2. **Reporter Formatting**: Error report generation with varying error counts
3. **Repair Operations**: Preview and apply performance for automated repairs

## Running Benchmarks

```bash
# Run all benchmarks
make test-bench

# Run specific benchmark category
go test -bench=BenchmarkEPUBValidation -benchmem ./tests/integration/...
go test -bench=BenchmarkPDFValidation -benchmem ./tests/integration/...
go test -bench=BenchmarkReporter -benchmem ./tests/integration/...
go test -bench=BenchmarkRepairService -benchmem ./tests/integration/...

# Compare benchmark results (requires benchstat)
go test -bench=. -benchmem -count=10 > old.txt
# Make changes...
go test -bench=. -benchmem -count=10 > new.txt
benchstat old.txt new.txt
```

## Baseline Performance Targets

### EPUB Validation Throughput

| File Size Category | Target Ops/sec | Target Time/op | Memory/op Target |
|-------------------|----------------|----------------|------------------|
| Small (<1MB)      | > 500 ops/sec  | < 2ms         | < 500 KB         |
| Medium (1-10MB)   | > 50 ops/sec   | < 20ms        | < 5 MB           |
| Large (>10MB)     | > 10 ops/sec   | < 100ms       | < 20 MB          |

**Rationale**: Small files should validate quickly for interactive workflows. Medium files represent typical ebooks. Large files should remain within acceptable batch processing times.

### PDF Validation Throughput

| File Size Category | Target Ops/sec | Target Time/op | Memory/op Target |
|-------------------|----------------|----------------|------------------|
| Small (<1MB)      | > 1000 ops/sec | < 1ms         | < 200 KB         |
| Medium (1-10MB)   | > 100 ops/sec  | < 10ms        | < 2 MB           |
| Large (>10MB)     | > 20 ops/sec   | < 50ms        | < 10 MB          |

**Rationale**: PDF structural validation is faster than full EPUB validation. Targets reflect simpler validation logic.

### Reporter Formatting Performance

| Error Count | JSON Target | Markdown Target | Text Target | Memory Target |
|------------|-------------|-----------------|-------------|---------------|
| 10 errors  | < 50µs     | < 100µs        | < 80µs      | < 50 KB       |
| 100 errors | < 500µs    | < 1ms          | < 800µs     | < 500 KB      |
| 1K errors  | < 5ms      | < 10ms         | < 8ms       | < 5 MB        |
| 10K errors | < 50ms     | < 100ms        | < 80ms      | < 50 MB       |

**Rationale**: Linear scaling with error count. JSON is fastest (direct serialization), Markdown requires more string processing, Text includes formatting logic.

### Repair Service Performance

| Operation           | Target Time/op | Memory/op Target | Notes                           |
|--------------------|----------------|------------------|---------------------------------|
| Preview (5 errors) | < 100µs       | < 100 KB         | Analysis only, no I/O           |
| Preview (50 errors)| < 1ms         | < 1 MB           | Scales linearly with error count|
| Apply (Small file) | < 50ms        | < 2 MB           | Includes file I/O and ZIP ops   |
| CreateBackup Small | < 10ms        | < 1 MB           | File copy operation             |
| CreateBackup Med   | < 100ms       | < 10 MB          | Larger file copy                |

**Rationale**: Preview operations should be fast for interactive use. Apply operations are I/O bound but should complete quickly for responsive UX.

## Performance Regression Detection

The CI pipeline will fail if benchmarks show:
- **Time/op regression > 20%**: Indicates algorithmic or implementation inefficiency
- **Memory/op regression > 30%**: Suggests memory leaks or inefficient data structures
- **Allocations/op regression > 40%**: Points to unnecessary object creation

### CI Benchmark Threshold Example

```yaml
# Performance regression thresholds
thresholds:
  time_regression: 1.20    # 20% slower fails
  memory_regression: 1.30   # 30% more memory fails
  allocs_regression: 1.40   # 40% more allocations fails
```

## Optimization Priorities

### High Priority
1. **EPUB Large File Validation**: Optimize manifest item processing
2. **Reporter Large Error Sets**: Implement streaming/buffering for >1K errors
3. **Repair Apply Operations**: Optimize ZIP file reconstruction

### Medium Priority
1. **PDF Medium File Validation**: Improve UniPDF parsing efficiency
2. **Reporter Memory Usage**: Reduce string allocations in formatting
3. **Backup Operations**: Consider streaming for large file copies

### Low Priority
1. **Small File Operations**: Already within targets
2. **Preview Operations**: Minimal I/O, acceptable performance

## Benchmark History

### v1.0.0 Baseline (Current)

To be established after initial implementation. Run:

```bash
go test -bench=. -benchmem ./tests/integration/... > benchmarks-v1.0.0.txt
```

### Expected Results Template

```
BenchmarkEPUBValidation_Small_Minimal-8                    500    2.5 ms/op     450 KB/op    2500 allocs/op
BenchmarkEPUBValidation_Medium_100Chapters-8               50     20 ms/op      4.5 MB/op    25000 allocs/op
BenchmarkEPUBValidation_Large_500Chapters-8                10     95 ms/op      18 MB/op     120000 allocs/op
BenchmarkPDFValidation_Small_Minimal-8                     1000   0.9 ms/op     180 KB/op    1000 allocs/op
BenchmarkReporter_JSON_SmallErrorSet-8                     30000  45 µs/op      40 KB/op     100 allocs/op
BenchmarkReporter_JSON_LargeErrorSet-8                     250    4.5 ms/op     4.5 MB/op    10000 allocs/op
BenchmarkRepairService_EPUB_Preview_SmallReport-8          15000  80 µs/op      90 KB/op     200 allocs/op
```

## Monitoring and Analysis

### Key Metrics to Track
- **ns/op**: Time per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

### Red Flags
- Linear operations showing O(n²) growth
- Memory usage growing faster than input size
- Allocation counts increasing disproportionately

### Profiling for Deep Analysis

```bash
# CPU profile
go test -bench=BenchmarkEPUBValidation_Large -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profile
go test -bench=BenchmarkReporter_JSON_LargeErrorSet -memprofile=mem.prof
go tool pprof mem.prof

# Trace analysis
go test -bench=BenchmarkRepairService -trace=trace.out
go tool trace trace.out
```

## Contributing Performance Improvements

When submitting performance optimizations:
1. Run benchmarks before and after changes
2. Use `benchstat` to compare results statistically
3. Include benchmark comparisons in PR description
4. Explain the optimization technique used
5. Verify no functionality regression

Example PR template section:
```
## Performance Impact

```
benchstat old.txt new.txt
name                                old time/op    new time/op    delta
EPUBValidation_Large_500Chapters    95.2ms ± 2%    72.1ms ± 1%  -24.26%  (p=0.000 n=10+10)

name                                old alloc/op   new alloc/op   delta
EPUBValidation_Large_500Chapters    18.0MB ± 0%    14.2MB ± 0%  -21.11%  (p=0.000 n=10+10)
```

Optimization: Replaced slice append in tight loop with pre-allocated capacity.
```

## Future Enhancements

- [ ] Add parallel validation benchmarks for batch processing
- [ ] Benchmark network I/O scenarios (remote file validation)
- [ ] Add benchmarks for concurrent validator usage
- [ ] Profile garbage collection impact under load
- [ ] Benchmark database operations if persistence layer added
