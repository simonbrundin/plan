---
title: pappa: dubblettknapp - mänsklig granskning, delvis matchning och beständig sammanslagning
type: note
tags: [pappa, databas, gdpr, verktyg, granskning, heuristik, projekt:simon]
created: 2026-11-03
updated: 2026-11-17
status: active
related: [notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen, notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking, notes/personer-kan-laggas-in-manuellt-i-pappa]
---

2026-11-08: Användaren bad om en knapp som försöker slå ihop dubbletter. Byggd som knapp + API + databasfunktion.

VARFÖR INTE AUTO-MERGE: dagens data innehåller tre Söderberg på Arkaden Konsult (Gert, Per, Robert) och två Mikael Nordlund-liknande par. En automatisk sammanslagning på namnlikhet skulle utplåna släktingar. Därför är kandidaterna förslag, och en människa väljer vilken post som ska behållas.

2026-11-08: Användaren rapporterade att knappen tog "Mikael Gidlund" och "Mikael Nordlund" som samma person. ORSAKEN: helnamnstrigramheuristiken similarity(normalized_name, normalized_name) >= 0.45 mäter likheten över hela namnet, och ordvis jämförelse (word_similarity) gör "gidlund"/"nordlund" till 0.50 eftersom de delar suffixen "lund". LÖSNINGEN: namnet delas upp i förnamn och efternamn och varje del jämförs separat.

VARJE NAMNDEL JÄMFÖRS SEPARAT (fn_part_match / fn_name_parts / fn_name_match):
- exact (1.00): identiskt.
- prefix (0.90): den ena delen är prefix av den andra ("lind"/"lindqvist"). Kräver minst 4 tecken i den kortare delen, annars skulle "per" matcha "petra".
- near (0.75): 1-3 tecken fel ("anders"/"andreas", "jonatan"/"jonathan"). Kräver samma första bokstav, annars matchar "gidlund" och "nordlund" (levenshtein 3, gemensam suffix "lund").
- none (0.00).

TVÅ NIVÅER AV KANDIDATER:
- STARK (kan slås ihop): samma e-post, samma telefon, identiskt namn, eller BÅDE förnamn och efternamn stämmer.
- SVAG (visas i egen utfällbar sektion, kan inte slås ihop): bara en namndel stämmer. "Mikael Gidlund"/"Mikael Nordlund" (gemensamt förnamn), "Gert"/"Robert Söderberg" (gemensamt efternamn).

RESULTAT: starka kandidater gick från 10 (alla falska) till 1 ("Mikael Nordlander"/"Mikael Nordlund" - båda delarna stämmer, near på efternamnet). Svaga par: 146.

TVÅ KNAPPAR, enligt önskemål:
- "Slå ihop - behåll den här": man väljer själv post. Utan val behålls den rikare posten (flest källor och anställningsrader), deterministiskt via it_experience_years och external_id.
- "Behåll båda": registreras i tabellen duplicate_dismissal, som utesluter paret ur listan så att det inte behöver granskas igen.

LÄRDOM: en trigramlikhet över ett helt namn är en dålig personindikator - den mäter suffix- och prefixlikhet mellan efternamn. Namn ska alltid jämföras strukturellt (förnamn mot förnamn, efternamn mot efternamn), aldrig som en hel sträng. Ett gemensamt förnamn är en MYCKET svagare signal än ett gemensamt efternamn, och båda är svaga jämfört med e-post och telefon.

BUGGAR SOM FIXADES UNDER VÄGEN:
- (a.email is not null and a.email = b.email) ger NULL, inte false, när a har e-post men b inte. Det gjorde att weak blev NULL och alla "gemensamt förnamn"-par klassades som starka. Måste vara (a.email is not null and b.email is not null and a.email = b.email).
- Ett CTE som heter "p" kan inte refereras till sig själv i en korrelerad subquery - "p.id" i FROM-satsen pekar på CTE:n, inte på tabellaliasen. Lösning: aliasen "person p" + CTE-namnet "pair_posts".
- "drop" är ett reserverat ord och kan inte användas som tabellalias.

TVÅ DESIGNVAL SOM SPELAR ROLL:
1. Skrivning kräver admin-token (NUXT_ADMIN_TOKEN, hemligheten pappa-admin i klustret). Läsning är publik - det är samma data som /api/people visar. En publik merge-knapp hade varit en knapp som raderar personer ur masterdatabasen.
2. Sammanslagningen registreras i tabellen person_merge. Utan det vore sammanslagningen OBESTÄNDIG: frölistan (YAML) innehåller fortfarande båda posterna, så nästa import skulle skapa tillbaka den raderade dubbletten. Importer hoppar nu över external_id som finns i person_merge.

SEMANTIK I fn_merge_people: kopplingar flyttas (källor, kompetenser, anställningar, bevakningslista, kontaktlogg), tomma fält fylls från den raderade posten men befintliga värden skrivs aldrig över, och högst en aktuell anställningsrad per person (den kuraterade raden med ref vinner över legacy-raden).

TEST: db/test_merge.sql kör sammanslagningen i en transaktion som rullas tillbaka och fäller med undantag om kopplingar inte flyttas, om två aktuella rader finns kvar, om person_merge saknas eller om spåret i audit_log saknas. Nu även negativa matchningstest (ett gemensamt ord ska inte räcka), positiva test (identiskt namn, samma e-post, samma telefon, prefix) och test av att "behåll båda" utesluter paret.

## Relaterade

- [[notes/personer-kan-laggas-in-manuellt-i-pappa]]
- [[notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking]]
- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]
