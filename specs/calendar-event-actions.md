# Ta bort och duplicera kalenderhändelser

## Problem

Användaren kan skapa och ändra kalenderhändelser, men inte ta bort dem.
En händelse som inte längre ska finnas måste redigeras till en tom rad
(titel och beskrivning), och den ligger kvar i tidslinjen.

Det finns inte heller något sätt att duplicera en händelse. En vanlig
uppgift - att göra samma möte igen, eller kopiera en träning till en annan
dag - kräver att man skriver om allt från början.

Återkommande serier gör det ännu svårare. Att ta bort en enskild förekomst
måste göras på serienivå, och då försvinner hela serien.

## Lösning

Varje händelse får två åtgärder i redigeringsformuläret: **Duplicera** och
**Ta bort**.

### Duplicera

Kopian får samma titel, beskrivning, ikon, kalender, mål och samma tid.
Den öppnas direkt i formuläret, så att tiden kan justeras med en gång.

En förekomst i en serie kopplas loss från serien: kopian blir en egen
händelse med samma tid. Hela serier kopieras däremot med samma
upprepningsregel, så att kopian upprepas precis som originalet.

### Ta bort

Omfattningen beror på vilken rad som egentligen ska försvinna, och väljs
med samma omfattningsväljare som används vid redigering:

| Omfattning | Vad som händer |
| ---------- | --------------- |
| Bara denna förekomst | Förekomsten stängs av. Serien och de övriga förekomsterna finns kvar. |
| Denna och följande | Serien slutar strax före förekomsten, och de förekomster som försvann tas bort. |
| Hela serien | Serien tas bort, tillsammans med sina förekomster. |

En händelse som inte hör till en serie tas bort som den är.

### Data

Ingen schemaförändring behövs. Enstaka förekomster stängs av med en
override-rad vars `status` är `cancelled` - en rad som redan finns
(`parent_event_id` + `original_start_at`) uppdateras i stället för att
en ny rad ska skapas.

Serien kortas av med en ny upprepningsregel: den befintliga regeln får ett
`UNTIL` strax före den valda förekomsten. Ett befintligt `UNTIL` eller
`COUNT` tas bort först, annars skulle den nya gränsen ignoreras.

Rader för förekomster som ligger utanför den avslutade serien tas bort.
Annars skulle de ligga kvar som egna händelser i vyn.

### API

Inga nya endpoints behövs. Befintliga används:

| Method | Route | Beskrivning |
| ------ | ----- | ----------- |
| POST | `/api/v1/calendar-events` | Skapar kopian vid duplicering. |
| DELETE | `/api/v1/calendar-events/:id` | Tar bort en händelse eller en hel serie. |
| PATCH | `/api/v1/calendar-events/:id` | Stänger av en förekomst (`status: cancelled`) eller kortar av en serie (`recurrence_rule`). |
| POST | `/api/v1/calendar-events/:id/occurrences` | Stänger av en förekomst som saknar egen rad. |

### Gränssnitt

- **Redigeringsformuläret** har knapparna "Duplicera" och "Ta bort".
- Knappen "Ta bort" frågar först vad som ska tas bort, så att en hel serie
  inte försvinner av misstag.
- Seriernas förekomster pekar på seriens masterrad. En förekomst med egen
  rad pekar i stället på föräldern - det är den raden som ska ändras eller
  tas bort.

## Avgränsningar

- Duplicering kopierar tiden som den är. Kopian ska flyttas manuellt, eller
  i formuläret som öppnas.
- Det går inte att ångra en borttagning.
- Att duplicera en hel serie kopierar serien som den sågs, inte de enskilda
  förekomsterna.

## Acceptance Criteria

- [x] En händelse kan dupliceras, och kopian får samma innehåll och tid.
- [x] Kopian öppnas i formuläret efter duplicering.
- [x] En enstaka händelse kan tas bort.
- [x] En enskild förekomst i en serie kan tas bort, utan att serien försvinner.
- [x] En serie kan avslutas vid en vald förekomst.
- [x] En hel serie kan tas bort.
- [x] En förekomst med egen rad ändras och tas bort på rätt rad.
