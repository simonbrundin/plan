# Design System

> **Senaste uppdaterad:** 2024-10-19

Design language för Agent System UI.

---

## Översikt

Agent System använder ett **minimalt dark theme** med en ren, professionell estetik. Designen balanserar teknisk tydlighet med visuell enkelhet.

**Implementation**: Nuxt UI Pro med custom Tailwind CSS-konfiguration.

---

## Färgpalett

### Kärnfärger

| Token        | Hex       | Usage                        |
| ------------ | --------- | ---------------------------- |
| `background` | `#09090b` | Primär bakgrund (zinc-950)   |
| `surface`    | `#18181b` | Cards, elevated (zinc-900)   |
| `border`     | `#27272a` | Borders (zinc-800)           |
| `divider`    | `#3f3f46` | Subtle separators (zinc-700) |

### Accentfärger

| Token     | Hex       | Usage                  |
| --------- | --------- | ---------------------- |
| `primary` | `#3b82f6` | Primary actions, links |
| `success` | `#22c55e` | Success states         |
| `warning` | `#f59e0b` | Warnings               |
| `error`   | `#ef4444` | Errors                 |

### Textfärger

| Token          | Hex       | Usage         |
| -------------- | --------- | ------------- |
| `text-primary` | `#fafafa` | Primär text   |
| `text-muted`   | `#a1a1aa` | Sekundär text |

---

## Implementation

### main.css

```css
.dark {
  --ui-bg: #09090b;
  --ui-bg-muted: #18181b;
  --ui-bg-elevated: #18181b;
  --ui-bg-accented: #27272a;
  --ui-border: #27272a;
}
```

### app.config.ts

```ts
export default defineAppConfig({
  ui: {
    primary: "blue",
    gray: "zinc",
  },
});
```

---

## Komponenter

### Card

```vue
<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
  <!-- Content -->
</div>
```

### Hover

```vue
<div class="hover:border-blue-500/30 transition-colors">
  <!-- Content -->
</div>
```

---

## Relaterat

- [Arkitektur](/docs/architecture)
- [Workflow](/docs/workflow)
- [Design Tools](/docs/design-tools)
