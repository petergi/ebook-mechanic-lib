# Benchmark Quick Reference Card

## Common Commands

```bash
# Quick benchmark run (1 iteration)
make test-bench

# Create baseline (10 iterations for stability)
make bench-baseline

# Compare with baseline (5 iterations)
make bench-compare

# Run specific benchmark
go test -bench=BenchmarkEPUBValidation_Large -benchmem ./tests/integration/...

# Run with profiling
go test -bench=BenchmarkName -cpuprofile=cpu.prof -memprofile=mem.prof ./tests/integration/...

# Analyze profile
go tool pprof cpu.prof
```

## Performance Targets (at a glance)

| Operation | Target | Memory |
|-----------|--------|--------|
| EPUB Small | < 2ms | < 500KB |
| EPUB Medium | < 20ms | < 5MB |
| EPUB Large | < 100ms | < 20MB |
| PDF Small | < 1ms | < 200KB |
| PDF Medium | < 10ms | < 2MB |
| PDF Large | < 50ms | < 10MB |
| Reporter (100) | < 1ms | < 500KB |
| Reporter (1K) | < 10ms | < 5MB |
| Repair Preview | < 100µs | < 100KB |

## CI Regression Thresholds

- ⚠️ Time/op: **+20%** = regression
- ⚠️ Memory/op: **+30%** = regression  
- ⚠️ Allocations/op: **+40%** = regression

## Benchmark Output Format

```
BenchmarkName-8    500    2.5 ms/op    450 KB/op    2500 allocs/op
                   │      │            │            │
                   │      │            │            └─ Allocations/operation
                   │      │            └─ Memory/operation
                   │      └─ Time/operation
                   └─ Iterations
```

## Quick Profiling

```bash
# CPU hotspots
go test -bench=Slow -cpuprofile=cpu.prof ./tests/integration/...
go tool pprof -top cpu.prof

# Memory allocations
go test -bench=Slow -memprofile=mem.prof ./tests/integration/...
go tool pprof -top mem.prof

# Interactive
go tool pprof cpu.prof
> top       # Show top consumers
> list Func # Show function details
> web       # Visual graph (needs graphviz)
```

## Benchmark Categories

### EPUB Validation
- `BenchmarkEPUBValidation_Small_*` - Files < 1MB
- `BenchmarkEPUBValidation_Medium_*` - Files 1-10MB
- `BenchmarkEPUBValidation_Large_*` - Files > 10MB

### PDF Validation
- `BenchmarkPDFValidation_Small_*` - Files < 1MB
- `BenchmarkPDFValidation_Medium_*` - Files 1-10MB
- `BenchmarkPDFValidation_Large_*` - Files > 10MB

### Reporter Formatting
- `BenchmarkReporter_*_SmallErrorSet` - 10 errors
- `BenchmarkReporter_*_MediumErrorSet` - 100 errors
- `BenchmarkReporter_*_LargeErrorSet` - 1,000 errors
- `BenchmarkReporter_*_VeryLargeErrorSet` - 10,000 errors

### Repair Service
- `BenchmarkRepairService_*_Preview_*` - Analysis only
- `BenchmarkRepairService_*_Apply_*` - Full repair with I/O
- `BenchmarkRepairService_CreateBackup_*` - File copy ops

## Tips for Accurate Results

1. **Close other applications** to reduce system noise
2. **Run multiple times**: `-count=5` or higher
3. **Use longer benchtime**: `-benchtime=5s` for stability
4. **Check system load**: `uptime` should show < 1.0
5. **Compare statistically**: Use `benchstat` for significance testing

## Writing New Benchmarks

```go
func BenchmarkFeature_Scenario(b *testing.B) {
    // Setup (not measured)
    data := setupTestData()
    
    // Reset timer before benchmark loop
    b.ResetTimer()
    b.ReportAllocs()
    
    // Benchmark loop
    for i := 0; i < b.N; i++ {
        result := Feature(data)
        if result == nil {
            b.Fatal("unexpected nil")
        }
    }
}
```

## Common Issues

**Skipped benchmarks**: Run `make generate-fixtures`

**High variance**: Close apps, increase `-count`, check system load

**OOM errors**: Reduce test data size or skip large benchmarks

**Timeout**: Increase `-timeout=30m` or skip slow benchmarks

## Resources

- Full guide: [docs/BENCHMARKING.md](BENCHMARKING.md)
- Baseline metrics: [tests/integration/BENCHMARKS.md](../tests/integration/BENCHMARKS.md)
- Test integration: [tests/integration/README.md](../tests/integration/README.md)

## Quick Decision Tree

```
Need to...
├─ Run quick check? → make test-bench
├─ Start optimization work? → make bench-baseline
├─ Check for regressions? → make bench-compare
├─ Find bottleneck? → Profile (cpu/mem)
├─ Compare two versions? → benchstat old.txt new.txt
└─ Add new benchmark? → See BENCHMARKING.md
```
