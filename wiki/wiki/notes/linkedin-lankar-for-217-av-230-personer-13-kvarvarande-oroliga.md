---
title: LinkedIn-länkar för 217 av 230 personer - 13 kvarvarande oroliga
type: note
tags: [linkedin, pappa, sourcing, manuell-verifiering, projekt:pappa]
created: 2026-11-19
updated: 2026-11-19
status: active
---

Slutligt resultat av LinkedIn-kopplingen i pappa: 217 av 230 personer i frölistan har nu en verifierad /in/-länk, databasen har 218 av 231 (231 = 230 + Mattias Thorsell, som lagts till manuellt).

Varför det blev 217 och inte 230: 13 personer har bekräftad organisation och roll, men ingen entydig LinkedIn-profil har kunnat beläggas. De ligger nu dokumenterade i data/linkedin-unresolved.yaml med organisation, beviskälla och de kandidater som rensats bort - så nästa körning inte ska göra om arbetet, och en lucka ska vara medveten.

Viktigaste lärdomar (varför de här 13 inte gick att lösa):
- Namnkollision är det vanligaste problemet, inte svagt matchande. Kristin Åberg finns i tre versioner (Sundsvalls kommun, Ånge kommun, frölistans Mittuniversitetet) - utan entydig organisationskoppling är ingen av dem godkänd.
- Organisationskataloger bekräftar personen men anger inte slugen. Arkadens styrelselista, SDK:s kontaktlista, GDM:s ledningssida och Soulmates /api/site har alla LinkedIn-länkar saknade eller tomma fält.
- Kvadrats egna konsultsidor har JSON-LD med worksFor och knowsAbout men bara för vissa personer ett sameAs-fält (Sofie Collini, Mikael Thuné). För Mikael Falck, Mikael Hansson och Niclas Hellberg finns ingen sameAs - och sidans /mikael-falck är inte en LinkedIn-URL, bara Kvadrats eget slug. Det är en närmiss som nästan ser rätt ut.
- GitHub kan vara det bästa identitetsspåret: github.com/mickevi heter Mikael Viklund och anger "Soulmate Sundsvall AB" + Timrå, vilket matchar Soulmates egen bio ("bor numera i Timrå"). Men GitHub-profilen har ingen LinkedIn-länk, så spåret räcker inte för att ange en slug.
- Den loggade LinkedIn-sessionen får inte användas för massläsning (bot-spärr). Allt är gjort via Exa-sökning + /contents, plus rå HTML och JSON-API:er.

Metod som fungerade för att få nya träffar: organisationens egen webbplats, ibland bara i rå HTML (Sogetis direktlänk till henrik-carlsson-0045091), plus personens egna LinkedIn-inlägg som bevis när profilens headline är tunn. Ett inlägg kan bekräfta en roll vid tidpunkten för inlägget, inte nödvändigtvis nu.

Process (reproducerbart): db/find_linkedin_profiles.py -> db/rescore_linkedin_profiles.py -> db/apply_linkedin_slugs.py, med undantag i data/linkedin-approved.yaml (35 poster, var och en med reason) som skyddas från borttagning och vinner över sökresultaten.

Kontroll efter applicering: inga dubbletter av sluggar, inga avvikelser YAML->DB, 13 personer utan sluggar i båda (samma 13), extra->>'linkedin_urn' ifylld för alla 218.
