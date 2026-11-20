# Inställningar för kalendern

## Problem

Kalendern har några beteenden som är svåra att avgöra i förväg. Framför allt
hur många förekomster av en återkommande serie som får synas samtidigt. En
gräns på sju är ett bra utgångsläge, men den är beroende av hur kalendern
används: den som har ett dagligt möte varje morgon vill se alla, den som har
en träningsrutin varje dag vill se dagsformen och inte alla sju.

Det går inte att lösa med en bättre standardinställning. Gränsen måste gå
att ändra.

## Lösning

En egen sida, `/settings`, med kalenderns inställningar. Meningsmenyn pekar
redan dit, så sidan fyller en lucka som redan fanns.

Inställningarna:

- **Max antal instanser per serie** - hur många förekomster av en återkommande
  serie som får synas samtidigt i vyn. Så snart det blir en till, döljs hela
  serien. Standard är sju.
- **Dölj återkommande serier** - en strömbrytare. Med avstängd döljning visas
  alla serier, och kolumnen kan bli full.

Inställningarna sparas i webbläsaren (`localStorage`) och gäller på den
enhet de gjordes på. De är avsedda att ändras sällan, och en inställning per
enhet är rimligare än att bygga ett konto-underlag för dem. Att synka dem
mellan enheter är en möjlig senare utbyggnad, som kräver en endpoint i
Go-API:t.

## Implementation

- `app/composables/useCalendarSettings.ts` läser och sparar inställningarna.
  En trasig sparad inställning ignoreras och faller tillbaka på standarden.
- `app/pages/settings.vue` visar sidan.
- `app/layouts/default.vue` har en länk till sidan i sidomenyn.

## Avgränsningar

- Inställningarna följer webbläsaren, inte kontot. På en ny enhet gäller
  standarden.
- Det finns ingen inställning för döljning av enskilda händelser. Det är
  avsiktligt: enstaka händelser döljs aldrig, de hamnar i "+N till".

## Acceptance Criteria

- [x] Sidan `/settings` går att nå från sidomenyn och från användarmenu.
- [x] Gränsen för antalet instanser per serie går att ändra, och ändringen
      gäller direkt i kalendern.
- [x] Döljningen av återkommande serier går att stänga av.
- [x] Inställningarna överlever en omladdning av sidan.

## Status

- [x] Define
- [x] Plan
- [x] Execute
- [x] Verify
