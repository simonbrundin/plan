---
title: Konventioner i repot
type: note
summary: Språk, struktur, test- och dokumentationskonventioner för agents-repot.
tags: [konventioner, stil, go]
created: 2026-10-21
updated: 2026-10-21
status: active
related: [notes/arkitektur, notes/utvecklingsflode]
---

## Språk

- **Svenska** i UI-strängar, kommentarer och commit-meddelanden.
- **Engelska** för identifierare, typer och funktionsnamn.
- Kommentarerna förklarar **varför**, inte vad. Koden säger vad.

## Struktur

- `internal/` per ansvar, inte per lager. `cmd/` innehåller bara
  orchestration.
- Små, namngivna typer med en uppgift var. `Session`, `Keystroke`,
  `WordResult`, `TextBuffer` - alla i `internal/session`.
- Privata hjälpfunktioner ligger i samma fil som sin användare
  (`renderBox`, `padText`, `endingOf`, `sameFamily`).
- Magiska tal och strängar är konstanter (`DisplayLength`,
  `ProgressBarWidth`, `boxWidth`, `familyPrefixLen`, `maxPerFamily`).

## Test

- `*_test.go` bredvid koden, i samma paket (inga `xxx_test`-paket).
- Tabellbaserade tester där det passar.
- Testnamnen på engelska, testkommentar och felmeddelanden på svenska.

## Beteende

- Cobra-kommandon: `runAdaptive`, `runSprint` - korta funktioner i
  `main.go` som laddar config, väljer ord och startar sessionen.
- I/O-fel hanteras alltid. `stats.Load()` faller tillbaka på default
  istället för att krascha, men `Save()` skriver utan att kontrollera
  felet - det är en känd lucka.
- `encoding/json` för persistens, `os.UserHomeDir()` + `filepath` för
  kataloger.

## Dokumentation

- `SPEC.md` är designdokumentet: datamodell, algoritm, konfigurationsschema
  och feature-lista. Det är **delvis inaktuellt** - se
  [utvecklingsflöde](utvecklingsflode.md) för avvikelser.
- `README.md` är i praktiken tom - där finns bara projektnamnet.

## Ändringsdisciplin

Ändra bara det som behövs. Rör inte angränsande kod, kommentarer eller
formatering som du inte har med i uppgiften. Faller det en rad utanför
begäran ska den bort - annars går det inte att läsa diffen.

## Relaterade

- [[notes/arkitektur]]
- [[notes/utvecklingsflode]]
