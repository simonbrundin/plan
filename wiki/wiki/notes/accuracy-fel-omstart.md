---
title: Accuracy-felet vid omstart
type: note
summary: Så här upptäcktes att omstart behöver skiljas från skrivfel, och varför sekvensfelet inte gav en tydlig signal.
tags: [bug, postmortem, accuracy, session]
created: 2026-10-21
updated: 2026-11-17
status: active
related: [notes/median-i-stallet-for-snitt-for-wpm, notes/session-livscykel]
---

## Symptom

Accuracy visade 100 % trots att användaren gjort upprepade fel i ett ord.
"Totalt fel" visade rätt antal, men accuracy var alltid 100.

## Orsak

`restartSequence()` gjorde:

```go
m.session.Keystrokes = nil
```

Den nollställde **hela** tangenttryckslistan varje gång användaren tryckte
Backspace/Tab för att börja om ordet. Felaktiga tangenttryck raderades
alltså bort ur historiken, och accuracy räknades bara på de tangenttryck
som återstod - de var ju alla korrekta.

Bigrams som hörde till det omstartade ordet försvann på samma sätt, vilket
gjorde att den statistik som ska driva träningen inte registrerades.

## Fix

Tangenttrycken behålls nu, och en ny gräns markerar varje försök:

```go
m.session.AttemptStart = len(m.session.Keystrokes)
```

- `getPreviousRune()` returnerar bara föregående tecken **inom** det
  aktuella försöket (`len(Keystrokes) <= AttemptStart` ger 0), så att
  bigram-kontexten inte läcker mellan försök.
- Accuracy och raw WPM räknas på hela `Keystrokes`, som nu innehåller
  felen.

## Bär på

Att "nollställa" ett felsamt spelläge genom att kasta historik är nästan
alltid fel. Historik ska separeras från tillstånd, inte raderas. Här fanns
två olika saker i samma fält: *aktuellt läge* och *historia*. Lösningen
var att dela upp dem (`AttemptStart`).

## Regressionstest

`src/go/internal/session/session_test.go`:
`TestModel_ErrorIsCountedAfterRestart` och
`TestModel_RestartKeepsBigramContext`.

Se [session-livscykel](session-livscykel.md) för hur omstart fungerar i
detalj.

## Relaterade

- [[notes/median-i-stallet-for-snitt-for-wpm]]
- [[notes/session-livscykel]]
