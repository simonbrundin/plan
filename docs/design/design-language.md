# Agent System Design

> Design language och komponenter för Agent System UI.

**Senaste uppdaterad:** 2024-10-19

---

## Översikt

Agent System använder ett **minimalt dark theme** med en ren, professionell estetik. Designen balanserar teknisk tydlighet med visuell enkelhet — informativ utan att vara överväldigande.

**Inspiration**: Professionella developer tools (Linear, Raycast, Vercel Dashboard) kombinerat med en subtilt tech-estetik.

**Implementation**: Nuxt UI Pro med custom Tailwind CSS-konfiguration i `main.css`.

---

## Designprinciper

1. **Mörk som standard** — Dark theme med zinc-950 som bas
2. **Blå accent** — Blue-500 (#3b82f6) för primary actions och highlights
3. **Ren typografi** — Inter för UI, JetBrains Mono för kod
4. **Subtila borders** — Zinc-800 för cards och separators
5. **Luftig layout** — Massor av whitespace och tydlig hierarki

---

## Färgpalett

### Kärnfärger

| Token | Hex | Tailwind | Usage |
|-------|-----|----------|-------|
| `background` | `#09090b` | `zinc-950` | Primär bakgrund |
| `surface` | `#18181b` | `zinc-900` | Cards, elevated elements |
| `border` | `#27272a` | `zinc-800` | Borders, dividers |
| `divider` | `#3f3f46` | `zinc-700` | Subtle separators |

### Accentfärger

| Token | Hex | Usage |
|-------|-----|-------|
| `primary` | `#3b82f6` | Primary actions, links, active states |
| `primary-hover` | `#2563eb` | Hover states |

### Semantiska färger

| Token | Hex | Usage |
|-------|-----|-------|
| `success` | `#22c55e` | Success states, passed gates |
| `warning` | `#f59e0b` | Warnings, pending states |
| `error` | `#ef4444` | Errors, failed gates |
| `info` | `#3b82f6` | Info states (samma som primary) |

### Textfärger

| Token | Hex | Tailwind | Usage |
|-------|-----|----------|-------|
| `text-primary` | `#fafafa` | `zinc-50` | Primär text |
| `text-secondary` | `#a1a1aa` | `zinc-400` | Sekundär text |
| `text-muted` | `#71717a` | `zinc-500` | Muted text, labels |

---

## Typografi

### Fontfamiljer

```css
/* Sans - UI text */
font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;

/* Mono - Technical content */
font-family: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace;
```

### Typografisk skala

| Element | Size | Weight | Usage |
|---------|------|--------|-------|
| H1 | 2.25rem (36px) | 700 | Sida titlar |
| H2 | 1.5rem (24px) | 600 | Sektion rubriker |
| H3 | 1.125rem (18px) | 600 | Underrubriker |
| Body | 1rem (16px) | 400 | Brödtext |
| Small | 0.875rem (14px) | 400 | Labels, captions |
| Mono | 0.875rem (14px) | 400 | Kod, paths, status |

---

## Komponenter

### Cards

```vue
<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
  <!-- Content -->
</div>
```

**Egenskaper:**
- Background: `zinc-900`
- Border: 1px solid `zinc-800`
- Border-radius: `rounded-xl` (12px)
- Padding: `p-6` (24px)

**Hover state:**
```vue
<div class="hover:border-blue-500/30 transition-colors">
  <!-- Content -->
</div>
```

### Buttons

**Primary:**
```vue
<UButton color="primary">Action</UButton>
```

**Ghost/Outline:**
```vue
<UButton variant="outline">Secondary</UButton>
```

### Badges

```vue
<UBadge color="primary" variant="subtle">Status</UBadge>
```

### Navigation

Aktiva links i sidebar:
```vue
<!-- Blue highlight för active state -->
<div class="bg-blue-500/10 text-blue-400">
```

---

## Layout

### Page Structure

```
┌─────────────────────────────────────────────────┐
│ Sidebar (240px)    │ Main Content Area          │
│                    │                            │
│ ┌────────────────┐│ ┌────────────────────────┐ │
│ │ Logo/Title    ││ │ Page Header           │ │
│ ├────────────────┤│ ├────────────────────────┤ │
│ │ Navigation    ││ │                        │ │
│ │ • Home        ││ │ Content               │ │
│ │ • Goals       ││ │                        │ │
│ │ • Checkpoints ││ │                        │ │
│ │ • Conversations││ │                        │ │
│ ├────────────────┤│ │                        │ │
│ │ User Menu     ││ │                        │ │
│ └────────────────┘│ └────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

### Responsivt

| Breakpoint | Behavior |
|------------|----------|
| `md` (768px) | Sidebar kan döljas |
| `lg` (1024px) | Full layout |

---

## Motion & Transitions

### Transitions

```css
transition: all 150ms ease;
```

### Hover Effects

- Cards: border-color → blue-500/30
- Buttons: opacity → 90%
- Links: color → primary

### Ingen animation

Undvik dekoration animationer (bouncing, pulsing) förutom:
- Loading states (subtle pulse)
- Status indicators (gentle pulse för "active")

---

## CSS Implementation

### main.css

```css
@import "tailwindcss";
@import "@nuxt/ui";

.dark {
  --ui-bg: #09090b;
  --ui-bg-muted: #18181b;
  --ui-bg-elevated: #18181b;
  --ui-bg-accented: #27272a;
  --ui-text-dimmed: #71717a;
  --ui-text-muted: #a1a1aa;
  --ui-text-toned: #d4d4d8;
  --ui-text: #fafafa;
  --ui-text-highlighted: #fafafa;
  --ui-border: #27272a;
  --ui-border-muted: #27272a;
}
```

### app.config.ts

```ts
export default defineAppConfig({
  ui: {
    primary: 'blue',
    gray: 'zinc'
  }
})
```

---

## Checkpoint-Gate Visual System

### Workflow-färger

| Gate | Färg | Hex |
|------|-------|-----|
| Plan | Neutral | `zinc-400` |
| Execute | Primary | `#3b82f6` |
| Verify | Warning | `#f59e0b` |
| Ship | Success | `#22c55e` |

### Status Indicators

| Status | Visual |
|--------|--------|
| Running | Pulsing dot i accent-färg |
| Passed | Green checkmark |
| Failed | Red X |
| Pending | Gray/muted |

---

## Accessibility

### Kontrast

- Primär text på dark: minimum 4.5:1
- Muted text: minimum 3:1

### Focus States

```css
outline: 2px solid accent;
outline-offset: 2px;
```

---

## Filstruktur

```
src/nuxt/app/
├── app.config.ts          # UI config (primary: blue)
├── app.vue               # Root component
├── assets/css/
│   └── main.css          # Global styles + dark mode
├── components/
│   ├── SidebarToggle.vue
│   ├── UserMenu.vue
│   ├── NotificationsSlideover.vue
│   ├── goals/
│   │   ├── GoalTreeNode.vue
│   │   └── KanbanBoard.vue
│   └── messages/
│       ├── ChatInterface.vue
│       └── ConversationList.vue
├── layouts/
│   └── default.vue        # Main layout with sidebar
└── pages/
    ├── index.vue          # Home
    ├── goals.vue          # Goals tree
    ├── kanban.vue         # Kanban board
    ├── checkpoints.vue     # Workflow
    ├── conversations.vue   # Chat
    ├── docs/              # Documentation
    ├── readme.vue         # README
    ├── roadmap.vue        # Roadmap
    ├── repo-map.vue       # Repository map
    └── login.vue          # Login
```

---

## Changelog

### v2.0.0 (2024-10-19)

- Nytt minimalistiskt dark theme
- Blå accentfärg (#3b82f6)
- Ren kort-baserad layout
- Zinc-färgpalett för consistency
- Uppdaterade alla sidor och komponenter
