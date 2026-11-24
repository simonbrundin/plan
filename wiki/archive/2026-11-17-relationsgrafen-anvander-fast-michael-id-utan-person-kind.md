---
title: Relationsgrafen använder fast Michael-id utan person_kind
type: note
tags: [pappa, relationer, datamodell, beslut, projekt:pappa]
---

Beslutet om person_kind ändrades: det behövs inte. Michael Brundin är en vanlig person med statiskt id 00000000-0000-4000-8000-000000000001. relation_editor.person_id pekar på detta id och /relationer filtrerar Mina genom att kontrollera om detta id är någon av relationens två parter. Kontakter är också vanliga person-rader och alla kopplingar är person_relation. Den tidigare person_kind-modellen ska inte användas.
