---
name: go-saas-patterns
description: Applies Go-specific design patterns for testability, extensibility, and readability in SaaS backends. Use when implementing or refactoring Go services, handlers, repositories, middleware, webhooks, and domain logic.
---

# Go SaaS Patterns

Use this skill for production-shaped Go code that is easy to test and evolve.

## Design goals

- Keep business logic independent from transport and provider SDKs.
- Prefer explicit dependencies and small interfaces.
- Make side effects easy to mock in tests.
- Keep request handlers thin and deterministic.

## Core patterns

## 1) Package boundaries by responsibility

- `internal/api`: HTTP transport only (decode, validate, map errors to status codes).
- `internal/<domain>`: business rules and orchestration.
- `internal/<provider>` adapters: external APIs and SDK calls.
- `internal/db` repositories: persistence concerns.

Never place business rules directly in router/handler code.

## 2) Constructor-injected dependencies

- Use `NewService(...)` constructors with explicit dependencies.
- Avoid global state and hidden singletons.
- Validate required dependencies at construction time.

Example:

```go
type Service struct {
    repo UserRepo
    clock Clock
}

func NewService(repo UserRepo, clock Clock) *Service {
    return &Service{repo: repo, clock: clock}
}
```

## 3) Interface-at-boundary pattern

- Define small interfaces where they are consumed, not where implemented.
- Keep interfaces focused (`CreateCheckoutSession`, not broad provider contracts).
- Use compile-time checks for adapter compliance when useful.

## 4) Context and timeouts

- Pass `context.Context` through all I/O boundaries.
- Add request-scoped timeouts in handlers and external calls.
- Never ignore cancellation for DB or HTTP operations.

## 5) Error modeling

- Wrap errors with context (`fmt.Errorf("load user: %w", err)`).
- Use sentinel/domain errors for branchable behavior (auth, not found, forbidden).
- Keep HTTP status mapping in transport layer.

## 6) Idempotent side-effect handlers

- For webhooks, upsert by stable external IDs.
- Treat event replay as normal operation.
- Validate signatures before payload processing.

## Readability conventions

- Keep functions short and intention-revealing.
- Prefer plain data structs over deep inheritance-like patterns.
- Use naming that reflects domain intent (`ResolveOrganization`, `UpsertSubscription`).
- Keep comments for non-obvious constraints, not trivial narration.

## Testability checklist

```text
- [ ] Service logic testable without HTTP server
- [ ] External calls behind interfaces
- [ ] No global mutable state in logic paths
- [ ] Domain errors asserted directly in tests
- [ ] Time/randomness abstracted when behavior depends on them
- [ ] Handler tests cover status mapping and payload validation
```

## Testing strategy

- Unit tests: services and policies with fake repos/providers.
- Contract tests: adapter behavior against provider API assumptions.
- Integration tests: DB query behavior and migration compatibility.
- HTTP tests: `httptest` for route wiring and middleware behavior.

## Reject these patterns

- Fat handlers with DB and provider calls mixed together.
- Catch-all interfaces with unrelated methods.
- `panic` for expected business failures.
- Logging secrets or full webhook payloads.
