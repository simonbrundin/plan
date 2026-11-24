---
title: Utbyggnad av pappas personlista: bolagsregister + konferensbios gav 140 namngivna personer
type: note
tags: [pappa, sourcing, bolagsregister, gdpr, konferens, projekt:simon]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken, notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]
---

2026-11-08: pappas frölista gick från 88 till 140 namngivna personer.

VARFÖR: listan var stock med just de personer som själva publicerat sig (team-sidor, konferensprogram, bolagsposter). SCB Yrkesregistret visar 3 469 IT-arbetsplatser i Sundsvalls kommun, så 88 var ~2,5 % - och de 88 var de mest synliga, inte de mest värdefulla.

VAD SOM GJORDES:
- Offentligt bolagsregister (Bolagsverket via allabolag/alltombolag/bolagsfakta/rating.se) för ledningen i 16 mindre lokala IT-bolag. Registertiteln (VD, styrelseordförande, styrelseledamot) skiljs från en verifierad yrkestitel - styrelsenedamöter tas med som ledare/konsulter, men deras tekniska specialisation beskrivs inte som verifierad.
- Good Tech Conference 2025 och 2026: talarlistan gav 35 + 40 namn. Bios gav titel, arbetsgivare och tidigare roller. Ort härleddes bara när sambandet var källstyrkt - annars location_confidence: low.
- Anställningshistorik lades till för de GTC-profiler som hade belagbara datum.

RESULTAT: 20 av 36 målbolag har nu namngiven ledning (51 personer). 16 målbolag saknar fortfarande namngiven ledning.

LÄRDOM: bolagsregistret är den billigaste källan för 'vem bestämmer i det lilla bolaget' - det ger fler personer per timme än LinkedIn-sökning, och det är offentlig handling. Konferensbios ger både titel och tidigare roller, alltså anställningshistoria för personer som aldrig skriver sitt CV.

VIKTIGT FÖR GDPR: registeruppgifter är offentliga personuppgifter. Rättslig grund är legitimt intresse (rekrytering) och det krävs Art. 14-information till varje kontaktad person - se legal-sektionen i data/sundsvall-it-people.yaml.

## Relaterade

- [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]]
- [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]]
