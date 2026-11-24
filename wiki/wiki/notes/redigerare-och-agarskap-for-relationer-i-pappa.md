---
title: Redigerare och ägarskap för relationer i pappa
type: note
tags: [pappa, relationer, behörighet, auth, modell, ui, projekt:pappa]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater, notes/min-kontaktlista-jag-som-nod-i-grafen-pappa, notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin, notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer]
---

2026-11-10: Frågan var 'jag vill se vilka som kan redigera mina relationer och deras styrka'. Två läsningar: (a) vem ÄNDRADE relationen, (b) vem FÅR ändra den. Lösningen täcker båda.

MODELL: appen har ingen användarinloggning (samma som rörare/grundare i pipen), så identiteten är ett VALT NAMN ur en namnlista - relation_editor (name, role 'editor'|'lead', active). Klienten skickar bara namnet i 'actor'; servern slår upp rollen ur listan, så klienten kan inte skicka sig till 'lead', och okänt namn avvisas med 403. Namnet är ett påstående, inte ett bevis - det är samma tilltro som rörarvalet, och det står uttryckligen i specen.

REGLER: relation med ägare -> bara ägaren eller en chef får ändra styrka/ta bort; annars 403 med förklaringen 'Relationen ägs by Anna' (så vet man vem man ska fråga). Relation UTAN ägare (t.ex. importerad ur YAML) är gemensam - alla får röra den. Skillnaden är 'min' mot 'vår'.

FALLGROP: relationen är symmetrisk och lagras EN gång per par. Ägandet ligger alltså på PARET, inte på riktningen - 'A:s relation till B' och 'B:s relation till A' har samma ägare. Det är konsekvent (relationen är ett påstående om paret) men innebär att man inte kan ha olika ägare per sida.

FALLGROP 2: ägaren sätts bara vid INSERT (on conflict do update nollställer inte owner_name), så en ändring av styrka byter inte ägare - bara updated_by. Det är meningen: ägandet flyttar inte för att någon rättar en styrka.

UI: 'Redigerar som'-väljaren ligger i relationsdialogen (där all skrivning sker), listan över redigerare syns i samma dialog och admins kan lägga till/stänga av där. Relationerna visar 'ägs av X' och 'ändrad av Y' både i dialogen och i detaljvyn.

VAD SOM INTE GÖRS: inga riktiga konton, ingen sessionsbindning av namnet. Det skulle kräva en inloggningsmodell och är en egen uppgift.

## Relaterade

- [[notes/min-kontaktlista-jag-som-nod-i-grafen-pappa]]
- [[notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer]]
- [[notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin]]
- [[notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater]]
