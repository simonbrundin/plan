---
title: Learn: kunskapssystem implementerat i src/python
type: note
tags: [feature, learn, projekt:learn, python, session]
created: 2026-11-03
updated: 2026-11-17
status: active
related: [notes/learn-calibre-synk-och-automatisk-import-via-systemd, notes/learn-kunskapssystemet-portat-fran-python-till-go]
---

Repo 'learn' hade en README med visionen om ett e-pub-baserat kunskapssystem, men koden var enbart Go/Nuxt-template. Implementerade nu hela Python-delen i src/python (inget externt beroende: e-pub-parser via zip/OPF/NCX, heuristisk konceptextraktion, kunskapsgraf, quizmotor med SM-2, förklaringar, brainstorm, CLI via python -m learn.cli). 36 pytest-tester gröna.

**Status: superseded (2026-11-03).** Användaren ville inte ha Python i repot, så kunskapssystemet portades till Go – se [[notes/learn-kunskapssystemet-portat-fran-python-till-go]]. Ingen Python-kod finns kvar i 'learn'-repot.

Varför inga beroenden: Spacy/EbookLib/NetworkX i visionen är tunga och kräver modellnedladdning; heuristikerna fungerar utan dem och repoet kan köras överallt.

Viktig fälla under utveckling: rubriktext i e-pub-XHTML hamnade i samma 'mening' som första stycket, vilket gjorde att definitioner attribuerades till fel term (t.ex. 'Slow Thinking System 2'). Fix: heading-data skrivs inte till textbufferten i _TextExtractor.

Ännu ej gjort: ingen Go/Nuxt-koppling till kunskapsbasen, ingen AI-förklaring med lokal modell, ingen grafvisualisering.

## Uppdatering 2026-11-03

SUPERSEDED (2026-11-03): den här noten beskriver en tidigare implementation i Python. Användaren ville inte använda Python i repot, så kunskapssystemet portades till Go (se noten 'Learn: kunskapssystemet portat från Python till Go'). Ingen Python-kod finns kvar i repot.

Originaltext (för historiken): Repo 'learn' hade en README med visionen om ett e-pub-baserat kunskapssystem, men koden var enbart Go/Nuxt-template. Implementerade då hela Python-delen i src/python (inget externt beroende: e-pub-parser via zip/OPF/NCX, heuristisk konceptextraktion, kunskapsgraf, quizmotor med SM-2, förklaringar, brainstorm, CLI via python -m learn.cli). 36 pytest-tester gröna.

Varför inga beroenden: Spacy/EbookLib/NetworkX i visionen är tunga och kräver modellnedladdning; heuristikerna fungerade utan dem.

Viktig fälla under utveckling: rubriktext i e-pub-XHTML hamnade i samma 'mening' som första stycket, vilket gjorde att definitioner attribuerades till fel term (t.ex. 'Slow Thinking System 2'). Fix: heading-data skrivs inte till textbufferten.

Ännu ej gjort: ingen Go/Nuxt-koppling till kunskapsbasen, ingen AI-förklaring med lokal modell, ingen grafvisualisering.

## Relaterade

- [[notes/learn-calibre-synk-och-automatisk-import-via-systemd]]
- [[notes/learn-kunskapssystemet-portat-fran-python-till-go]]
