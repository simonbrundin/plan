---
title: pappa: anställningshistorik i CV-format + fix av kompetensordboken
type: note
tags: [databas, schema, sourcing, gdpr, projekt:pappa]
---

2026-11-03: Användaren föreslog en tabell med kandidaternas arbeten ("var, vad,
hur länge precis som på ett CV"). Genomfört.

BESLUT: ingen ny tabell. person_employment fanns sedan starten men var tom på
precis de delar som gör en CV-rad: 88 rader = 1 per person = bara nuvarande
tjänst, valid_from/valid_to 0/88 ifyllda, och ingen "med vad" per tjänst.
Tabellen byggdes ut i stället.

NYTT I SCHEMAT:
- person_employment: ref (idempotent nyckel mot YAML), employment_type
  (employment|freelance|consultant|project|education), seniority (rollnivå
  under perioden - person.seniority är den aktuella), confidence
  (high|medium|low), stack text[].
- employment_competence: teknik per anställning ("vilken stack på CGI?").
  Skiljer sig från person_competence som är livslångsaggregat.
- fn_documented_experience(uuid) -> antal månader ur daterad historik.
- v_cv: en rad per anställning med längd, stack, källa, confidence.
- person: available_from, spoken_languages; it_experience_basis fick värdet
  'documented'.

SEMANTIK (viktig att komma ihåg):
- valid_to IS NULL + is_current = false betyder "slutdatum okänt", INTE
  "pågående". is_current = true är den pågående tjänsten.
- Rader utan datum bidrar inte till årssumman - de är värdefulla ("var på
  Sogeti") men kan inte räknas.
- Dokumenterade månader räknas som UNION av intervall: en konsult kan ha
  haft parallella roller, en månad får inte räknas två gånger. Testade med
  2015-2018 + 2017-2019 + 2019-idag -> 140 månader (11,7 år), inte 13,8.

VARFÖR confidence per rad: en rad kan vara belagd av ett pressmeddelande
(high), utläst ur fritext (medium) eller vara en hypotes (low). Utan den
skiljningen blir gissningar data, och det bryter mot projektets egen regel
"utan källa -> inte i DB". Rader med low märks "osäker" i UI:t.

TVÅ BUGGAR SOM FIXADES SAMTIDIGT:
1. Kompetensordboken importerades fel: kategorinamn ("languages", "backend")
   lagrades som om de vore taggar, och de 104 riktiga taggarna hade ingen
   category alls. Nu: 133 taggar, 0 utan kategori, och UI:t grupperar
   filterchips per kategori. 31 använda taggar saknades i YAML-taxonomin ->
   tillagda. 'net'/'support'/'it' slogs ihop med sina dubbletter.
2. Importen lade aldrig till kopplingar bort. Nu är YAML-filen källan: taggar
   som tas bort från YAML tas bort i DB, och överflödiga taggar rensas.

MEDVETEN KONTROLL: kontaktloggning (write path mot contact_event) byggdes
INTE. Appen saknar autentisering, och en skriv-endpoint mot en offentlig
webbsida skulle vara öppen för manipulation av alla som hittar URL:en.
Lösningen kräver auth först (token eller idP) - det är ett eget beslut.

EFFECT: 101 anställningsrader för 88 personer, 22 kuraterade med källa,
3 personer med dokumenterad anställningstid (Tommy Gustafsson 34,8 år,
Thomas Nordin 17,8 år, Tobias Bergman 6,9 år), 0 rader utan källa, 0 dubbla
aktuella rader, 0 överlappande perioder.

Relaterat: [[notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund]]
[[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]
[[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]]
