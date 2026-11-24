---
title: Learn: korsanalys av varje bok mot den samlade kunskapsbasen
type: note
tags: [learn, analys, kunskapsbas, projekt:learn]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/learn-kunskapsbasen-kan-lagras-i-postgresql]
---

Användaren bad om en "nätverkar" som plockar ut kunskap ur varje importerad bok och jämför den med den samlade kunskapsdatabasen: vad boken tillför, och vad som är okonventionellt.

Lösning i Go (repot /home/simon/repos/learn, modul github.com/simon/repo-template):
- Nypaketet internal/analysis jämför en bok med resten av biblioteket via store.Repository. Compare(bookID) ger en BookInsight, Library() ger en sammanfattning av hela basen.
- Fyra signaler: konceptnamn (samma term = samma kunskap), definitionslikhet (Jaccard over innehållsord), negationer ("inte", "never", "unlike" -> motsagelse) och nyckelord (nya ord + korskopplingar mellan bocker). Negationsmarkorerna ska vara entydiga negationer: mjuka formuleringar som "instead of" och "rather than" togs bort eftersom de bara betyder att ett alternativ valdes och skapade falska motsagelser.
- Traktlar: consistent >= 0.70, nuance 0.35-0.70, unconventional < 0.35, contradiction nar ena sidan har negation.
- CLI: learn compare <bok>. API: GET /books/{id}/insight och GET /insight. SPA: kortet "Analys mot kunskapsbasen" med flikarna Ny kunskap / Okonventionellt / Kopplingar, plus en summering av hela basen i sidokolumnen.

Den viktigaste insikten: analysens kvalitet bestams av konceptens kvalitet, inte av jmfrelsens kvalitet. Forsta körningen gav 60 avvikelser per bok varav de flesta var nonsens ("Java" mot "Effective Java", 0% likhet). Orsaken var extraktorn, som tog varje versal i meningens borjan som ett begrepp: Build, Comments, Writing, Given, Always.

Fyra fixar i internal/knowledge och internal/text, alla med tester:
1. text.GenericWords - en lista med 391 vanliga ord som aldrig ar koncept (given, build, system, value...). Ord- och frasformer ("Pro Tip", "Key Terms") filtreras via genericTerms.
2. text.LooksLikeProse (exporterad, anvands och av analysen) - en definition som ser ut som kod (@, {}, ::, =>) ar inte en definition. Detta tog bort de flesta falska motagelserna.
3. Kontraktioner ("I've", "Can't") och fotnotsiffror ("Ruby Capybara 17", "ActiveRecord1") filtreras bort i cleanTerm. Observera: ett tal som ar en del av termen (System 1) maste inte strippas - det krossar testet och System 1/2.
4. preferStrongConcepts - en term raknar bara som begrepp om den har en definition eller skrivs alltid med versal (aldrig med gemener). Svagare termer tas bara med om boken har fara tydliga begrepp.

Dessutom: text.LooksLikeIndex hoppar over register i e-puber (andelen ord med siffra > 25 procent). Utan det fick koncepten definitioner som var hela bokens index.

Viktiga lardomar:
1. Jmfrelse av goda beskrivningar ar trivialt; svartigheten ligger i att overhuvudtaget fa tag bra beskrivningar. Prioritera extraktorn, inte analysen.
2. Ord som skrivs med versal i borjan av en mening ar inte begrepp. Fallet ar det vanligaste felet i heuristisk NLP over blandad text.
3. Go marshallerar nil-slice som null, inte []. Ett API-svar med "novel": null tog ner SPA:t med "Cannot read properties of null (reading 'length')". Alla slice-fält i svaren maste initieras.
4. Ett test som anvande standarddatabasen lade pa sig SQLITE_BUSY medan en synk kora. Testet ska alltid peka pa en egen db via LEARN_DB.

Resultat pa maskinen: 82 bocker, 7 660 koncept, 5 368 unika begrepp, 998 begrepp som delas mellan bocker. En bok som "The Cucumber Field Guide" har nu 87 nya koncept (73 procent), 32 delade, 32 avvikelser, 0 motagelser och 8 korskopplingar.

Slutgiltiga lardomar fran sista varvet:
5. "instead of" och "rather than" ar INTE negationer. Nar de raknades som negationer fick varje bok 4-8 falska motagelser.
6. En korskoppling kraver minst två UNIKA delade nyckelord, inte två matchningar - annars fick en koppling med ett enda nyckelord styrka 1,0.
7. Diskursord ("Tyvarr", "Alternatively", "However") ar inte kunskap. De filtreras bade i extraktorn (text.GenericWords, nu 391 ord) och i analysen, sa att gamla importer inte behover omimportas.

Tekniska bocker som UNIX and Linux System Administration Handbook ger koncept som DNSSEC, DKIM, VxFS och Udev istallet for "Build" och "Consider".

## Relaterade

- [[notes/learn-kunskapsbasen-kan-lagras-i-postgresql]]

