---
title: Sessionens livscykel
type: note
summary: Sessionens tillstånd, tangentflöde, freeze och omstart i TypeTrainers TUI.
tags: [session, tui, bubbletea, tillstand]
created: 2026-10-21
updated: 2026-10-21
status: active
related: [notes/accuracy-fel-omstart, notes/arkitektur, notes/sprint-timer-saknas]
---

En session är en bubbletea-modell (`internal/session/model.go`) runt ett
`Session`-objekt (`internal/session/session.go`).

## Tillstånd

`Session` håller:

- `Buffer` - all text som ska skrivas, ord åtskilda av mellanslag, samt
  `Cursor` (position i runes).
- `Keystrokes` - varje tangenttryck med `Expected`, `Actual`, `TimeMs`,
  `Correct` och föregående förväntade tecken.
- `WordStartPos` - buffertposition där aktuellt ord börjar.
- `AttemptStart` - index i `Keystrokes` där nuvarande försök började.
- `HasSequenceErrors` - sant om aktuellt försök har fel.
- `Bigrams` - sessionens egen bigram-databas, som mergas in i globalen
  efteråt.
- `Summary` - fryst resultat, se nedan.

## Flödet

1. `RunSession` bygger ordlistan, skapar `Session` och `Model`, startar
   programmet.
2. Varje tangenttryck går genom `handleKeyPress`:
   - `Ctrl+C` avslutar.
   - **Om sekventensen har fel** accepteras bara omstartstangenter
     (Tab, Ctrl+U, Backspace, Ctrl+H, ctrl+backspace).
   - Annars hanteras utskrivning: markören flyttas framåt **bara vid rätt
     tecken**.
3. Rätt tecken -> bigram-statistik registreras, `WordStartTime` nollställs.
   När markören passerar ett mellanslag sätts `WordStartPos` till nästa ord.
4. När all text är skriven: `completeSession()` sätter `CompletedAt`,
   anropar `Freeze()` och stoppar timern.
5. Efter att programmet avslutats mergas sessionens bigrams in i den
   globala statistiken, sessionen sparas och statistiken skrivs ut.

## Två viktiga begrepp

**Freeze.** `Session.Freeze()` beräknar resultatet en gång och sparar det
i `Summary`. Utan det fortsätter WPM att räknas medan användaren läser
sammanfattningen, så siffrorna ändrade sig under tiden. `Result()` fryser
lazy vid första anropet.

**Omstart.** `restartSequence()` flyttar tillbaka markören till
`WordStartPos` och sätter `AttemptStart = len(Keystrokes)`. Tangenttrycken
**behålls** - det är hela poängen. Se
[accuracy-fel-omstart](accuracy-fel-omstart.md).

## Format

`runProgram` installerar en `tea.WithFilter` som skriver om `Ctrl+H` och
`Ctrl+W` till en syntetisk `restartKeyType` (0x1000). Det behövs eftersom
kanata-mappar om ctrl+backspace, och bubbletea levererar det som Ctrl+H.

Se [arkitektur](arkitektur.md) för paketstrukturen.

## Relaterade

- [[notes/accuracy-fel-omstart]]
- [[notes/arkitektur]]
- [[notes/sprint-timer-saknas]]
