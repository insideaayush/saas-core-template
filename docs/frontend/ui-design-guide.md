# UI Design Guide (shadcn/ui + Tailwind + Radix)

This template keeps UI styling intentionally simple, but the recommended direction for “production SaaS UI” is:

- shadcn/ui component patterns (Tailwind + Radix + variants).
- Tailwind CSS for styling and design tokens.
- Radix UI primitives for accessibility-correct components.

## Principles

- Accessibility is non-negotiable: keyboard support, focus rings, aria labels.
- Consistency beats perfection: prefer reusing a small set of primitives.
- Composition over abstraction: wrap Radix primitives with thin styling, don’t hide behavior.
- Design tokens, not ad-hoc colors: define a palette and use semantic tokens.

## Component conventions

- Put primitives in `frontend/components/ui/*` (shadcn-style).
- Use variants via `class-variance-authority` and utility composition via `frontend/lib/utils.ts`.
- Always support:
  - `disabled` state
  - `loading` state (with `aria-busy`)
  - focus-visible ring
  - consistent spacing and typography

## Adding components (recommended workflow)

This repo is set up to support shadcn’s generator config (`frontend/components.json`). To add new primitives:

- From `frontend/`, run `npx shadcn@latest add <component>`
- Commit generated files under `frontend/components/ui/`

Prefer adding a small set of primitives and using them everywhere instead of one-off custom markup.

## Recommended primitives

- `Button`, `Input`, `Textarea`, `Select`, `Badge`
- `Dialog`, `Popover`, `DropdownMenu`, `Tooltip`
- `Toast` / `Toaster`

## Layout guidance

- Top-level pages should have one primary CTA.
- Marketing pages:
  - hero + 3–6 feature bullets
  - social proof
  - pricing
  - FAQ
- App pages:
  - left nav or top nav
  - consistent page header with title + actions
  - empty states for first-run UX

## i18n-friendly UI

- Do not embed strings in deeply nested components; pass copy in from the page/screen layer.
- Avoid concatenating translated strings; prefer full sentences in message catalogs.
