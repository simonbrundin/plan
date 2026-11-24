---
title: Michael Brundin blir en personnod i relationsgrafen
type: note
tags: [pappa, relationer, datamodell, projekt:pappa]
---

Beslut: ersätt den separata my_contact/my_contact_relation-modellen med vanliga person- och person_relation-rader. person_kind skiljer candidate, network och self. Michael Brundin skapas idempotent som self-nod. Gamla kontaktdata migreras av db/migrate_self_node.sql: kontakter utan kandidatmatch blir network-personer och både jag-kontakter samt introduktioner blir vanliga symmetriska person_relation-rader. /relationer visar hela grafen och Mina filtrerar på redigerarens person_id. Verifierat med Nuxt typecheck, Bun-test (135 pass) och Nuxt build.
