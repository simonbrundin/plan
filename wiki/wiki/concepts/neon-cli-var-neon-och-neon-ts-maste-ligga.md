---
title: Neon CLI: var .neon och neon.ts måste ligga
type: concept
tags: [neon, cli, neon.ts, projekt:tmp]
created: 2026-11-02
updated: 2026-11-02
status: active
---

Neons CLI hittar projektkopplingen och policyfilen genom två separata regler:

1. KONTEXTFILEN .neon hittas genom att söka UPPÅT från cwd (som .git).
   Finns ingen .neon någonstans i kedjan används cwd.
2. POLICYFILEN neon.ts letas BREDVID .neon (dvs. i samma katalog). Finns den
   inte där faller sökningen tillbaka till cwd.
3. Utan .neon kan CLI:et inte välja projekt -> "ERROR: Multiple projects
   found, please provide one with the --project-id option".

FÖLJD: om du flyttar .neon + neon.ts till en underkatalog (t.ex. src/neon/),
slutar `neon deploy` fungera från repo-roten. Det finns två utvägar:
- Kör kommandot i underkatalogen (`cd src/neon && neon deploy`), eller
- Skicka --context-file src/neon/.neon (eller --config src/neon/neon.ts).

ÖVRIGA FLAGS SOM HJÄLPER:
- --config <sökväg>   Pekar ut en annan neon.ts.
- --context-file <s>  Pekar ut en annan .neon.
- --project-id / --branch  Väljer projekt/branch explicit.
- neon env pull --file <sökväg>  Skriver .env-filen dit du vill.
- neon deploy --no-env-pull    Hoppar över env-hämtningen.

VAR .env hamnar: "an existing .env, otherwise .env.local, in the project
directory (next to .neon, or cwd without one)". Med --context-file hamnar den
alltså bredvid context-filen, inte i cwd.

SE OCKSÅ: notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring
