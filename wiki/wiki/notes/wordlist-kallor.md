---
title: Ordlistan finns på två ställen
type: note
summary: Ordlistan finns på två ställen och måste hållas i synk.
tags: [ordlista, data, risk, teknik]
created: 2026-10-21
updated: 2026-10-21
status: open
related: [notes/bigram-algoritmen, notes/utvecklingsflode]
---

`words/words.txt` och `src/go/internal/wordlist/words.txt` är **två kopior
av samma fil**. De var identiska vid senaste kontroll (2026-10-21).

## Varför det är ett problem

- Det är `src/go/internal/wordlist/words.txt` som används - den embeddas
  med `//go:embed words.txt` och läses vid `wordlist.New()`.
- `words/words.txt` i roten är bara där som "källa", men **ingenting synkar
  dem**. Ingen script, ingen Makefile, ingen CI.
- Ändrar du den ena och glömmer den andra får appen en annan ordlista än
  den du tror.

## Format

```text
# Kommentar-rad
ord,frekvensrang
```

Frekvensrangen är lägre = vanligare. `parseLines` hoppar över tomma rader
och rader som börjar med `#`, `splitLine` delar på första kommatecknet.

Orden sorteras efter frekvens *innan* bigram-indexet byggs - det var en bugg
i den första commiten ("Fixed bug in wordlist bigram index (was built before
sorting)").

## Dubbletter

Filen innehöll tidigare upprepade ord ("sig", "ett", "ganska" med olika
frekvensrang). `wordlist.New()` hoppar över dubbletter via en `seen`-mapp -
men bara det **första** ordet behålls, vilket gör att den överlevande
frekvensrangen kan vara den sämre av de två. Filen rensades den
2026-10-21 (1025 -> 694 rader).

## Öppen fråga

Ska `words/words.txt` tas bort, eller ska ett `go:generate`-steg kopiera
den till embeddingsplatsen? Just nu är svaret "ingen vet", vilket är
sämst.

Se [bigram-algoritmen](bigram-algoritmen.md) för hur listan används.

## Relaterade

- [[notes/bigram-algoritmen]]
- [[notes/utvecklingsflode]]
