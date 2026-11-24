---
title: Michael Brundin blir en personnod i relationsgrafen
type: note
tags: [pappa, relationer, datamodell, projekt:pappa]
created: 2026-11-17
updated: 2026-11-18
status: active
---

Beslut: ersätt den separata my_contact/my_contact_relation-modellen med vanliga person- och person_relation-rader. Michael Brundin är en vanlig person-rad (`external_id = 'self-michael-brundin'`) med statiskt id `00000000-0000-4000-8000-000000000001`. Ingen särskild "person_kind" behövs för att skilja hon från kandidater.

VARFÖR DET SPARAR TID: en graf med två sorters noder (kandidater + nätverkskontakter) tvingar varje sökväg, vy och import att ta ställning till "vilken sort är detta?". Med en enda sort är "Michael känner X" exakt samma relation som "A känner B", och ett fast id räcker för att hitta hans egna relationer.

`db/migrate_self_node.sql` (idempotent, körs efter `db/schema.sql` mot den DIREKTA Neon-anslutningen) gör följande:

1. skapar min nod om den saknas,
2. **viktar in** varje annan "Michael Brundin"-rad i min nod - relationer, källor, kompetenser, anställningar, bevakningslista, pipeline och kontaktlogg flyttas, och den gamla raden tas bort först när allt är flyttat. Dubbletter i relationerna slås ihop (den starkaste styrkan vinner) och paren canonicaliseras om - min uuid är den lägsta, så relationen hamnar på rätt sida,
3. gör om `my_contact`/`my_contact_relation` till personer och vanliga relationer, och tar sedan bort tabellerna,
4. knytar `relation_editor.person_id` till min nod,
5. äger mina relationer, så att "Mina" och "Gemensamma" inte överlappar.

Kördes 2026-11-18 mot Neon (production) efter backup: 172 personer, 43 relationer (40 + 3 kontakter), exakt en Michael-rad, och relationen Michael ↔ Mattias Lindahl bevarad. Verifierat med Nuxt typecheck, 135 tester och Nuxt build.

Se också [[notes/relationsgrafen-anvander-fast-michael-id-utan-person-kind]].
