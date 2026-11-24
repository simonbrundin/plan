---
title: Kalenderns nivåväxling animeras: crossfade, pop-in och hysteresis
type: note
tags: [kalender, animation, css, gotcha, session, projekt:plan]
---

## Varför

Kalendern byter presentationsnivå (full, compact, daily, monthly, yearly) när
tidsfönstret passerar gränsvärden. Innan detta byttes nivån omedelbart, vilket
gjorde att händelser "hoppade" till nya former mitt i en scrollrörelse. En
kontrollerad övergång gör det lättare att följa med i vad som händer.

## Lösning

Tre delar, i `src/nuxt/app/components/VerticalCalendar.vue` och
`src/nuxt/app/utils/calendarDetail.ts`.

### 1. Hysteresis mot fladdring

Nivågränserna (48 h, 7 d, 45 d, 3 år) bestäms av tidsfönstrets längd. Utan
hysteresis kan en användare som stannar precis vid en gräns få nivån att växla
fram och tillbaka varje scrollrörelse.

`DETAIL_HYSTERESIS = 0.08` i `calendarDetail.ts` gör att nivån byts först när
fönstret passerat gränsen med 8 procent marginal. Mätt i Chromium: 6 nivåbyten
per passage över tre gränser fram och tillbaka - exakt det förväntade, ingen
fladdring.

### 2. Crossfade mellan lager

`detailLevel` är en `ref`, inte en `computed`, eftersom den måste uppdateras
i samband med zoomningen. Innehållet renderas inom ett `<Transition
name="detail-swap">` med en `:key` på detaljnivån. Utan `mode="out-in"` - det
ger en tom period mellan nivåerna. Lagerintervallet tonar in (opacity + liten
upptrörelse) medan det gamla tonar ut.

Mätt övergångstid: cirka 250 ms, med två lager parallellt under själva
crossfaden.

### 3. Pop-in för enskilda element

Varje händelskort och varje prick får sin egen in-animation, med en liten
stagger (`18 ms` per kort, `14 ms` per prick).

## Gotchas

**Vue:s enter-klasser läggs på lagret, inte på barnen.** En selektor som
`.detail-swap-enter-from .calendar-marker` matchar alltså aldrig prickarna
under enter-fasen - barnen ärver inte förälderns enter-klass. Lösningen blev
CSS `animation` med `animation-fill-mode: backwards` på barnen. Animationen
körs när elementet monteras, alltså vid nivåbytet.

**`transform` används redan av prickarna.** Prickarna har
`transform: translateY(-50%)` för att centreras vertikalt. En keyframe som
animerar `transform` skriver över det, och pricken hoppar. Lösningen är de
fristående CSS-egenskaperna `scale` och `translate`, som appliceras ovanpå
`transform` utan att kollidera med den.

**Animationer vinner över inline-stilar.** Det påverkade prickarnas
intensitet, som sattes som inline `opacity` (`0.45 + 0.55 * count/3`). Med
`opacity` reserverad för in-animationen hoppade pricken tillbaka till 0.63
efter att animationen var klar. Intensiteten flyttades därför till
`filter: opacity(...)`, som appliceras efter opacity och därför ger samma
utseende utan att stöda animationen.

**Hysteresisen fördröjer bytet lite.** Det är avsiktligt - utan den skulle
nivån fladdra. Användaren upplever bytet som en snabb övergång i stället för
ett omedelbart byte.
