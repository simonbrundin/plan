---
allowed-tools: Glob, Read, Edit, Write, question
description: Förfina och förbättra en feature-fil
argument-hint: [valfri-feature]
---

# Förfina feature

## Syfte

Fråga användaren vilken feature-fil från `.features/` mappen att granska och
förbättra dess rubrik, beskrivning och alla scenarier enligt BDD-best-practices.

## Variabler

FEATURE_FIL: $1 <valfri feature att förfina: om ingen anges, lista alla features
och låt användaren välja>

## Workflow

1. Hämta alla `.feature`-filer från `.features/` mappen med Glob
2. Om `$FEATURE_FIL` inte anges:
   - Visa lista på alla tillgängliga features
   - Använd verktyget `question` för att låta användaren välja en
3. Läs den valda feature-filen fullständigt
4. Analysera feature-filen och identifiera förbättringar:
   - Rubrik: Är den tydlig och beskriver business value?
   - Beskrivning: Finns tydlig "Som...", "Vill jag...", "För att..."?
   - Scenarier: Följer de Given-When-Then strikt? Är de korta och fokuserade?
   - Finns anti-patterns som bör undvikas?
5. Presentera en lista med specifika förbättringsförslag
6. Ifall du anser att det borde finnas med fler scenarion, föreslå dem också.
7. Fråga användaren om de vill att du ska applicera förslagen
8. Om ja, applicera förbättringarna med Edit-verktyget

## Relevanta filer

- `.features/` - katalog för alla feature-filer
- /home/simon/repos/software-agent/app/commands/test-create-feature.md - tips
  för bra BDD-scenerier

## Exempel

Vid körning av `/refine-feature` utan argument:

```
Tillgängliga features:
1. message-trigger-workflow.feature
2. user-login.feature
...

Välj en feature att förfina: [1]
```

## Rapport

Presentera en sammanfattning av:

- Vilken feature som valdes
- Vilka förbättringar som föreslogs
- Om förslagen applicerades eller inte
