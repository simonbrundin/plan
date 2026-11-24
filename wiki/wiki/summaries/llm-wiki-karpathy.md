---
title: Sammanfattning: LLM Wiki (Karpathy)
type: summary
summary: Karpathys LLM-wiki-mönster - en persistent, sammanhängande kunskapsgraf som LLMs underhåller, i tre lager.
tags: [wiki, meta, kunskapsbas, llm, karpathy]
sources: [raw/2026-10-24-llm-wiki-karpathys-monster-for-llm-underhallna-kunskapsbaser]
created: 2026-10-24
updated: 2026-11-23
related: [notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet, notes/kunskaps-wiki, notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin]
---

# Sammanfattning: LLM Wiki (Karpathy)

## Källa

Råkälla: `raw/2026-10-24-llm-wiki-karpathys-monster-for-llm-underhallna-kunskapsbaser.md`
Origin: https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f

## Kärna

RAG hämta chunks ur rådokument vid frågetillfället. Varje fråga börjar om
från noll, och ingenting ackumuleras.

LLM-wiki byter ut det mot en **persistent, sammanhängande kunskapsgraf**
som LLM:en bygger upp och underhåller. Kunskapen kompileras en gång och
hålls sedan aktuell, i stället för att härledas om vid varje fråga.

> "the wiki is a persistent, compounding artifact"

Skillnaden i praktiken: korsreferenser finns redan, motsägelser är redan
flaggade, syntesen reflekterar allt man läst.

## Tre lager

| Lager | Innehåll | Ägare |
|-------|----------|-------|
| **Raw sources** | Kurerade källdokument. Immutable - LLM:en läser men ändrar aldrig | Människan |
| **The wiki** | Markdown-sidor: summaries, entities, concepts, comparisons, index, synthesis | LLM:en |
| **The schema** | Dokumentet (AGENTS.md) som definierar struktur, konventioner och workflows | Människa + LLM |

## Operationer

**Ingest** - Ny källa i raw-samlingen. LLM:en läser den, diskuterar
huvudpunkter, skriver en summary, uppdaterar index, uppdaterar berörda
entity- och concept-sidor och appendar en loggrad. En källa kan röra
10-15 wikisidor.

**Query** - Sök i wikin, läs sidorna, syntetisera svar med citat. Bra svar
**filas tillbaka** som nya sidor, så att utforskningar också ackumuleras.

**Lint** - Hälsokoll: motsägelser mellan sidor, föråldrade påståenden som
nyare källor har ersatt, ödelappar utan inkommande länkar, viktiga koncept
utan egen sida, saknade korsreferenser, dataluckor som kan fyllas med
webbsök. Linten ska också föreslå nya frågor och nya källor.

## Index och logg

- **index.md** - innehållsorienterad katalog över allt i wikin, organiserad
  per kategori, med länk + enradig sammanfattning per sida. LLM:en läser
  den först vid fråga. Fungerar bra till ~hundratals sidor utan
  embedding-RAG.
- **log.md** - kronologisk, append-only. Med konsekvent prefix
  (`## [YYYY-MM-DD] operation | Titel`) blir loggen parsebar med enkla
  unix-verktyg: `grep "^## \[" log.md | tail -5`.

## Verktyg

Karpathy rekommenderar `qmd` (hybrid BM25/vector + LLM reranking, on-device,
CLI + MCP) för sök när wikin växer. Ärligt att säga: indexfilen räcker
länge.

## Varför det fungerar

Det jobbiga är inte läsandet eller tänkandet - det är bokföringen:
korsreferenser, aktuella summaries, motsägelser, konsistens över tiotals
sidor. Människor ger upp wikis när underhållsbördan växer snabbare än
värdet. LLMs tröttnar inte, glömmer inte att uppdatera en korsreferens,
och kan röra 15 filer i en svep.

Människans jobb: kurera källor, rikt analysen, ställa bra frågor, tänka
på vad det betyder. LLM:ens jobb: allt annat.

Historisk anknytning: Vannevar Bush "Memex" (1945) - en privat, aktivt
kuraterad kunskapslagring där kopplingarna mellan dokumenten är lika
värdefulla som dokumenten själva. Bushs olösta problem var vem som
skulle underhålla den. Det gör LLM:en.

## Tillämpning här

Den här wikin är en konkret instans av mönstret, byggd för agenter i
flera repor:

- `raw/` - immutabla källor, idag bl.a. Karpathys gist.
- `wiki/` - genererade sidor i `entities/`, `concepts/`, `summaries/`,
  `notes/`.
- `AGENTS.md` (denna repo: `wiki/AGENTS.md`) - schemat, dvs. den tredje
  lagret.
- `wikictl` - CLI:t som gör sök, skrivande, indexering och lint
  reproducerbart av alla agenter.

Skillnaden mot stockmodellen: wikin är global och delas mellan repor,
och `notes/` fångar den situationsbundda kunskap som inte hör hemma på
en konceptsida.

## Kopplingar

- [[notes/kunskaps-wiki]]
- [[notes/utvecklingsflode]]

## Relaterade

- [[notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet]]
- [[notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin]]
