---
title: Rekryteringspipeline: board per person, sourcing som attribut
type: note
tags: [rekrytering, pipeline, kanban, beslut, projekt:pappa]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Pappa har nu en kanban för rekryteringsflödet (/pipeline, commit 1285d2f).

Varför listan inte blev en kanban rakt av: steg 1 i listan (hitta kandidat, sökvägar + 'solla') är inte ett stadium. Sökvägarna är ett attribut per kandidat och 'solla' (yrke, stad, teknikkompetens) är de filter som redan fanns på /kandidater. Att göra sourcing till en kolumn hade gett 140 kort i kolumn 1.

Beslut som är värda att minnas:
- Boarden är per PERSON, inte per uppdrag (krav från Simon). En person kan ha flera roller, så en pipeline-rad per (person, roll) och boarden grupperar per person - en person ska aldrig syns som två kort.
- Kolumnen är den roll som kommit längst i flödet.
- Granskning kräver TVÅ röster (en anställd, en grundare); först när båda är gröna går kortet vidare till kontaktad. En röd röst = bortfall med orsak.
- Rörare identifieras via namnlista (pipeline_voter) - appen har ingen inloggning alls, så röstaren väljer sitt namn. Entra ID-inloggning är den naturliga uppgraderingen senare.
- Teams-korten ligger i en PLANNER-board (B1=c), inte i Teams-kanaler. Admin consent i tenanten är okänt (B2=c), så listan importeras från CSV med samma matchningsmotor (fn_match_person). Automatiska felmatchningar är oacceptabla - omatchade kort länkas manuellt.
- Ingen prognos av utfall (D) - bara verkliga antal per steg.

Lärdom: när någon säger 'kanban för X', fråga först om alla objekt faktiskt passerar samma steg. Här är listan en tratt (~5x gallring), inte parallell arbetskö - kanbans värde är WIP, och det finns inte i en sekventiell filtreringskedja.
