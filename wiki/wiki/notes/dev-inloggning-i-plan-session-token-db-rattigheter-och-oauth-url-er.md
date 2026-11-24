---
title: Dev-inloggning i plan: session-token, DB-rättigheter och OAuth-URL:er
type: note
tags: [bug, auth, dev, zitadel, session, projekt:plan]
created: 2026-11-10
updated: 2026-11-10
status: active
---

Fix av dev-inloggningen i plan-repot (kind-klustret, ns plan).

Grundorsak till 401 trots korrekt cookie: signSession skrev created_at/expires_at
som Unix-integer, medan Session-structen saknade JSON-taggar. Vid avkodning
fick ExpiresAt nolltiden -> "session expired" -> API:t svarade 401.
Fix: json-taggar + RFC3339-timestamps + test (sessions_test.go).

Övriga dev-brister som hittades:
- DATABASE_URL pekade pa "postgres" (fanns inte) -> ingen DB-anslutning.
  Nu FQDN mot CNPG-tjänsten plan-db-rw.
- Approllen "plan" saknade rättigheter (tabellerna ägdes av admin).
  Nu äger approllen tabellerna; Atlas applicerar schemat som approllen.
- AtlasSchema var stalled: urlFrom pekade pa plan-db-user utan FQDN, och
  GRANT-satser i schema.yaml kraschade Atlas ("atlas login" krävs).
- Nuxt-deploymenten saknade NUXT_PUBLIC_GO_API_URL -> /auth/login pekade
  pa localhost:8080. Env-override heter NUXT_PUBLIC_DEV_AUTH_ENABLED
  (camelCase -> SNAKE_CASE), inte ..._DEV_AUTH.
- /auth/login dubbelade /api/v1 (goApiUrl innehaller redan prefixen) -
  det var även trasigt i prod.
- Dev-portarna är nu fasta (nuxt 3100, go 4100, pgadmin 4200) så att
  URL:er och OAuth-redirects är stabila mellan omstarter.

Övrigt: Zitadel (auth.simonbrundin.com) var onåbar från laptop/kluster under
arbetet, så dev-inloggningen körs via dev-login (DEV_AUTH_ENABLED=true ->
POST /api/v1/auth/dev-login) utan Zitadel. En fast redirect-URI
(http://localhost:4100/api/v1/auth/callback) måste registreras i Zitadel
innan riktig Zitadel-inloggning kan testas i dev.

FALLGROP: Nuxt 4 läser serverDir från rootDir/server, inte src/app/server.
Filerna i src/nuxt/app/server/ är död kod - de första rutter som gäller
ligger i src/nuxt/server/.
