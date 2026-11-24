---
title: Leader-key gjordes global (mellanslag + sekvenser på alla sidor)
type: note
tags: [feature, nuxt, keyboard, leader, projekt:plan]
---

Leader-läget (mellanslag) fanns tidigare bara på målsidan och delades i två
implementeringar: useGlobalNavigation (slog ner tangenttryck globalt) och en
egen modal i pages/goal/[id].vue. Det gjorde att sekvenser bara fungerade på
målsidan, och att en tangentserie kunde trigga både leader-kommandot och
sidans vanliga vim-kommandon.

Ändring (2026-11):
- Ny composable src/nuxt/app/composables/useLeaderKey.ts äger leader-läget.
  Den lyssnar på window i CAPTURE-fasen så att den hinner före sidornas egna
  tangentbevakare oavsett monteringsordning (sidans onMounted körs före
  layoutens). Konsumeraade händelser stoppas med stopPropagation och märks
  i en WeakSet (isLeaderConsumed).
- Globala kommandon: g+i (inkorg), g+g (mål), g+p (prioritet), g+c
  (kalender), g+s (inställningar). Kommandolistan visas i LeaderKeyModal,
  som nu renderas i layouts/default.vue och därmed syns på alla sidor.
- Sid-specifika kommandon registreras med registerPageCommands() - målsidan
  registrerar f+d (toggle avklarade) och i (byt ikon).
- Sekvensmatchningen ligger i app/utils/leaderKeys.ts (matchLeaderKey) så att
  den är testbar utan Nuxt-kontext; test/unit/leaderKeys.test.ts täcker den.
- useGlobalNavigation:s leader-logik togs bort (den blockerade bland annat
  mellanslag i textfält - se noten "mellanslag blockerades i målbeskrivningen").
- useSearchState delar nu sin isSearchOpen-ref mellan anropare (tidigare
  skapades en ny ref per anrop, så sökpaletten bara syntes för layouten).

Verifierat: bun test test/unit (145 pass), bun run build, samt Playwright mot
den byggda appen - mellanslag öppnar modalen på /, /priority, /calendar och
/goal/:id, g+p navigerar till /priority, Escape stänger, och vanliga tangenter
(d) på målsidan öppnar inte leader-läget.
