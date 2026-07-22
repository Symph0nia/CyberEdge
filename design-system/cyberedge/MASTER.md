# CyberEdge Web Design System

Product: strictly read-only enterprise attack-surface observer.

Design dials: variance 3/10, motion 2/10, density 8/10.

## Foundation

Use the official IBM Carbon React and Carbon Styles packages. Do not mix another component system into the application.

The interface is an operations dashboard, not a marketing page and not a hacker-themed terminal. Avoid matrix green, neon effects, gradients, glass, oversized display type, decorative status dots, and fake live data.

## Tokens

- Theme: Carbon `g100` dark theme.
- Accent: Carbon blue 40 (`#78a9ff`) only.
- Background: Carbon gray 100 (`#161616`).
- Raised surface: Carbon gray 90 (`#262626`).
- Border: Carbon gray 80 (`#393939`).
- Primary text: Carbon gray 10.
- Secondary text: Carbon gray 30 (`#c6c6c6`) or lighter when required for WCAG AA.
- Typography: IBM Plex Sans for interface text, IBM Plex Mono for identifiers, timestamps, hashes, and numeric metrics.
- Shape: Carbon square geometry. Do not introduce rounded cards.
- Spacing: Carbon 8px-based scale. Dense tables use 8-16px internal spacing; sections use 32-48px.

## Information architecture

The persistent side navigation contains Overview, Assets, Tasks, Evidence, Audit, and Coverage. Each destination is a view into the same read model.

The overview order is:

1. Scope, asset, observation, and evidence metrics.
2. Filterable asset inventory.
3. Task history and terminal state.
4. Agent and Skill invocation audit.
5. Authorization scope coverage.
6. Evidence retention summary.

Do not render create, edit, delete, retry, cancel, triage, upload, or execute controls. Search, filtering, navigation, expansion, and export of already-visible read data are allowed.

## States and accessibility

- Provide structural skeletons while loading.
- Provide explicit empty states instead of blank tables.
- Announce request failures with `role="alert"`.
- Preserve visible keyboard focus from Carbon.
- Keep body text contrast at least 4.5:1.
- Use semantic tables with headers.
- At widths below 1056px, remove the persistent side navigation and give content the full viewport.
- At widths below 672px, stack headings and filters and permit tables to scroll horizontally.
- Respect `prefers-reduced-motion`; the interface has no automatic animation.

## Pre-delivery checks

- No Web mutation route or mutation control exists.
- One Carbon system and one blue accent are used.
- No emoji or hand-drawn icon is used; icons come from Carbon Icons.
- Loading, empty, error, desktop, tablet, and narrow layouts are covered.
- Security headers include CSP, `nosniff`, and no-referrer.
- Unknown API paths return JSON 404 rather than the SPA document.
