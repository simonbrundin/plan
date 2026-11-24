---
title: Learn: kunskapssystem implementerat i src/python
type: note
tags: [learn, python, feature, session, projekt:learn]
---

Repo 'learn' hade en README med visionen om ett e-pub-baserat kunskapssystem, men koden var enbart Go/Nuxt-template. Implementerade nu hela Python-delen i src/python (inget externt beroende: e-pub-parser via zip/OPF/NCX, heuristisk konceptextraktion, kunskapsgraf, quizmotor med SM-2, förklaringar, brainstorm, CLI via python -m learn.cli). 36 pytest-tester gröna.

Varför inga beroenden: Spacy/EbookLib/NetworkX i visionen är tunga och kräver modellnedladdning; heuristikerna fungerar utan dem och repoet kan köras överallt.

Viktig fälla under utveckling: rubriktext i e-pub-XHTML hamnade i samma 'mening' som första stycket, vilket gjorde att definitioner attribuerades till fel term (t.ex. 'Slow Thinking System 2'). Fix: heading-data skrivs inte till textbufferten i _TextExtractor.

Ännu ej gjort: ingen Go/Nuxt-koppling till kunskapsbasen, ingen AI-förklaring med lokal modell, ingen grafvisualisering.
