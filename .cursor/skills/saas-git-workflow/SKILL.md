---
name: saas-git-workflow
description: Enforces repository git workflow, branch strategy, and versioning discipline for SaaS template development. Use when creating branches, preparing commits, cutting releases, or updating version metadata.
---

# SaaS Git Workflow

Use this skill to keep branch flow, CI behavior, and versioning consistent.

## Branch strategy

- `main`: release branch (production-ready only).
- `develop`: integration branch (stabilization before release).
- `dev` or `feature/*`: feature work branches.

## PR flow

1. Create feature work in `dev` or `feature/<scope>`.
2. Open PR into `develop`.
3. Merge validated changes into `develop`.
4. Cut release PR from `develop` into `main`.
5. Tag release on `main` using template semantic versioning.

## Versioning model

- Use SemVer: `MAJOR.MINOR.PATCH`.
- Source of truth is repository `VERSION`.
- Bump:
  - `PATCH`: fixes/internal improvements
  - `MINOR`: backwards-compatible features
  - `MAJOR`: breaking changes

## Required checks before merge

```text
- [ ] CI passes on target branch
- [ ] VERSION follows SemVer format
- [ ] Changelog/release notes updated for main releases
- [ ] Docs updated for behavioral/contract changes
- [ ] No secrets or .env files in commit
```

## Commit hygiene

- Keep commits scoped and intentional.
- Use commit messages that explain "why" and "impact".
- Avoid combining refactors with feature work unless requested.

## Reject these actions

- Direct feature merges to `main`.
- Bumping version without documenting release intent.
- Force-pushing protected branches.
