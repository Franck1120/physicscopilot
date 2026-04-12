# Testing Guide

## Overview

| Layer | Framework | Command |
|-------|-----------|---------|
| Go unit tests | `testing` + `testify` | `go test ./...` |
| Go e2e tests | `net/http/httptest` | `go test ./cmd/server/...` |
| Flutter unit | `flutter_test` | `flutter test` |
| Flutter widget | `flutter_test` | `flutter test` |
| Flutter integration | `integration_test` | `flutter test integration_test/` |

Run the full suite:
```bash
bash scripts/test.sh
bash scripts/test.sh --coverage   # with HTML reports
```

---

## Go Server Tests

### Structure

```
server/
├── internal/
│   ├── db/
│   │   ├── pool_test.go                # DB pool config
│   │   └── session_repository_test.go  # repository integration tests
│   ├── handlers/
│   │   ├── health_handler_test.go
│   │   ├── session_handler_test.go
│   │   ├── websocket_handler_test.go
│   │   ├── auth_handler_test.go
│   │   ├── feedback_handler_test.go
│   │   └── benchmark_test.go           # performance benchmarks
│   ├── logger/
│   │   ├── logger_test.go
│   │   └── security_test.go
│   ├── metrics/
│   │   ├── metrics_test.go
│   │   └── error_tracker_test.go
│   └── middleware/
│       └── (auth, rate limit tests)
└── cmd/server/
    ├── main_test.go                    # server startup tests
    └── e2e_test.go                     # end-to-end HTTP/WS tests
```

### Running specific tests

```bash
cd server

# Single package
go test ./internal/handlers/

# Single test
go test -run TestHealthHandler ./internal/handlers/

# Benchmarks
go test -bench=. -benchmem ./internal/handlers/

# Race detector (always use in CI)
go test -race ./...

# Verbose output
go test -v ./internal/handlers/

# Coverage profile
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out     # per-function summary
go tool cover -html=coverage.out     # open browser report
```

### Test patterns used

- **Table-driven tests**: `for _, tc := range cases { t.Run(tc.name, ...) }`
- **httptest.NewRecorder**: For HTTP handler tests (no real server)
- **Fake DB interface**: Handlers receive interface, tests inject mock
- **`testify/assert`**: For readable assertions

### Mocking

Handlers accept interfaces (`DBBackend`, `AIService`). Tests pass lightweight fakes:

```go
type fakeDB struct{}
func (f fakeDB) SaveSession(ctx context.Context, s *models.Session) error { return nil }
```

No code-gen mocks — keep mocks minimal and inline.

---

## Flutter Tests

### Structure

```
app/test/
├── services/
│   ├── api_service_test.dart
│   └── websocket_service_test.dart
└── widgets/
    └── session_screen_test.dart
```

### Running tests

```bash
cd app

# All tests
flutter test

# Single file
flutter test test/services/websocket_service_test.dart

# With coverage
flutter test --coverage
# Report at: coverage/lcov.info

# Verbose
flutter test --reporter=expanded
```

### Test patterns used

- **`FakeServer`** in websocket tests: lightweight `dart:io` HTTP server that upgrades to WebSocket. Tests real network I/O without mocking the WS channel.
- **`ProviderContainer`** for Riverpod: spin up providers in isolation, override dependencies.
- **`testWidgets`**: Widget tests use `pumpWidget` with a minimal `ProviderScope`.

### Coverage targets

| Component | Target | Notes |
|-----------|--------|-------|
| Services | >= 80% | Critical path — must be high |
| Providers | >= 70% | State logic |
| Widgets | >= 50% | UI rendering is hard to test |
| Models | >= 90% | Pure data, easy to test |

---

## Adding New Tests

### Go handler test template

```go
func TestMyHandler(t *testing.T) {
    tests := []struct {
        name       string
        method     string
        body       string
        wantStatus int
    }{
        {"happy path", http.MethodPost, `{"field":"value"}`, http.StatusCreated},
        {"bad request", http.MethodPost, `not json`, http.StatusBadRequest},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            req := httptest.NewRequest(tc.method, "/api/target", strings.NewReader(tc.body))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()

            handler := NewMyHandler(fakeDB{})
            handler.ServeHTTP(w, req)

            assert.Equal(t, tc.wantStatus, w.Code)
        })
    }
}
```

### Flutter service test template

```dart
void main() {
  group('MyService', () {
    test('description of behavior', () async {
      // Arrange
      final service = MyService(dependency: FakeDependency());

      // Act
      final result = await service.doSomething();

      // Assert
      expect(result, equals(expectedValue));
    });
  });
}
```

---

## CI Test Results

Tests run automatically on every push and PR via `.github/workflows/ci.yml`. Coverage artifacts are uploaded for each run.

To download the latest coverage report:
```bash
gh run download --name go-coverage
gh run download --name flutter-coverage
```
