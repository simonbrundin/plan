# Flera kalendrar med egna färger

## Problem

Plan har en tidslinjekalender där alla händelser ligger i samma flöde. En
användare som vill separera jobb från fritid, eller ha en kalender per
projekt, kan inte göra det - händelserna blandas och alla har samma färg.

Det finns dessutom ingen plats att välja kalender när en händelse skapas,
eftersom den enda fungerande vägen att skapa händelser är snabbfältet.

## Lösning

Användaren kan skapa hur många kalendrar som helst. Varje kalender har ett
namn och en egen färg. Händelser hör till en kalender och har alltid
kalenderns färg.

### Data

- `calendars`: `id`, `user_id`, `name`, `color`, `icon`, `is_default`,
  `sort_order`, `created_at`, `updated_at`.
- `calendar_events.calendar_id`: främmande nyckel till `calendars` med
  `ON DELETE SET NULL`. Null betyder "ingen kalender".

### Standardkalendern

En användare kan ha högst en standardkalender. Den markeras med
`is_default`, och en unik index (`calendars_user_default_idx`) på
`(user_id) WHERE is_default` hindrar att det finns två.

En standardkalender innebär att nya händelser hamnar där utan att man
behöver välja kalender:

- Händelser som skapas utan `calendar_id` hamnar i standardkalendern.
- `calendar_id: 0` betyder däremot uttryckligen "ingen kalender".
- Formuläret och snabbfältet preselectar standardkalendern.

Sätter man en ny kalender som standard avmarkeras den gamla i samma
transaktion, så att det aldrig kan bli två standardkalendrar.

### Färgregler

Färgen löses upp i denna ordning:

1. Kalenderns färg.
2. Standardfärgen `#6366f1` för händelser utan kalender.

En händelse kan inte ha en egen färg - färgen ägs av kalendern. Logiken
ligger i `src/nuxt/app/utils/calendarColors.ts` och är testad med
enhetstester.

### Färgkod

Kalenderns färg är en hexkod som användaren väljer själv. Både `#6366f1`
och `#abc` accepteras; tre siffror skrivs ut som sex (`#abc` blir `#aabbcc`),
och versaler sänks. En färgkod som inte är giltig avvisas - av servern med
`Invalid color` och i gränssnittet med ett felmeddelande i fältet.

### API

| Method | Route | Beskrivning |
| ------ | ----- | ----------- |
| GET  | `/api/v1/calendars` | Listar användarens kalendrar. |
| POST | `/api/v1/calendars` | Skapar en kalender. |
| PATCH | `/api/v1/calendars/:id` | Ändrar namn, färg, ikon eller standardmarkering. |
| DELETE | `/api/v1/calendars/:id` | Tar bort kalendern. |

Kalenderhändelserna får ett `calendar_id`-fält. `PATCH` med `calendar_id: 0`
tar bort kopplingen, `calendar_id` utan värde lämnar den som den är.

En händelse som pekar på en kalender som inte tillhör användaren avvisas
med 403.

### Gränssnitt

- **Kalenderlistan** ligger i en lista som öppnas med knappen "Kalendrar" i
  verktygsraden. Kalendrarna ligger alltså inte som etiketter ovanför
  tidslinjen - de öppnas med en knapp när man behöver dem.
- **Kalenderhanteraren** skapar, döper om, färgar och tar
  bort kalendrar. Stjärnan sätter kalendern som standardkalender, och ögat
  döljer eller visar kalenderns händelser i tidslinjen.
- **Färgen sätts med färgkod eller ur paletten.** Både i skapa-formuläret
  och i varje rad finns ett fält där hexkoden skrivs, bredvid en
  förhandsvisning av färgen. Paletten med åtta färger är genvägar för den
  som inte vill skriva kod, och den egna färgknappen öppnar systemets
  färgväljare.
- **Händelseformuläret** har en kalenderväljare. Händelsen får kalenderns
  färg, och färgen följer med om kalendern byts.
- **Snabbfältet** har en kalenderväljare när det skapar en händelse, så att
  det går att placera händelser i rätt kalender direkt från tangentbordet.
- **Kalenderlistan** markerar standardkalendern med en stjärna.

## Avgränsningar

- Kalendrar har ingen egen vy. Alla kalendrar visas i samma tidslinje och
  filtreras fram och tillbaka.
- Att ta bort en kalender tar inte bort händelserna i den - de blir
  kalenderlösa och får standardfärgen.
- Delade kalendrer mellan användare finns inte.

## Acceptance Criteria

- [x] Användaren kan skapa flera kalendrar, utan gräns för antalet.
- [x] Varje kalender har en egen färg.
- [x] Händelser i en kalender visas i kalenderns färg.
- [x] En händelse kan inte ha en egen färg.
- [x] En kalender kan döljas och visas i tidslinjen.
- [x] Kalendrarna ligger i en lista som öppnas med en knapp, inte som
      etiketter ovanför tidslinjen.
- [x] Händelser kan flyttas mellan kalendrar.
- [x] En borttagen kalender lämnar händelserna kvar utan kalender.
- [x] Snabbfältet kan placera en händelse i en vald kalender.
- [x] En kalender kan sättas som standardkalender, och bara en kan vara det.
- [x] Nya händelser hamnar i standardkalendern när ingen kalender väljs.
- [x] Formulär och snabbfält preselectar standardkalendern.
