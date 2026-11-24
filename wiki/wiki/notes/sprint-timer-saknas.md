---
title: Sprint-läget saknar timer
type: note
summary: Sprint-läget saknar timer, så tidsåtgången går inte att mäta per session.
tags: [bug, sprint, offen, tui]
created: 2026-10-21
updated: 2026-10-21
status: open
related: [notes/session-livscykel]
---

`typetrainer sprint 30` visar en nedräkning som aldrig räknas ned.

## Symptom

`timeLeft` står kvar på startvärdet (`cfg.SprintDuration`) hela sessionen.
`Update` hanterar `*time.Timer` och decrementerar `timeLeft`, men **någon
ställer aldrig timern**.

## Orsak

I `internal/session/model.go`:

- `Model.timer` finns som fält (`timer *time.Timer`).
- `createModel` i `runner.go` sätter `timeLeft: cfg.SprintDuration` men
  skapar ingen timer.
- `Update` saknar fallback: det finns ingen `tea.Tick` och ingen
  `time.AfterFunc` som skickar ett meddelande.

Greppar man efter `timer|Tick|timeLeft` i `internal/session/` får man
exakt fyra träffar, varav ingen startar en timer.

## Konsekvens

Sprint-läget slutar aldrig på tid. Sessionen tar slut först när hela
`TextBuffer` skrivits, och den genereras som `RandomWords(duration * 3)`
ord - alltså styrs sprinten av ordlistans storlek, inte av klockan.

## Trolig fix

Antingen en `tea.Tick` som `Update` hanterar:

```go
func (m *Model) Init() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}
```

eller en `time.AfterFunc` som sätter `m.timer` och skickar `tea.KeyMsg` /
`tea.Quit`. Förstahandsvalet är `tea.Tick`, eftersom bubbletea då styr
livscykeln och timern inte behöver stoppas manuellt.

## Att verifiera

Kör `typetrainer sprint 10` och kontrollera att nedräkningen rör sig och
att sessionen avslutas efter 10 sekunder. Lägg till ett test som
`Update` hanterar tick-meddelandet korrekt.

Se [session-livscykel](session-livscykel.md).

## Relaterade

- [[notes/session-livscykel]]
