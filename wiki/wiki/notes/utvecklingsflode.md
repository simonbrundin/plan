---
title: Utvecklingsflöde
type: note
summary: Bygg, test och installera av TypeTrainer och wikictl, plus verktyg som inte finns.
tags: [build, test, verktyg]
created: 2026-10-21
updated: 2026-10-24
status: active
related: [notes/arbetsstationen-ar-omarchy-arch-apt-get-finns-inte-anvand-pacman-yay, notes/arkitektur, notes/konventioner, summaries/llm-wiki-karpathy]
---

Allt Go-kod ligger i `src/go` (modul `typetrainer`). Kör alltid kommandona
därifrån.

## Bygga

```bash
cd src/go
go build -o typetrainer ./cmd/cli
```

**Faller:** `go build -o typetrainer .` misslyckas med "no Go files in
src/go" - det finns ingen `main.go` i modulroten. Entrypointen ligger i
`cmd/cli/`. (SPEC.md har fortfarande det gamla kommandot.)

## Installera

```bash
cd src/go
go build -o ~/.local/bin/typetrainer ./cmd/cli
```

Den installerade binären ligger i `~/.local/bin/typetrainer`.

## Testa

```bash
cd src/go
go test ./...          # hela sviten, ~1 sekund
go test ./internal/... # bara biblioteken
go test -run TestModel ./internal/session/
```

Testerna är rena enhetstester med stdlib `testing` - inga mocks, inga
nätverk, inga fixtures. `session_test.go` driver bubbletea-modellen
direkt via `m.handleKeyPress(...)` utan att starta ett TUI.

## Verktyget för wikin

```bash
cd src
go run ./agent/cmd/wikictl list
go run ./agent/cmd/wikictl search "bigram"
go run ./agent/cmd/wikictl ingest --dry-run
go run ./agent/cmd/wikictl ingest
```

Se [kunskaps-wiki](kunskaps-wiki.md).

## Saker som inte finns

- Ingen linter-konfiguration (golangci-lint), ingen CI, ingen Makefile.
- `gofmt` är inte konsekvent tillämpad i repot - `internal/bigram/bigram.go`,
  `internal/stats/config.go` och `internal/stats/session_test.go` ligger
  kvar oformaterade sedan den första commiten. Formatera det du rör, men
  gör inte en stor omformatering av filer du inte ändrar.
- Inga externa beroenden utöver cobra, bubbletea och lipgloss.

Se [konventioner](konventioner.md).

## Relaterade

- [[notes/arbetsstationen-ar-omarchy-arch-apt-get-finns-inte-anvand-pacman-yay]]
- [[notes/arkitektur]]
- [[notes/konventioner]]
- [[summaries/llm-wiki-karpathy]]
