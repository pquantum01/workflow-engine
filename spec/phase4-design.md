# Phase 4 — Frontend Design Language

## Design Principles
- **Calm + functional.** No decoration for decoration's sake.
- **Information-dense but not cluttered.** Tables over cards where data is tabular.
- **State communicates status.** Color carries meaning (green=active, amber=warning, red=error/cancelled, gold=terminal, gray=deleted/cancelled).
- **Fast perceived performance.** Skeletons immediately, stale data labeled, polling intervals tuned.

---

## Color Palette

### Light Mode
```
Background:         #FAFAFA  (near-white)
Surface:            #FFFFFF  (card/panel bg)
Border:             #E4E4E7  (zinc-200)
Text primary:       #18181B  (zinc-900)
Text secondary:     #71717A  (zinc-500)
Text muted:         #A1A1AA  (zinc-400)

Brand accent:       #6366F1  (indigo-500)   — interactive, links, active states
Brand hover:        #4F46E5  (indigo-600)

Status green:       #22C55E  (green-500)    — active instances, health OK
Status amber:       #F59E0B  (amber-500)    — warnings
Status red:         #EF4444  (red-500)      — cancelled, rejected, errors
Status gold:        #EAB308  (yellow-500)   — completed, end state
Status gray:        #A1A1AA  (zinc-400)     — deleted, pending, muted

Graph start:        border #22C55E, bg #F0FDF4
Graph end:          border #EAB308, bg #FEFCE8
Graph current:      border #6366F1, bg #EEF2FF  (Phase 5 live state)
Graph regular:      border #E4E4E7, bg #FFFFFF
Graph visited:      border #A1A1AA, bg #F4F4F5   (Phase 5)
```

### Dark Mode (class-based via Tailwind `dark:`)
```
Background:         #09090B  (zinc-950)
Surface:            #18181B  (zinc-900)
Border:             #27272A  (zinc-800)
Text primary:       #FAFAFA  (zinc-50)
Text secondary:     #A1A1AA  (zinc-400)
Text muted:         #71717A  (zinc-500)

(Status + brand colors same hue, same saturation — dark bg provides contrast)
```

---

## Typography

Font stack: `Inter, system-ui, -apple-system, sans-serif` (loaded via `@fontsource/inter`)

| Role | Class |
|------|-------|
| Page title | `text-2xl font-semibold tracking-tight` |
| Section heading | `text-lg font-medium` |
| Card title | `text-base font-medium` |
| Body | `text-sm text-zinc-700 dark:text-zinc-300` |
| Caption / meta | `text-xs text-zinc-500` |
| Code / IDs | `font-mono text-xs bg-zinc-100 dark:bg-zinc-800 px-1 py-0.5 rounded` |

---

## Spacing & Layout

- Base grid: 4px (Tailwind default)
- Page padding: `px-6 py-6` (desktop), `px-4 py-4` (mobile)
- Card padding: `p-4`
- Section gap: `gap-6`
- Sidebar width: 220px expanded, 52px collapsed

---

## Component Aesthetics

### Topbar
- Height: 56px
- Subtle bottom border (`border-b border-zinc-200 dark:border-zinc-800`)
- Logo: indigo dot + "Workflow Engine" text
- Health dot: 8px circle, green/red, pulsing animation when checking

### Sidebar
- Smooth slide animation (Framer Motion, 200ms ease)
- Active item: `bg-indigo-50 dark:bg-indigo-950 text-indigo-600 rounded-lg`
- Inactive: `text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800`
- Icons: lucide-react, 18px

### Table
- Header: `bg-zinc-50 dark:bg-zinc-900`, `text-xs font-medium uppercase text-zinc-500`
- Row hover: `hover:bg-zinc-50 dark:hover:bg-zinc-800/50`
- Sortable column: hover shows chevron icon
- Alternating rows: no (keep clean)

### Badge / Chip
```
Active:     bg-green-100  text-green-700   dark:bg-green-900/30 dark:text-green-400
Cancelled:  bg-red-100    text-red-700     dark:bg-red-900/30   dark:text-red-400
Completed:  bg-yellow-100 text-yellow-700  dark:bg-yellow-900/30 dark:text-yellow-400
Deleted:    bg-zinc-100   text-zinc-500    dark:bg-zinc-800      dark:text-zinc-400
Warning:    bg-amber-100  text-amber-700   dark:bg-amber-900/30  dark:text-amber-400
```
Shape: `rounded-full px-2.5 py-0.5 text-xs font-medium`

### Graph Nodes
- Border radius: 8px
- Shadow: `shadow-sm`
- Min width: 160px
- Task count badge: small indigo pill bottom-right of node
- Hover: scale 1.02, shadow-md (Framer Motion)
- Selected: ring-2 ring-indigo-500

### Graph Edges
- Stroke width: 2px
- Animated dot: 4px circle moves along path on hover
- Label: `text-xs font-mono bg-white dark:bg-zinc-900 px-1 border border-zinc-200 rounded`

### Buttons
```
Primary:   bg-indigo-600 text-white hover:bg-indigo-700
Secondary: border border-zinc-300 text-zinc-700 hover:bg-zinc-50
Danger:    bg-red-600 text-white hover:bg-red-700
Ghost:     text-zinc-600 hover:bg-zinc-100
```
All: `rounded-lg px-3 py-1.5 text-sm font-medium transition-colors`

### Form Inputs
- `border border-zinc-300 dark:border-zinc-700 rounded-lg px-3 py-2 text-sm`
- Focus: `ring-2 ring-indigo-500 ring-offset-1 outline-none`
- Error: `border-red-500`

### Loading Skeleton
- `animate-pulse bg-zinc-200 dark:bg-zinc-800 rounded`
- Match shape of content being loaded

### Empty State
- Centered in content area
- Icon (lucide, 48px, muted)
- Title `text-base font-medium text-zinc-700`
- Body `text-sm text-zinc-500`
- CTA button (when applicable)

### Warning Banner
```
bg-amber-50 dark:bg-amber-950/20
border border-amber-200 dark:border-amber-800
rounded-lg p-4
Icon: AlertTriangle (lucide, amber-500)
Title: text-sm font-medium text-amber-800
Items: ul text-sm text-amber-700
```

---

## Motion Grammar

Library: Framer Motion. All motion: `ease: [0.25, 0.1, 0.25, 1]` (ease-out cubic).

| Element | Animation |
|---------|-----------|
| Page enter | fade-in + slide-up 8px, duration 150ms |
| Sidebar expand/collapse | width tween, 200ms |
| Graph node mount | scale 0.9→1 + fade, 200ms |
| Graph node hover | scale 1→1.02, 150ms |
| Task card complete (Phase 5) | slide-left + fade, 250ms |
| State highlight change (Phase 5) | color tween, 400ms |

No motion when `prefers-reduced-motion: reduce`.

---

## Tailwind Config Additions

```ts
// tailwind.config.ts
export default {
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      colors: {
        brand: { DEFAULT: '#6366F1', hover: '#4F46E5' },
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4,0,0.6,1) infinite',
      },
    },
  },
}
```
