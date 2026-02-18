# Git Branching and Versioning

This template uses a release-oriented branch model and SemVer versioning.

## Branch model

- `main`
  - Release branch.
  - Only release-ready code should be merged here.
- `develop`
  - Integration branch for validated feature work.
  - Default PR target for active development.
- `dev` and `feature/*`
  - Feature branches for implementation work.
  - PR target should be `develop`.

## Recommended flow

1. Branch from `develop` into `feature/<name>` (or use `dev`).
2. Implement and open PR into `develop`.
3. Stabilize on `develop` with CI passing.
4. Open release PR from `develop` into `main`.
5. Tag release on `main` and publish release notes.

## Versioning

- Source of truth: repository `VERSION` file.
- Format: `MAJOR.MINOR.PATCH` (SemVer).

Version bump guidance:

- `PATCH` for bug fixes or internal improvements.
- `MINOR` for backwards-compatible features.
- `MAJOR` for breaking changes.

## CI expectations

- CI runs on pushes for `main`, `develop`, `dev`, and `feature/*`, and on PRs targeting `main`, `develop`, and `dev`.
- CI validates:
  - backend tests/build
  - frontend lint/typecheck/build
  - `VERSION` format (SemVer)

## Release branch policy

- `main` should be protected with required CI checks.
- No direct feature commits to `main`.
- Merge from `develop` to `main` using reviewed PRs.
