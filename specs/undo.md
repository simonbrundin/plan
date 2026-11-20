# Ångra senaste handlingen

## Problem

Tangenttrycken i appen gör det lätt att råka göra fel. Ett "d" på ett mål
slutför det direkt - men slutförandet går inte att backa. Det enda sättet
att återställa ett mål är att öppna det och ändra status manuellt.

Det gäller även de andra tangenttrycken: "s" (pågående), "e" (evigt),
"c" (deadline) och "r" (upprepning) ändrar målet direkt utan återställnings
möjlighet.

En ångra-tangent gör att misstag kan rättas utan att lämna tangentbordet -
och utan att riskera att den rätta handlingen också sker.

## Lösning

Tangenten **Z** ångrar den senaste handlingen. Den är global och fungerar
på alla sidor: en handling som gjorts på målsidan kan ångras från
kalendern.

### Undo-stapel

Stapeln ligger i `app/composables/useUndoStack.ts` och är global (modulnivå),
eftersom appen körs utan SSR och en handling ska kunna ångras från en annan
sida än den gjordes på.

En post består av en beskrivning av handlingen och en funktion som
återställer tillståndet:

```ts
interface UndoEntry {
  label: string;
  undo: () => Promise<void> | void;
}
```

Den rena logiken ligger i `app/utils/undoStack.ts` (push/pop och gränsen på
antalet poster) så att den går att enhetstesta utan Nuxt.

Stapeln rymmer 25 handlingar. När gränsen nås faller den äldsta posten bort.

### Vad som kan ångras

Målens handlingar registreras via `app/composables/useGoalUndo.ts`, som
sparar det gamla värdet innan ändringen:

| Handling | Ångras med |
| -------- | ---------- |
| Slutför ett mål (`d`) | `PATCH goals/:id` med `finished: null` |
| Slutför en förekomst av ett återkommande mål (`d`) | `DELETE goals/:id/completions` |
| Markera som pågående (`s`) | `PATCH goals/:id` med `started: null` |
| Evigt-märke (`e`) | `PATCH goals/:id` med `eternal: false` |
| Deadline (`c`) | `PATCH goals/:id` med `due: null` |
| Upprepning (`r`) | `PATCH goals/:id` med `recurrence_rule: null` |

Ingen backendändring behövs - API:t har redan alla anrop som krävs.

### Gränssnitt

- **Z** ångrar senaste handlingen. Tangenten ignoreras i textfält, när
  sökningen är öppen och i leader-läge.
- Ett meddelande visas efter ett ångrande: "Ångrade" + beskrivningen av
  handlingen. Finns inget att ångra visas "Inget att ångra".
- En liten etikett i sidans nedre vänstra hörn visar vilken handling Z
  ångrar, så att det är synligt vad som händer om tangenten trycks.
- Hjälpmodalen listar Z under Normal Mode.

### Uppdatering av vyn

Sidor som har egna listor (/goals, /priority) lyssnar på undo-stapelns
`revision`, som räknas upp efter varje ångrande, och hämtar om listan.
Målsidan bygger om sin data ur grafen i stället.

## Avgränsningar

- Det går inte att ångra en borttagning av ett mål eller en kalenderhändelse.
- Det går inte att ångra en redigering av titel eller beskrivning.
- Ångrar man flera förekomster av samma återkommande mål tar API:t bort den
  senast slutförda förekomsten - den behöver alltså vara den senaste.
- Att ångra en upprepning startar en ny serie: de tidigare slutförandena
  hör till den gamla regeln och finns inte kvar.

## Acceptance Criteria

- [x] Z ångrar den senaste handlingen.
- [x] Slutförandet av ett mål (`d`) kan ångras.
- [x] Slutförandet av en förekomst av ett återkommande mål kan ångras.
- [x] Pågående, evigt, deadline och upprepning kan ångras.
- [x] Ett ångrande syns i gränssnittet (meddelande + etikett).
- [x] Z fungerar från alla sidor.
- [x] Det finns enhetstester för undo-stapeln.
