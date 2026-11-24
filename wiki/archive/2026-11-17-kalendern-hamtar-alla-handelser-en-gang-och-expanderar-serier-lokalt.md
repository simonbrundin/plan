---
title: Kalendern hämtar alla händelser en gång och expanderar serier lokalt
type: note
tags: [kalender, feature, api, prestanda, projekt:plan]
---

Kalendern laddade tidigare händelserna för det synliga tidsfönstret, och varje
scrollning eller zoomning av tidsfönstret triggade ett nytt anrop till Go-API:t
(180 ms debounce). Det gav många små anrop och synlig fördröjning när vyn rörde
sig.

Nu är flödet:

- Klienten hämtar ALLT en gång med `GET /api/calendar-events?scope=all`.
  Serier kommer oexpanderade (en rad per serie med `recurrence_rule`) plus
  override-rader (`parent_event_id` + `original_start_at`).
- Serier expanderas lokalt i `src/nuxt/app/utils/calendarRecurrence.ts` med
  npm-biblioteket `rrule` (samma datumalgoritm som `rrule-go` på servern).
  Override ersätter förekomsten den hör till, och en avstängd override tar bort
  förekomsten - samma semantik som serverns `List`.
- Tidsfönstret filtreras fram client-side, så scrollning och zoom ger noll
  anrop. Expansionen cachas per dygnsrutat fönster, och rrule-instanser cachas
  per serie.

Varför detta väg: servern kan inte returnera "allt" för en serie som upprepas
varje dag i oändlighet, och den gamla 1000-gränsen i `expandCalendarEvent` gör
att en serie skulle bli avklippt om fönstret gjordes stort. Lokal expansion
löser båda: servern skickar rader, klienten räknar.

Verifierat i browser (Playwright): exakt 1 anrop vid öppning, 0 anrop vid
scrollning och zoomning, och samma händelser visas som tidigare (inklusive
flyttad och avstängd förekomst).

Viktigt att veta: `rrule` och `rrule-go` räknar i UTC. Servern expanderar i
sin egen tidszon (UTC i produktion, distroless-bilden har ingen TZ), så
klienten och servern ger samma förekomster. Men en daglig händelse kl 08:00
ligger alltså kvar på samma UTC-instans över ett DST-byte - den flyttas en
timme i lokal tid. Det är befintligt beteende, inte något den här ändringen
introducerade, men det blir synligt när klienten expanderar själv.
