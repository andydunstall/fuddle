# Contributing

## Build
```
go build ./...
```

## Testing

### Types
Fuddle contains different types of tests:

#### Unit
Unit tests test a method directly in a single goroutine, with no external
dependencies. They must run in a few milliseconds.

Unit tests live in the packages under test in `foo_test.go`, which can be run
with `go test ./...`.

#### Integration
Integration tests either run a method directly, or via a network API. They may
run multiple goroutines though should still run very fast.

Each integration test must provision its own resources as part of the test,
instead of requiring setting up shared dependencies up front. This means tests
can be run with `go test` and ensures isolation between each test.

Integration tests live in the packages under test in `foo_integration_test.go`
with an `integration` build tag. So then can be run with
`go test ./... -tags=integration`.

### Testing Strategy
All code must aim for high unit test coverage. This forces the code to be well
designed, such as using dependency injection and avoiding complex dependencies.

Integration tests should only be used if you can't get sufficient coverage from
unit tests, such as when testing an network API. These tests shouldn't aim for
100% test coverage, though just cover the areas that can't be tested with
unit tests.

For example, if you build a HTTP server, each handler should have unit test
coverage, but you can also add an integration test to spin up a server locally
in a new goroutine and test the routes are integrated using a HTTP client.

## Style
Fuddle uses the [Uber style guide](https://github.com/uber-go/guide/blob/master/style.md).

Linters are run using `golangci-lint run`.
