---
title: cookies.json var versionsstyrd i offentligt neuro-repo
type: note
tags: [neuro, bug, säkerhet, projekt:simon]
created: 2026-11-03
updated: 2026-11-17
status: active
related: [notes/neon-api-token-hamnade-i-dotfiles-git-historiken, notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]
---

Bakgrund: neuro (github.com/simonbrundin/neuro) är ett offentligt repo. cookies.json var spårad av git och innehöll Clerk-sessionstoken (__session, __session_0eD4h1jX) för app.neuroquant.ai plus LinkedIn-spårningscookies. Autentiseringscookien hade inte kommit med i någon commit ännu, men filen skulle ha publicerat token nästa gång den committades.

Fix (2026-11-03): git rm --cached cookies.json + cookies.json i .gitignore. Historiken skrivs om med git filter-repo --path cookies.json --invert-paths (filter-repo tar bort origin - lägg till remote igen) och force-push. Gamla commits med filen är fortfarande nåbara via SHA tills GitHub GC:ar dem, men de innehåller bara _ga/_fbp/bcookie/li_gc/lidc.

Följd på pxe: git fetch + git reset --hard origin/main raderar cookies.json ur arbetsytan (det var spårat) - backupera den först. .env, state.json och runs.json är gitignorade och överlever.

Generell regel: sessionscookies och liknande ska alltid ligga i .gitignore i publika repor, och historiken måste rensas om filen redan committats.

## Relaterade

- [[notes/neon-api-token-hamnade-i-dotfiles-git-historiken]]
- [[notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]]
