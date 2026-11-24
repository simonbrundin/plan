---
title: IT-arbetsplatser i Sundsvall: 3 469 enligt SCB Yrkesregistret
type: note
tags: [sourcing, it, sundsvall, statistik, projekt:pappa]
---

Mått på IT-marknadens storlek i Sundsvall, gjord 2026-10-27 via SCB:s PxWeb-API
(tabell AM0208D/YREG58BAS, "Anställda med arbetsplats i regionen (dagbefolkning)
efter region, yrke (3-siffrig SSYK 2012) och kön", baserat på BAS/Yrkesregistret).

RESULTAT 2024:
- Sundsvalls kommun: 3 469 anställda i IT-yrken
  (189 IT-chefer, 2 576 IT-arkitekter/systemutvecklare/testledare, 704 drift-/support-/nätverkstekniker)
- Västernorrlands län: 4 674
- Riket: 248 417
- Sundsvall utgör 74 % av läns IT-arbetsplatser men 41 % av befolkningen.
  3,5 % av Sundsvalls befolkning ar anstalld i IT-yrke mot 2,3 % i riket och 1,9 % i lanet.
- Sektorfördelning i Sundsvall: privat näringsliv 2 112, staten 1 245, kommun 88, region 9.
  (Staten är stor i Sundsvall: Bolagsverket, Försäkringskassan, Trafikverket, Länsstyrelsen m.fl.)
- Näringsgren: 45 % av IT-jobben i Sundsvall ligger i "information och kommunikation"
  (IT-/telekomföretag), 33 % i offentlig förvaltning, 8 % i finans/försäkring, 4 % i tillverkning.
- Trend: Sundsvall 3 109 (2020) -> 3 469 (2024), +11,6 %.
- Fyrasiffrig yrkesstatistik finns bara på länsnivå: Västernorrland har 1 540
  "Mjukvaru- och systemutvecklare m.fl." (riket 110 698), 513 systemanalytiker/IT-arkitekter,
  176 systemtestare, 57 IT-säkerhetsspecialister.

VARFÖR DETTA SPELAR ROLL:
Den gamla grova uppskattningen ("1 000-1 500 IT-personer i Västernorrland") var
fel - den byggde pa ett urval foeretagslistor och missade intern IT i stat, industri
och finans. Den riktiga siffran ar mer an dubbelt sa stor. Det gora framgangen i
sourcing: poolen ar stor nog for att kallkontakt kan ge resultat, men marknaden
ar fortfarande liten nog att man maste vara rak med antalet kontakter.
Sundsvall har dessutom oproportionerligt manga IT-arbetsplatser - det ar regionens
IT-hubb, inte bara en kommun bland andra.

METODISK LÄRDOM:
Använd SCB:s Yrkesregister (PxWeb) i stället for bolagsregister eller
Google-sökning nar marknadsstorlek ska uppskattas. Aggar varje anstalld enligt
SIN (ett) yrke, sa "IT-yrkena" ar en exakt avgransad grupp, inte en gissning.
Dagbefolkning (arbetsplatsens ort) ar rätt mättstall for rekrytering, inte
nattbefolkning (bostadsorten).

ARTIFAKT: /home/simon/repos/pappa/data/scb-it-statistik-sundsvall.json (rådata med
tabell-ID och hämtningsdatum). Frölistan /home/simon/repos/pappa/data/sundsvall-it-people.yaml
är uppdaterad med den mätta populationen (88 namngivna = ~2,5 % av 3 469).
