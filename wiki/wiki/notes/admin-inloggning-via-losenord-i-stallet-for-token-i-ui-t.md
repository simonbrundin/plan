---
title: Admin-inloggning via lösenord i stället för token i UI:t
type: note
tags: [pappa, admin, auth, projekt:pappa]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Pappa (src/nuxt) har bytt från 'mata in admin-token i webbläsaren' till en enkel inloggningssida på /admin.

Varför: tokenfält i UI:t var både jobbigt och riskabelt (token i sessionStorage, i URL-fragment, delad mellan sidor). Användaren ville aldrig se eller mata in tokenen.

Hur det fungerar nu:
- /admin har ett formulär med ett lösenordsfält. Lösenordet skickas i POST-kroppen till /api/admin/login och byts mot en httpOnly-cookie 'pappa_admin_session' (sju dagar).
- Cookien är 'utgångsstämpel.HMAC', HMAC-nyckeln är NUXT_ADMIN_TOKEN - samma värde används som adminlösenord, så det finns bara en hemlighet.
- requireAdmin godkänner antingen cookien eller headern x-admin-token (för CLI/script).
- Tokenfält, statusrutor och instruktioner är borttagna från kandidater, uppdrag, pipeline, dubblettpanelen och sidomenyn. 401 ger nu 'Logga in på /admin.'
- scripts/admin-token.sh login öppnar nu /admin utan tokenfragment; 'show' finns kvar för CLI.

Observera: logout raderar cookien i webbläsaren men ogiltigförklarar inte det signerade värdet - en session är stateless.
