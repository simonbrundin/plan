---
title: pappa: hemlighet i build-context - .env i Docker-imagen
type: note
tags: [pappa, hemligheter, docker, sakerhet, projekt:pappa]
---

När admin-token lades i src/nuxt/.env (gitignorad) uppstod ett problem
som inte syntes i UI:t: Dockerfile har "COPY . ." utan .dockerignore, sa
.env-filen med NUXT_ADMIN_TOKEN kopierades in i byggsteget och hamnade i
byggcachen pa maskinen. Runtime-imagen var ren (den kopierar bara
.output), sa problemet var osynligt - man maste granska byggsteget
direkt for att se det.

Lardom: en hemlighet i arbetsytan ar inte en hemlighet sa lange
byggkontexten ar okontrollerad. .dockerignore ar en del av
hemlighetshanteringen, inte bara en byggoptimering. Regressionstestet
blev: bygg --target build och kontrollera att /app/.env inte finns.

Den andra farligheten: att en hemlighet bara finns pa ratt stalle ar
inte nog - servern maste hitta den oavsett vilken katalog appen startas
fran. Sökningen ga nu uppat i katalogtradet i stallet for att vara
bunden till process.cwd().

Verifierat mot den korande podden (Tilt-bygd): läsning 200, skrivning
utan token 401, med token 200.
