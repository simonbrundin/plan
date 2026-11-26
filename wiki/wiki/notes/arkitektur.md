---
title: Arkitektur
type: note
summary: TypeTrainers paketstruktur, beroenderiktning och dataflöde genom en session.
tags: [arkitektur, go, oversikt]
created: 2026-10-21
updated: 2026-11-26
status: active
related: [notes/bigram-algoritmen, notes/dokumentationssidor-i-granssnittet-docs, notes/session-livscykel, notes/utvecklingsflode]
---

TypeTrainer är en CLI-app (cobra) med ett bubbletea-TUI som tränar svensk
skrivning genom att prioritera svaga **bigram** (teckenpar).

## Paket

| Paket | Ansvar |
|---|---|
| `cmd/cli` | Cobra-kommandon. Bara orchestration - ingen affärslogik. |
| `cmd/wikictl` | Verktyg för att hantera [kunskaps-wikin](kunskaps-wiki.md). |
| `internal/session` | TUI: bubbletea-modell, tillstånd, rendering. |
| `internal/wordlist` | Ordlistan, bigram-index, inflectionsfamiljer, adaptiv urvalsalgoritm. |
| `internal/stats` | Persistenta sessioner, konfiguration, svaghetsanalys. |
| `internal/bigram` | Ren domänstatistik per bigram. Noll beroenden. |

## Beroenderiktning

```
cmd -> session -> {wordlist, stats, bigram} -> bigram
```

`wordlist` och `stats` importerar **inte** `session` - endast `session`
orchestrerar. `stats` importerar `wordlist` (för `WeakBigrams`), `wordlist`
importerar `bigram` (för `BigramsInWord`).

## Dataflöde i en session

```
main.go
  -> stats.LoadConfig()          ~/.config/typetrainer/config.json
  -> wordlist.New()              embed: internal/wordlist/words.txt
  -> stats.Load()                ~/.config/typetrainer/stats.json
  -> urval av ord                AdaptiveSelection / RandomWords / PracticeWords
  -> session.RunSession(...)     se session-livscykel
  -> s.Bigrams.MergeBigrams(...) -> s.Save() -> s.PrintStats()
```

## Viktigt att veta

- `internal/bigram` är helt fristående och lätt att testa - all
  svaghetslogik bor där.
- `stats` importerar `wordlist`, inte tvärtom. Det gör att domänstatistiken
  inte är beroende av ordlistan.
- `session` är det enda paketet som känner till bubbletea.

Se [session-livscykel](session-livscykel.md) för hur en session faktiskt
körs, och [bigram-algoritmen](bigram-algoritmen.md) för urvalslogiken.

## Relaterade

- [[notes/dokumentationssidor-i-granssnittet-docs]]
- [[notes/bigram-algoritmen]]
- [[notes/session-livscykel]]
- [[notes/utvecklingsflode]]
