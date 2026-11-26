---
allowed-tools: Write, Edit, WebFetch, Task, Fetch
description: Skapa nytt slash command
---

# Skapa slash command

Baserat på `INPUT_PROMPT` följ `Workflow` för att skapa en nytt kommando med
formatet i `Mall`. Innan du börjar, använd verktyget WebFetch för allt i
`Dokumentation`.

## Variabler

INPUT_PROMPT: $ARGUMENTS

## Workflow

- Vi bygger en ny prompt för att uppfylla begäran specificerad i `INPUT_PROMPT`.
- Spara den nya prompten till
  `/home/simon/repos/software-agent/app/commands/<name_of_prompt>.md`
  - Namnet på prompten ska vara logiskt baserat på den `INPUT_PROMPT`.
- IMPORTANT: Prompten måste vara i formatet i `Mall`.
  - Skapa inte några ytterligare sektioner eller rubriker som inte finns i
    `Mall`.
- IMPORTANT: När du arbetar igenom `Mall`, ersätt varje block av
  `<någon begäran>` med begäran som anges inom klammern.
- Observera att vi kallar dessa 'prompter', de är också kända som anpassade
  slash-kommandon.
- Använd ett Task-verktyg per dokumentationspost för att köra underuppgifter och
  samla in dokumentation snabbt i parallell med `Task` och `WebFetch`.
- Ultra Think - du driver en prompt som skapar en prompt. Håll fokus på
  detaljerna i att skapa den bästa högkvalitativa prompten för andra AI-agenter.
- Om den `INPUT_PROMPT` begärde flera argument, ge varje sitt eget h2-rubrik och
  placera sedan `$ARGUMENTS` precis under deras respektive h2-rubrik. Dessa
  prompter använder ett indexbaserat system.
- Observera, om inga variabler begärdes eller nämndes, skapa inte en
  variabler-sektion.
- Tänk igenom vilka som är statiska variabler kontra dynamiska variabler och
  placera dem därefter, med dynamiska variabler först och statiska variabler
  därefter.
  - Föredra notationerna `$1`, `$2`, ... över `$ARGUMENTS`.

## Dokumentation

- https://docs.anthropic.com/en/docs/claude-code/slash-commands
- https://docs.anthropic.com/en/docs/claude-code/common-workflows#create-custom-slash-commands
- https://docs.anthropic.com/en/docs/claude-code/settings

## Mall

```md
---
allowed-tools: <allowed-tools komma separerade>
description: <beskrivning vi använder för att identifiera den här prompten>
argument-hint: [<argument-hint för den första dynamiska variabeln>], [<argument-hint för den andra dynamiska variabeln>]
---

# <namn_på_kommando>

## Syfte

<prompt syfte: här beskriver vi vad prompten gör på en hög nivå och hänvisar
till eventuella avsnitt vi skapar som är relevanta, som avsnittet
`Instruktioner`. Varje prompt måste ha ett avsnitt `Instruktioner` där vi
redogör för instruktionerna för prompten i punktform>

## Variabler

<NAMN_PÅ_DYNAMISK_VARIABEL>: $1 <NAMN_PÅ_INPUT>: $ARGUMENTS
<NAMN_PÅ_DYNAMISK_VARIABEL>: $2 <NAMN_PÅ_STATISK_VARIABEL>: <NÅGOT STATISKT>

## Workflow

<stegvis numrerad lista över uppgifter som ska slutföras för att utföra
prompten>

## Relevanta filer

<en lista med relevanta filer att ha i åtanke när kommandot körs>

## Exempel

<exempel på filer eller kod hur kommandot ska göra>

## Rapport

<detaljer om hur prompten ska svara tillbaka till användaren utifrån prompten>
```
