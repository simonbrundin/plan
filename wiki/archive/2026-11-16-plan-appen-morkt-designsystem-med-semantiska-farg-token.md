---
title: Plan-appen: mörkt designsystem med semantiska färg-token
type: note
tags: [plan, ui, css, tailwind, dark-mode, projekt:repos]
---

Omarbetad frontend-styling i src/nuxt (2026-11-16). All funktioner/logik oförändrad - bara presentation.

Varför: sidorna hade hårda grå/blå-klasser (bg-gray-800, text-blue-400) blandat med ljusläge-lämningar (bg-white, text-gray-900), vilket gav ett rakt och osammanhängande utseende och risk för oläsbarhet om temat byttes.

Vad gjordes:
- Semantiska token i @theme (src/nuxt/app/assets/css/main.css): canvas/surface/surface-2/elevated, line/line-strong, ink/ink-muted/ink-dim, accent* (indigo). Alla templates använder nu utilities (bg-surface, text-ink, border-line ...) i stället för grå/blå-nyanser.
- Mörkt som standard: colorMode.preference = 'dark' (följer inte systemtemats ljusläge) + Nuxt UI-färgprimary=indigo, neutral=slate.
- Uppdaterade: layout (sidofält/panel/mobilrad), inbox, mållistan, målsidan, prioriteringslista, inloggning, signup, modaler.

Viktig fälla (hittad genom mätning i webbläsaren, inte genom läsning): egna CSS-klasser i @layer components fungerar INTE i den här appen. Nuxt UI:s komponent-CSS (@layer components) laddas före huvud-CSS:en, vilket gör att lagret registreras först och därmed får LÄGST prioritet - en .surface-card { border: 1px solid ... } förloras mot Tailwinds preflight (border: 0 solid i @layer base). Använd utilities eller olayoutade regler, aldrig @layer components.

Verifying: byggd med bun run build och kontrollerat i headless Chromium (computed styles + skärmdumpar) mot en mock-API - inte bara byggt utan tittat.
