---
title: Median i stället för snitt för WPM
type: note
tags: [beslut, stats, wpm, typetrainer, projekt:agents]
---

TypeTrainer visar nu median i stället för aritmetiskt snitt för WPM, Raw WPM och score i statistikvyn och WPM-grafen.

Varför: enstaka avvikande sessioner (t.ex. en session med mycket hög accuracy men låg fart, eller en avbruten session) drar snittet kraftigt. Medianen representerar den typiska sessionen bättre och är stabilare att följa över tid.

Följd: `stats.Median` är nu den gemensamma beräkningen. `session.renderGraph` använder `stats.Median` och den egna `calculateAverage` togs bort. Beräkningen är över de 30 senaste sessionerna i `PrintStats`.

Ändrad: 2026-11-08 (commit 1fbb53a i typetrainer).
