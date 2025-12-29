# ebm-lib Testing Guide

**Last Updated:** 

**Scope:** Current test assets and how to run them (backend Go, Angular Jest, Cypress E2E, API smokes)

## Test Structure Overview

```
ebm-lib/
├── internal/services/
│   ├── *_service_test.go              # Unit tests for business logic
│   ├── *_service_integration_test.go  # Integration tests with Testcontainers
│   └── ...
└── tests/
    └── [integration test suites]
```

## Backend Testing (Go)

### Unit Tests

Unit tests validate business logic in isolation using mocks and stubs.

**Inventory (Go):**

- Services: 

  

**Running:**

```bash
make test            # unit + integration
RUN_INTEGRATION=1 go test -v ./internal/services -run Integration
```

### Integration Tests

- Use `RUN_INTEGRATION=1` to enable Testcontainers for task service integration.
```bash
RUN_INTEGRATION=1 go test -v ./internal/services -run Integration
```

## API Smokes (Makefile)

```bash
make test-api-**
#fill in when api has been determined
```

## Coverage Goals

### Backend Coverage Targets

- **Services**: ≥80% line coverage
- **Critical paths**: 100% 

## Continuous Integration



## Debugging Tests



## Best Practices

### Test Organization

- ✅ Group related tests with `describe()` blocks
- ✅ Use descriptive test names that explain the scenario
- ✅ Follow AAA pattern (Arrange, Act, Assert)
- ✅ Use table-driven tests for multiple scenarios

### Test Isolation

- ✅ Clean up test data after each test
- ✅ Use mocks/stubs for external dependencies
- ✅ Avoid test interdependencies
- ✅ Reset state between tests

### Assertions

- ✅ Use specific assertions (not just `toBeTruthy()`)
- ✅ Test both success and failure paths
- ✅ Verify error messages are appropriate
- ✅ Assert on meaningful values

### E2E Best Practices

- ✅ Use `data-cy` attributes for element selection
- ✅ Avoid testing implementation details
- ✅ Focus on user-visible behavior
- ✅ Keep tests independent (no shared state)
- ✅ Use custom commands for repeated actions



## Troubleshooting

### Common Issues

**Tests timeout:**

- 

**E2E tests fail intermittently:**

- 

**Coverage reports missing:**

- Ensure coverage is enabled in test config
- Check file permissions
- Verify output directory exists

**Integration tests fail:**

- 

## Performance Testing

Load testing and performance metrics:

```bash

```

## Additional Resources

- [Go Testing Docs](https://golang.org/pkg/testing/)
- [Angular Testing Guide](https://angular.io/guide/testing)
- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [Cypress Documentation](https://docs.cypress.io/)
- [Testcontainers Go](https://golang.testcontainers.org/)
