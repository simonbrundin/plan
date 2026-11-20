# Startdatum på mål

## Problem

Ett mål ska ibland inte finnas i prioriteringslistan ännu. Ett mål som
"börjar den 1 mars" eller "gäller från att jag flyttar" ska inte dra
uppmärksamhet i /priority förrän det är dags.

Det finns redan tre datum på ett mål, men ingen av dem gör detta:

- `started` - när arbetet faktiskt började (tidsstämpel).
- `due` - när målet ska vara klart (deadline).
- `recurrence_anchor` - när en återkommande serie startade.

Inget av dem säger "det här målet får inte prioriteras förrän dess".

Att dölja målet i klienten (t.ex. med en flagga i localStorage) fungerar
inte: listan laddas från servern, och samma filter måste gälla för
agenten och för API-anrop.

## Lösning

Mål får ett **startdatum** (`goals.start_date`, ett kalenderdatum
"YYYY-MM-DD"). Det anger från vilket datum målet tidigast får dyka upp i
/priority.

- Saknas startdatum (null) får målet synas direkt - som tidigare.
- Ett startdatum i framtiden gör att målet inte tas med i prioriteringslistan.
- Ett startdatum i det förflutna (eller idag) släpper igenom målet.

Filtreringen sker i API:t (`GetPrioritizedGoals`), inte i klienten. Därför
gäller regeln överallt - i listan, för agenten och för andra API-anrop.

### API:t

`PATCH /api/goals/:id` med `{"start_date": "2027-01-01"}` sätter datumet,
`{"start_date": null}` rensar det. Fältet skickas alltid ut i
målssvaret (även när det är null), precis som `due`.

På `POST /api/goals` heter målets eget startdatum också `start_date`.
Seriens startdatum för ett återkommande mål heter däremot numera
`recurrence_start_date` (tidigare `start_date` i create/update) - det är
seriens anchor, inte målets synlighetsdatum.

### Gränssnitt

Tangenten **B** sätter startdatum på det markerade målet:

- på målsidan (/goal/:id) - knappen ligger bredvid deadline-knappen,
- i /goals-listan,
- i /priority-listan.

Datumet visas bredvid knappen. Ett startdatum i framtiden visas i blått,
vilket signalerar att målet ännu inte får prioriteras. Handlingen går att
ångra med **Z**.

## Acceptance Criteria

- [x] Ett mål kan få ett startdatum på målsidan, i /goals och i /priority.
- [x] Ett mål med framtida startdatum dyker inte upp i /priority.
- [x] Ett mål med startdatum i det förflutna (eller idag) dyker upp i /priority.
- [x] Startdatumet kan rensas.
- [x] Startdatumet går att ångra med Z.
- [x] API:t validerar att datumet har formatet "YYYY-MM-DD".
