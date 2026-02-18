---
name: typescript-saas-patterns
description: Applies TypeScript-specific design patterns for testability, extensibility, and readability in SaaS frontend and shared logic. Use when implementing or refactoring Next.js routes, UI flows, API clients, domain modules, and state management.
---

# TypeScript SaaS Patterns

Use this skill for maintainable TypeScript code across UI, API clients, and domain logic.

## Design goals

- Separate domain logic from framework glue.
- Keep types explicit at module boundaries.
- Make async and side effects testable.
- Prefer composable, single-purpose modules.

## Core patterns

## 1) Layered module boundaries

- `app/*`: route/page composition and framework lifecycle.
- `lib/api.ts` or `services/*`: network request wrappers.
- `domain/*`: pure business logic and policies.
- `components/*`: presentational UI with minimal side effects.

Avoid putting business rules directly inside page components.

## 2) Typed boundary contracts

- Validate external inputs at boundaries (request payloads, URL params, env vars).
- Use explicit request/response types for API calls.
- Use discriminated unions for state/result modeling.

Example:

```ts
type LoadState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "ready"; data: ViewerResponse };
```

## 3) Dependency-injection friendly services

- Wrap fetch/provider SDK calls in service functions.
- Pass adapters into domain functions for unit testing.
- Keep components consuming service interfaces, not raw SDK calls.

## 4) Pure logic first

- Extract non-UI calculations and policies into pure functions.
- Keep React hooks/components focused on orchestration and rendering.
- Unit test pure modules directly with minimal mocking.

## 5) Predictable async flows

- Centralize request lifecycle (`loading/error/success`) explicitly.
- Handle cancellation/race conditions in effects.
- Prefer single-responsibility hooks over giant multi-purpose hooks.

## 6) Extensibility via composition

- Prefer composition and prop-driven configuration over inheritance-like patterns.
- Keep feature flags and plan entitlements in dedicated policy utilities.
- Avoid hardcoding provider-specific assumptions in shared domain modules.

## Readability conventions

- Use clear names (`createCheckoutSession`, `fetchViewer`) and small functions.
- Keep files cohesive; split when module has mixed concerns.
- Favor early returns and guard clauses.
- Minimize deeply nested conditional rendering.

## Testability checklist

```text
- [ ] Domain logic is framework-agnostic and unit tested
- [ ] API client calls are wrapped and mockable
- [ ] UI tests focus on behavior, not implementation details
- [ ] Error/empty/loading states are explicitly represented
- [ ] No hidden dependency on process/env in deep modules
- [ ] Provider-specific logic is isolated
```

## Testing strategy

- Unit tests: domain utilities, mappers, entitlement rules.
- Component tests: render states and user interactions.
- Integration tests: page + API client interaction boundaries.
- Optional E2E: sign-in, tenant selection, pricing/checkout redirection flow.

## Reject these patterns

- Fat page components with mixed network, policy, and presentation logic.
- `any` in domain-critical paths.
- Silent catch blocks that swallow failures.
- Duplicated request logic across routes/components.
