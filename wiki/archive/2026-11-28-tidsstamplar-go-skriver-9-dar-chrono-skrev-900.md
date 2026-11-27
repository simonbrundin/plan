---
title: Tidsstämplar: Go skriver .9 där Chrono skrev .900
type: note
tags: [rust, paritet, tidsformat, projekt:wt-rustfix]
---

## Vad som hände

Under Rust-migreringen visade sig att Go och Rust skrev samma
tidsstämpel på två helt olika sätt:

```
Go:   2026-11-26T22:30:41.09356Z
Rust: 2026-11-26T22:30:41.093560Z
```

Go använder `time.Time`, som skriver RFC3339Nano: nio siffror i
andelen, alltså nollor tas bort från slutet. Rust använde Chrono,
som väljer mellan noll, tre, sex eller nio decimaler - och som
därför gav `.900` där Go gav `.9`.

Skillnaden syns i gränssnittet, där tiden visas, och den gjorde att
paritetsskripten inte kunde hitta avvikelser: skripten normaliserar
tidsstämplar till epok innan jämförelsen, just för att Go och Rust
kör i olika tidszoner. Normaliseringen döljer alltså formatet.

## Varför det var svårt att hitta

Tre saker höll ihop felen:

1. Skillnaden syns bara när nanosdelen slutar med noll. Ett värde som
   `93560000` nanos ger sex decimaler i båda språken, så det vanliga
   fallet ser identiskt ut. först `900000000` nanos avslöjade att
   Rust skrev `.900` och Go skrev `.9`.
2. Paritetsskriptet jämför epok, inte text. Ett formatfel syns inte
   förrän skriptet slutar normalisera.
3. Chrono gav sex decimaler i de flesta fall, vilket råkade matcha
   Go. Felet såg ut att saknas.

## Lösningen

Både `plan-api` och `plan-agent` har nu en egen tidsserialiserare
(`models/go_time.rs` respektive `go_time.rs`) som skriver nio
siffror och klipper bort avslutande nollor, precis som Go.

Testerna jämför mot utdata ur Go:s egen `time.Time.MarshalJSON`,
så att formatet inte kan glida isär igen utan att ett test
faller.

## Generell lärdom

När två implementationer ska bete sig lika, räcker det inte att
jämföra normaliserade värden. Normaliseringen är nödvändig för att
undvika brus från tidszoner, men den får inte dölja själva
formatet. Ett paritetsskript bör jämföra både den normaliserade
stformen och den råa texten för tidsstämplar - annars kan det
finnas fel som bara syns i gränssnittet.

En mindre avvikelse av samma slag hittades i wikin: Go tolkade
sidans datum som midnat i UTC när den flaggade föråldrade sidor,
medan Rust jämförde mot ett lokalt klockslag. Fyra sidor
flaggades därför i Rust men inte i Go. Rust använder nu UTC.
