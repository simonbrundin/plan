---
title: pappa: admin-token - ett stalle for token, inte handpallagning per miljo
type: note
tags: [pappa, hemligheter, drift, admin-token, projekt:pappa]
---

Felet "Satt NUXT_ADMIN_TOKEN i appens miljo for att kunna spara
kravprofiler" uppstod for att token bara kunde sattas via processmiljon
eller ett Kubernetes Secret. Losningen ar ett gemensamt stalle for
token, i samma anda som Neon-anslutningen:

- scripts/admin-token.sh genererar token i src/nuxt/.env (gitignorad,
  600), visar den och applicerar Secret pappa-admin i ett namespace.
- Servern slarar token ur NUXT_ADMIN_TOKEN och faller tillbaka pa
  src/nuxt/.env (server/utils/env.ts + server/utils/admin.ts). Det ar
  samma monster som DATABASE_URL redan anvande, sa det kunde inte
  ha blivit en ny sark - det var en generalisering av en beintlig.
- Dev, PR och prod delar hemlighetsnamnet pappa-admin, precis som
  pappa-neon-db. I klustret ligger dev och prod i samma namespace
  (repo-template), vilket gor detta till en icke-fraga - men det bor
  fixas, annars kan dev-secretet skriva over prod-secretet.

Varfor inte bara kora "kubectl create secret" i varje miljo: da far
granskaren mata in olika token i varje webblasare, och en miljo utan
token tyst stanger av skrivningen (503) utan att det syns i UI:t.

Kvar: PR-ns har ingen Secret och ingen podd. PR-token maste sattas nar
PR-miljon faktiskt anvands.
