---
title: Bigram-algoritmen
type: note
summary: Hur bigram-statistiken lagras, rankas och används för att välja ord som tränar svaga teckenpar.
tags: [algoritm, bigram, ordlista, urval]
created: 2026-10-21
updated: 2026-10-21
status: active
related: [notes/arkitektur, notes/wordlist-kallor]
---

Kärnidén: de flesta skrivfel och fördröjningar sitter i **teckenpar**
("sj", "er", "st", "an"), inte i enskilda bokstäver. Genom att mäta per
bigram kan träningslistan prioritera exakt de kombinationer som går dåligt.

## Mått

Per bigram (`internal/bigram/bigram.go`):

- `Count` - antal gånger bigramen har skrivits.
- `TotalMs` - total tid.
- `Errors` - antal fel.

Därifrån:

- `AvgMs = TotalMs / Count`
- `ErrorRate = Errors / Count`
- `IsWeak(avgTime)` - sant om `ErrorRate >= weak_threshold` (0.15) **eller**
  `AvgMs > avgTime * 1.5`. Kräver `Count >= min_samples` (10).
- `DifficultyScore(avgTime)` - `ErrorRate * error_weight (0.7)` plus
  tidskomponenten `(AvgMs/avgTime, takad vid 3) / 3 * time_weight (0.3)`.

Tidskomponenten är normaliserad mot **globala** medeltiden
(`GlobalAvgTime()`, default 150 ms om ingen data finns).

## Urval

`stats.WeakBigrams` rankar alla svaga bigram efter `DifficultyScore` och
returnerar en sorterad lista (svårast först).

`wordlist.AdaptiveSelection(weakBigrams, count)` gör sedan:

1. Bygger en kandidatlista per svag bigram: ord som innehåller bigramen,
   slumpade och sedan stabilt sorterade på antal svaga bigram i ordet
   (flera = träningsvärdesare).
2. **Varvarv** - ett ord från varje bigram i tur, så att träningen sprids
   över flera problemområden i stället för att köra samma mönster igen
   och igen.
3. Under urvalet gäller två tak:
   - `maxPerFamily = 1` - högst ett ord per inflectionsfamilj
     (prata/pratar/pratade hör ihop).
   - `maxPerEnding = 3` - högst tre ord med samma slutpar, så att inte
     hela sessionen blir ord på "-at".
4. Samma ord väljs aldrig två gånger i en session.
5. Om listan blir för kort fylls den på med slumpade ord (familj- och
   ändelsetaket hålls), och som sista utväg tas de återstående orden även
   om taken bryts.

## Familjgruppering

`computeFamilies` sorterar orden alfabetiskt och grupperar sedan ord som
delar ett prefix på minst 4 tecken (`familyPrefixLen`), eller där det ena
ordet är ett prefix till det andra med minst 3 tecken ("hus"/"huset").

## Ordval för övriga lägen

- `RandomWords(n)` - slumpade, utan återläggning.
- `PracticeWords(bigram, maxPerFamily)` - alla ord med bigramen, med
  familjetak. Fylls på med fler ord ur samma familjer om urvalet blir
  för litet.

Se [arkitektur](arkitektur.md) för var koden ligger.

## Relaterade

- [[notes/arkitektur]]
- [[notes/wordlist-kallor]]
