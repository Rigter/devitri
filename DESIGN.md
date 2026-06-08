---
project: Devitri
type: Design System
version: Nano Zinc (v1)
status: active
tags: [design-system, sveltekit, dashboard, obsidian]
updated: 2026-06-02
fonts: [Archivo, IBM Plex Mono]
---

# Nano v1 — Zinc Design System

**Philosophy:** Minimal with purpose, reading first, no decoration.

Nano Zinc is a mathematically neutral achromatic editorial design system built for writers and thinkers. It prioritizes long-form Markdown reading and file navigation.

The interface conveys **operational calm**: maximum visual silence, hierarchy through typography, and color as a signal — never as decoration. The user should forget there is a dashboard and focus on their notes.

Chromatic bias is zero (hue ~0°): neutral grey, neither cool blue-leaning nor warm yellow-brown.

Implementation lives in `frontend/src/lib/styles/themes/nano-zinc.css`. Components must use CSS custom properties only (`var(--bg)`, `var(--text)`, etc.) — no hardcoded hex in `.svelte` files.

---

## Colors

### Dark mode (default)

| Token | Value | Use |
|-------|-------|-----|
| `--bg` | `#18181B` | Main background |
| `--surface` | `#1E1E21` | Sidebars, cards, elevated panels |
| `--border` | `#3F3F46` | Dividers, component borders |
| `--text` | `#E4E4E7` | Primary text |
| `--muted` | `#A1A1AA` | Metadata, timestamps, inactive UI |
| `--color-conflict` | `#F09595` | Conflict state |
| `--color-success` | `#4ADE80` | Success / healthy sync |

### Light mode

| Token | Value | Use |
|-------|-------|-----|
| `--bg` | `#E4E4E7` | Main background |
| `--surface` | `#EBEBEE` | Elevated surfaces |
| `--border` | `#C5C5CC` | Borders |
| `--text` | `#18181B` | Primary text |
| `--muted` | `#71717A` | Secondary text |
| `--color-conflict` | `#A32D2D` | Conflict (with light bg `#FCEBEB`) |
| `--color-success` | `#166634` | Success (with light bg `#DCFCE7`) |

Theme switching: `[data-theme="dark"]` on `<html>` (see `frontend/src/lib/stores/theme.ts`).

---

## Typography

| Role | Font | Weight | Use |
|------|------|--------|-----|
| Headings (h1–h3) | Archivo | 500–600 | Section titles, logo |
| Body / UI | IBM Plex Mono | 400–500 | Body, labels, paths, code, badges |

### Scale

- **Display (2.5rem)** — landing / onboarding only
- **H1 (2rem)** — page title
- **H2 (1.25rem)** — vault name, column headers
- **H3 (1.125rem)** — card titles
- **Body (0.875rem)** — paragraphs, forms, markdown preview
- **Label (0.75rem)** — badges, chips (optional uppercase)
- **Caption (0.75rem)** — timestamps, breadcrumbs

---

## Layout

**3-zone matrix:** sidebar (280px), top bar (56px), flexible content.

- Panel padding: 24px
- List item gap: 8px
- Dashboard grid: 16px gap; responsive columns

---

## Elevation

**No box shadows.** Hierarchy via tonal layers and 1px borders (`--border`).

- Modals: overlay `rgba(0, 0, 0, 0.6)`
- Dropdowns float above content without shadow

---

## Border radius

| Token | Value | Use |
|-------|-------|-----|
| `sm` | 4px | Buttons, inputs, badges |
| `md` | 8px | Cards, panels |
| `lg` | 12px | Large containers, modals |
| `full` | 9999px | Avatars, icon buttons |

Focus: 1px border normal, 2px on focus (no browser outline).

---

## Components

### Buttons

Primary: `bg` + `text` + `border`. Hover inverts toward `border`. No gradients or scale animations.

### Inputs

Monospace body font; 1px → 2px border on focus.

### Cards

`surface` background, 1px `border`, 24px padding. Title in Archivo, content in IBM Plex Mono.

### Miller Columns

`surface` columns, 1px right divider. Active row: `bg` background; optional 2px left accent `muted`.

### Badges

- **Conflict:** diluted red background, `color-conflict` text
- **Success:** diluted green background, `color-success` text

---

## Do's and Don'ts

- **Do** use color only for state (conflict, success, danger).
- **Don't** use shadows, gradients, or decorative depth.
- **Do** respect the type scale.
- **Don't** mix arbitrary border radii in one view.
- **Do** use IBM Plex Mono for paths, timestamps, and technical strings.
- **Don't** add borders to file tables; use padding and horizontal rules.
- **Do** meet WCAG AA contrast (4.5:1) on surface backgrounds.

---

## Future themes

```
frontend/src/lib/styles/themes/
  nano-zinc.css      # v1 (default)
  nano-stone.css     # planned: warm neutrals
  nano-charcoal.css  # planned: deep black
```

---

*See also: `FOUNDATION.md` §5 (product context) and `frontend/src/lib/styles/devitri-markdown.css` (preview typography).*
