---
title: Gränssnitt och navigation
description: Sidor, vyindelning, sök och tangentbordsgenvägar i Plan - kortfattat så att du kan navigera utan att leta.
order: 2
icon: i-lucide-monitor
updated: 2026-11-28
group: Användaren
---

# Gränssnitt och navigation

Plan är en ensidig app (SPA) utan serverrendering. All rendering sker i
webbläsaren, och appen pratar direkt med Plan API:t - det finns ingen
proxy i Nuxt.

## Sidorna

| Sida | URL | Vad den är för |
| --- | --- | --- |
| Inbox | `/` | Snabb lägg till och de tankar som inte är mål än |
| Mål | `/goals` | Målen i träd, med undermål, filter och drag-and-drop |
| Gör nu | `/now` | Det som ska ske nu - prioriterat och återkommande |
| Prioritet | `/priority` | Prioriteringsläget, där du väljer vad som är viktigast |
| Målsida | `/goal/:id` | Ett mål i detalj: undermål, beroenden, beskrivning, metadata |
| Kalender | `/calendar` | Tidslinje av händelser, vertikal, med upprepningar |
| Avslutade | `/completed` | Klara mål och slutförda förekomster |
| Platser | `/platser` | Platser på karta - där uppgifter kan slutföras |
| Agent | `/agent` | Agentkörningar, händelser, godkännanden och chatt |
| Agentkostnad | `/agent-cost` | Tokenanvändning och kostnad per körning |
| Inställningar | `/settings` | Kalenderinställningar (sparas i webbläsaren) |
| Dokumentation | `/docs` | Den här dokumentationen |

## Sidomenyn

Menyn till vänster innehåller **Mål** (alla mål som grupper) och
**Navigering** (sidor ovan). Målen visas med antal undermål, så att
man snabbt ser var arbetet ligger. Klick på ett mål öppnar dess
målsida.

## Snabb lägg till

`n` öppnar ett fält där du kan skriva ett mål eller en kalenderhändelse
direkt. Fältet får fokus direkt, så att du kan börja skriva utan att
klicka. `Esc` stänger.

## Sök

`Ctrl+K` (eller `Cmd+K` på Mac) öppnar sökningen bland mål. Sökfältet
accepterar `↑`/`↓` för att markera ett mål och `Enter` för att öppna
det.

## Ångra

`z` ångra den senaste handlingen - att slutföra ett mål, sätta ett
datum, ändra en beskrivning. Ångringen gäller även återkommande mål,
där rätt förekomst markeras som gjord.

## Tangenter

Den kompletta listan finns i appen: tryck `?` för att öppna
hjälpmodalen. Snabbtolkningen:

| Tangent | Funktion |
| --- | --- |
| `?` | Visa hjälpen |
| `Ctrl+K` | Sök bland mål |
| `n` | Snabbt lägg till |
| `z` | Ångra senaste handlingen |
| `p` | Växla till/från Gör nu |
| `Space` | Leader-läge - tryck `Space` och sedan en tangent |
| `Esc` | Stäng modal / avbryt |

### Leader-läge

`Space` går in i leader-läge. Då kan du utföra kommandon som inte
krockar med textredigering. Lederns status visas i gränssnittet.

På målsidan:

| Kommando | Funktion |
| --- | --- |
| `Space` + `d` | Toggle slutfört |
| `Space` + `e` | Toggle evigt-märke |
| `Space` + `t` | Redigera beskrivning |
| `Space` + `c` | Sätt deadline |
| `Space` + `b` | Sätt startdatum |
| `Space` + `m` | Sätt uppskattad tidsåtgång |
| `Space` + `r` | Sätt upprepning |
| `Space` + `x` | Ta bort mål |
| `Space` + `s` | Toggle "påbörjat" |
| `Space` + `t` | Taggar på markerat mål |
| `Space` + `p` | Platser på markerat mål |

Utan leaderläge på målsidan: `j`/`k` för att navigera mellan
undermål, `h`/`l` för att gå till förälder eller öppna mål, `Enter`
för att lägga till undermål, `i`/`a` för att redigera titeln och `w`
för att göra målet beroende av ett annat.

## Målens utseende

En rad i listan visar bara de funktioner som faktiskt används. En rad
utan plats får ingen platsikon, en rad utan tidsåtgång får ingen
klocksymbol. Det gör att listorna blir lugna och att man snabbt ser
vad som är ifyllt.

Funktioner som kan sättas på ett mål:

| Funktion | Snabb tangent | Betydelse |
| --- | --- | --- |
| Slutförd | `d` | Målet är klart |
| Fas (status) | `s` / `Shift` + `s` | Definiera → Planera → Genomföra → Verifiera → Klar (framåt/bakåt) |
| Påbörjat | `Space` + `s` | Du har börjat på det |
| Evigt | `e` | Målet har ingen slutdatum - det pågår |
| Deadline | `c` | Senast då målet ska vara klart |
| Startdatum | `b` | När arbetet ska börja |
| Tidsåtgång | `m` | Uppskattad minuter |
| Upprepning | `r` | Återkommande - dagligen, varje vecka ... |
| Plats | `Space` + `p` | Var målet kan slutföras |
| Taggar | `Space` + `t` | Etiketter som "ute", "hem", "vinter" |
| Beroende | `w` | Ett annat mål måste vara klart först |

## Verifiering och "jag är klar"

På målsidan ligger en panel med verifieringen. Den har tre delar:

- **Verifieringsplan** - typ, villkor och styrka. Planen är valfri och
  kan läggas till, ändras eller tas bort. Utan plan fungerar målet
  precis som i dag.
- **Bevis** - det som faktiskt visar att målet blev gjort. Du kan
  skriva ett bevis, och agenten kan lämna ett.
- **Omdöme** - resultatet av verifieringen. Omdömet blir giltigt först
  när du godkänner det; du kan också avslå det.

Överst i panelen finns **Jag är klar** - en handling som avslutar
målet direkt, från vilken fas som helst. Det fungerar även för ett
mål som aldrig haft någon verifieringsplan.

Målet registrerar då hur det avslutades:

| Avslutades som | Betyder |
| --- | --- |
| Deklarerat | En människa sa att målet var klart |
| Verifierat | En verifiering gått igenom och en människa godkänt omdömet |

Det syns på målet och i listan av avslutade mål - skillnaden mellan
"jag tror det är klart" och "det här är bevisat" ska gå att se direkt.

I agent-docken, under fliken **Körning**, ligger samma information:
planen, bevisen och omdömet - med knapparna **Godkänn** och **Avslå**.

## Upprepning

Ett återkommande mål har en `recurrence_rule` (RRULE) och ett läge:

- **schedule** - nästa förekomst räknas fram från dagens datum.
- **push** - varje förekomst flyttas framåt när den slutförs.

Slutförda förekomster sparas i `goal_completions`, en rad per
förekomst med vilket datum den gällde. Det gör att ett återkommande
mål kan vara klart den 1:a och den 15:e men inte den 8:e.

## Prioritering

Prioritetsläget (tangent `p`) visar de mål som faktiskt går att göra
nu - filtrerat på startdatum, deadline och beroenden. Du markerar
målen i prioritetsordning, och listan följer din markering.

## Utseende

Appen är mörk i grunden och följer inte systemets ljustema - det är
ett val, inte ett fel. Färgerna kommer från designsystemets tokens
(`surface`, `line`, `ink`, `accent`), så att hela appen byter utseende
när en token ändras.
