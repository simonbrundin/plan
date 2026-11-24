---
title: pappa: uppdragsmatchning - delpoang forklarande ranking
type: note
tags: [pappa, rekrytering, ranking, databas, soring, projekt:pappa]
---

2026-11-08: Efter dubblettknappen blev nästa fråga hur kandidaterna ska rangordnas mot ett uppdrag. Byggt som (B) - databasfunktion + kravprofil + driftfrågor, INTE UI, så att vikterna kan prövas i SQL först.

VARFÖR INTE EN ENDA POÄNG: 124 av 140 it_experience_years-värden är gissade ur rollnivå (seniority), bara 12 är explicit och 4 har daterad anställningshistorik. En rankning som sorterar på det fältet sorterar i praktiken på titel, och en gissad träff kan slå en belagd. Därredan är varje poäng DELPOÄNG med förklaring, och konfidensen väger in som en egen del (60 % källdäckning + 40 % grund för erfarenhetsvärdet).

MODELL: job_profile (must_have, min_must_have, nice_to_have, min_seniority, min_experience_years, required_location, accepted_locations, required_languages, available_by, weights jsonb) + fn_match_candidates(p_profile uuid) som returnerar poäng 0-100, delpoäng per del, täckning och en förklaring per person. Vikterna ligger i profilen - ett uppdrag justeras utan kodändring.

VIKTER: must_have 40, nice_to_have 10, experience 15, seniority 10, location 10, language 5, availability 5, confidence 5. Delar som profilen inte styr är INAKTIVA och ingår inte i viktsumman - en profil utan språkkrav straffar inte folk för saknad språkdata. Saknad data ger neutrala 50 (språk, tillgänglighet), aldrig 0.

TVÅ LÄRDOMAR FRÅN MÄTNING PÅ DEN RIKTIGA DATABASEN (det viktigaste i noten):
1. MUST-HAVE SKA NORMALT VARA EN TAGG. "system_development + information_security + cloud" gav högst 1 av 3 täckt - ingen kandidat blev godkänd, eftersom taggarna är sällsynta och spridda. Lösning: min_must_have ("minst 2 av 3"). Diskvalifiering är därför mjukt: kandidaten hamnar UNDER de godkända med en förklaring, inte bort.
2. POÄNGEN ENSAM DISKRIMINERAR DÅLIGT. Nästan alla har ort rätt och senioritet över krav, så poängen hamnade ihop (94,4 mot 94,0 mellan en dokumenterad och en angiven). Listan sorterar därför först på hur väl kravet täcks (must-have-täckning, sedan nice-to-have), sedan på poängen. Det är viktigare än poängen i sig.

ÖVRIGA DESIGNVAL:
- Dokumenterad historik används i stället för angivet värde när den finns - konservativt, men en 3-årig dokumenterad histori kan underkänna en person som angett 12 år. Förklaringen visar därför båda ("dokumenterad erfarenhet 3 år - angivet totalt 12 år").
- Överkvalificering straffas inte i senioritetsdelen, men en nivå under kravet kostar 25 poäng.
- En profil utan krav returnerar inga rader - en tom profil är ett misstag, inte en matchning.
- Kompetenser matchas mot både person_competence (livslångs) och employment_competence (teknik per anställning).

TEST: db/test_match.sql (transaktion som rullas tillbaka) täcker eligible/diskvalifiering, att dokumenterad erfarenhet slår gissad, orten straffas, neutrala 50 för saknad språkdata, att vikterna påverkar poängen och att en tom profil inte matchar någon.

NÄSTA STEG: UI-vy med kravformulär och rankad lista (alternativ A), kundkonfliktkontroll mot kundens leverantörer, samt att fylla språk och available_from - de är tomma för alla 140 och kan därför inte ge någon signal ännu.

## Relaterade

- [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]]
- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]
- [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]]
