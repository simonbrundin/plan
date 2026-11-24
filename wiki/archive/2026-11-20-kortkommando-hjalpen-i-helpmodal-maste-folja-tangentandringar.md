---
title: Kortkommando-hjälpen i HelpModal måste följa tangentändringar
type: note
tags: [frontend, kortkommandon, helpmodal, projekt:plan]
---

Regel: varje nytt eller ändrat kortkommando i plan-appen ska uppdatera src/nuxt/app/components/HelpModal.vue i samma PR. Hjälpmodalen är enda platsen där användaren ser tangenterna, så en tangent som inte finns där är osynlig.

Var tangenterna ligger: globala i app/layouts/default.vue, leader-läget i app/composables/useLeaderKey.ts (globala + sidkommandon i pages/goal/[id].vue, pages/goals/index.vue, pages/priority.vue), målsidan i pages/goal/[id].vue + composables/useGlobalNavigation.ts, /goals i pages/goals/index.vue, /priority i pages/priority.vue + composables/usePriorityMode.ts, kalendern i components/VerticalCalendar.vue, bekräftelsedialogen i components/ConfirmDeleteGoalModal.vue.

Regeln ligger nu i AGENTS.md i plan-repot (avsnittet 'Kortkommandon - uppdatera ALLTID hjälpmodalen'). Hjälpmodalen är data-driven (sektioner och tangentlistor i script-blocket) så nya tangenter läggs in i listorna, inte i markupen.
