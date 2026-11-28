---
title: Rökprov efter rollout och två race-fynd i previewflödet
type: note
tags: [ci, preview, smoke-test, bug, projekt:plan]
---

Efter att Rust blevit huvudspåret saknades en automatisk verifiering av varje rollout. Nu finns den.

**Rökprovet.** API:t och agenten rapporterar vilken build som kör (PLAN_BUILD_SHA, satt via --build-arg i Dockerfilen). /.github/scripts/smoke-test.sh väntar in sig på rätt sha och kör sedan kontroller mot den publika URL:en. Två varianter: preview (root, Nuxt /api/health, API /health, dev-login, databasberoende /api/v1/goals, agentens endpoint) och prod (root, hälsokontroller, 401 utan session). Resultatet kommenteras på PR:et.

**Varför /health och inte /api/health.** /api/health tillhör Nuxt och säger ingenting om builden. API:ts /health exponeras nu via routern (både prod och preview). Ett första försök läste /api/health och rapporterade felaktigt att byggversionen aldrig kom ut - trots att API:t svarade rätt.

**Fynd 1: previewen kunde skapas efter att PR:en stängts.** En byggning tar minuter. Om PR:en mergades under byggtiden körde städscriptet, och den pågående körningen publicerade sedan overlayen igen - namespace och overlay hängde kvar. Jobbet frågar nu GitHub om PR:en fortfarande är öppen och hoppar över publiceringen om den stängts.

**Fynd 2: GitHub kunde inte tolka workflowen.** En step-output hette 'stängd' - variabelnamn med å/ä accepteras inte. Symtomet var en röd körning som bara sa 'workflow file issue' utan jobb alls. Samma fel uppstod på bot-PR:er, vilket gjorde det svårt att se verkliga fel.

**En grov lärdom om byggcachar.** Jag misstodde först ett cacheproblem för orsaken till ett byggfel. Felet var att Dockerfile:n refererade till en crate som inte fanns i branchen - jag hade kopierat filen från en annan arbetskopia. Scopen på gha-cachen ändrades tillbaka: per ref gav kalla byggen utan att lösa problemet.
