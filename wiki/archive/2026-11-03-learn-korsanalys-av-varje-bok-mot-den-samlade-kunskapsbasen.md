---
title: Learn: korsanalys av varje bok mot den samlade kunskapsbasen
type: note
tags: [learn, analys, kunskapsbas, projekt:learn]
---

Användaren bad om en "nätverkar" som plockar ut kunskap ur varje importerad bok och jämför den med den samlade kunskapsdatabasen: vad boken tillför, och vad som är okonventionellt.

Lösning i Go (repot /home/simon/repos/learn, modul github.com/simon/repo-template):
- Nypaketet internal/analysis jämför en bok med resten av biblioteket via store.Repository. Compare(bookID) ger en BookInsight, Library() ger en sammanfattning av hela basen.
- Fyra signaler: konceptnamn (samma term = samma kunskap), definitionslikhet (Jaccard over innehållsord), negationer ("inte", "never", "unlike" -> motsagelse) och nyckelord (nya ord + korskopplingar mellan bocker).
- Traktlar: consistent >= 0.70, nuance 0.35-0.70, unconventional < 0.35, contradiction nar ena sidan har negation.
- CLI: learn compare <bok>. API: GET /books/{id}/insight och GET /insight. SPA: kortet "Analys mot kunskapsbasen" med flikarna Ny kunskap / Okonventionellt / Kopplingar, plus en summering av hela basen i sidokolumnen.

Den viktigaste insikten: analysens kvalitet bestams av konceptens kvalitet, inte av jmfrelsens kvalitet. Forsta körningen gav 60 avvikelser per bok varav de flesta var nonsens ("Java" mot "Effective Java", 0% likhet). Orsaken var extraktorn, som tog varje versal i meningens borjan som ett begrepp: Build, Comments, Writing, Given, Always.

Fyra fixar i internal/knowledge och internal/text, alla med tester:
1. text.GenericWords - en lista med ~290 vanliga ord som aldrig ar koncept (given, build, system, value...). Ord- och frasformer ("Pro Tip", "Key Terms") filtreras via genericTerms.
2. text.looksLikeProse - en definition som ser ut som kod (@, {}, ::, =>) ar inte en definition. Detta tog bort de flesta falska motagelserna.
3. Kontraktioner ("I've", "Can't") och fotnotsiffror ("Ruby Capybara 17", "ActiveRecord1") filtreras bort i cleanTerm. Observera: ett tal som ar en del av termen (System 1) maste inte strippas - det krossar testet och System 1/2.
4. preferStrongConcepts - en term raknar bara som begrepp om den har en definition eller skrivs alltid med versal (aldrig med gemener). Svagare termer tas bara med om boken har fara tydliga begrepp.

Dessutom: text.LooksLikeIndex hoppar over register i e-puber (andelen ord med siffra > 25 procent). Utan det fick koncepten definitioner som var hela bokens index.

Viktiga lardomar:
1. Jmfrelse av goda beskrivningar ar trivialt; svartigheten ligger i att overhuvudtaget fa tag bra beskrivningar. Prioritera extraktorn, inte analysen.
2. Ord som skrivs med versal i borjan av en mening ar inte begrepp. Fallet ar det vanligaste felet i heuristisk NLP over blandad text.
3. Go marshallerar nil-slice som null, inte []. Ett API-svar med "novel": null tog ner SPA:t med "Cannot read properties of null (reading 'length')". Alla slice-fält i svaren maste initieras.
4. Ett test som anvande standarddatabasen lade pa sig SQLITE_BUSY medan en synk kora. Testet ska alltid peka pa en egen db via LEARN_DB.

Resultat pa maskinen: 82 boker, ~7700 koncept, ~5300 unika begrepp, ~1000 begrepp som delas mellan bocker. En bok som "The Cucumber Field Guide" har nu 85 nya koncept (71 procent), 35 delade, 35 avvikelser varav 8 motagelser. Tekniska bocker som UNIX and Linux System Administration Handbook ger koncept som DNSSEC, DKIM, VxFS och Udev istallet for "Build" och "Consider".
