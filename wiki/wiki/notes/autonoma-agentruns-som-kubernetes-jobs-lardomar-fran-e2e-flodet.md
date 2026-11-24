---
title: Autonoma AgentRuns som Kubernetes Jobs: lärdomar från E2E-flödet
type: note
tags: [plan, agent, kubernetes, job, e2e, bug, projekt:simon]
created: 2026-11-24
updated: 2026-11-24
status: active
---

Verifierat 2026-11-24 i ~/repos/plan med mål #540: Plan-mål -> köad AgentRun -> Kubernetes Job -> förslag -> human gate -> isolerad workspace -> verifiering -> publicerad branch agent/goal-540. Huvudgrenen orörd.

Lärdomar som kostade tid:

1. Jobbens image får inte vara en statisk tagg. Downward API kan inte läsa spec.containers[0].image - lösningen var att läsa agentens egen image ur podens spec via Kubernetes-API:t (POD_NAME + podens containers[0].image). Annars kör jobbet en gammal tagg än den som är utbyggd, och felet är förbrytande.

2. Två Tilt-instanser mot samma kind-kluster (agents + plan) fightar om deploymenterna. Symptom: en kubectl set image som strax återställs, och en gammal image i klustret. Kör bara en Tilt per kluster.

3. Modellen skriver ofta ett testkommando som matchar en fras i sitt eget innehåll - men frasen är delad över en radbrytning ("mänsklig\ngranskning"). Förlaget har därför alltid en självverifiering: kör testkommandot mot det föreslagna innehållet innan human gate. Låt modellen först skriva om testet (innehållet orört) och sedan hela förslaget.

4. git-flaggor måste ligga före subkommandot: git -c http.extraheader=... push origin branch, inte git push -c ....

5. GIT_AUTH_HEADER ska vara en komplett header ("Authorization: Basic <base64(x-access-token:TOKEN)>"). Bara base64-paren räcker inte - git frågar då efter användarnamn.

6. Agentens commits behöver en git-identitet. Sätt den lokalt i arbetsytan (git config user.email/name), aldrig globalt i poden.

7. När arbetsytan klonas lokalt från källkatalogen måste origin återställas till källans remote - annars når pushen aldrig det riktiga repot.

8. Ett köat AgentRun för ett redan definierat mål måste hanteras som en uttrycklig begäran - annars gör workern ingenting, eftersom det räcker med två körningar för att den automatiska kodfasen ska hoppas över.

9. Ett mål får ha högst ett pågående jobb - två jobb mot samma repo skriver över varandra.
