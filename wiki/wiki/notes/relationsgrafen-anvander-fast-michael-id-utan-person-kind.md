---
title: Relationsgrafen använder fast Michael-id utan person_kind
type: note
tags: [pappa, relationer, datamodell, beslut, projekt:pappa]
created: 2026-11-17
updated: 2026-11-18
status: active
---

Beslutet om person_kind ändrades: det behövs inte. Michael Brundin är en vanlig person med statiskt id `00000000-0000-4000-8000-000000000001`. `relation_editor.person_id` pekar på detta id och `/relationer` filtrerar **Mina** genom att kontrollera om detta id är någon av relationens två parter. Kontakter är också vanliga person-rader och alla kopplingar är `person_relation`. Den tidigare person_kind-modellen ska inte användas.

Aktören är också fast: UI:t, klienten och servern använder alltid Michael Brundin (`SELF_ACTOR_NAME`/`SELF_PERSON_ID` i `src/nuxt/shared/utils/relations.ts`). Klienten kan inte välja eller skicka ett annat aktörsnamn - servern slår upp rollen ur `relation_editor` och faller tillbaka på det fasta person-id:t.

VARFÖR DET SPARAR TID: den vanliga felkällan var att filtret "Mina" byggdes på `owner_name`. 0 av 40 relationer hade `owner_name`, så vyn visade "0 ägda av dig" trots att det fanns relationer - felet såg ut att vara "inga data" när det i själva verket var "fel filter". Med part-identifieringen är filtret en fråga om grafen, inte om metadata.

En tillbaka-kastad import är också en risk: frölistan innehåller Michael som `soulmate-michael-brundin`, så en import kunde skapa en andra Michael med ett nytt id - och då hittade "Mina" ingenting igen. `db/import_from_yaml.py` uppdaterar därför min nod i stället för att skapa en ny, låser titeln ("Rekryterare" = min roll i appen, inte på jobbet) och håller mig borta från bevakningslistan.

Se också [[notes/michael-brundin-blir-en-personnod-i-relationsgrafen]].
