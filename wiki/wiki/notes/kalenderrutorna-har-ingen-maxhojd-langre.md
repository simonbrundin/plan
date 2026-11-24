---
title: Kalenderrutorna har ingen maxhöjd längre
type: note
tags: [kalender, feature, ui, session, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Användaren klagade på att en händelse i kalendern såg ut att ha en maxhöjd: längden på raden stämde inte med hur mycket tid som visades för timmen, och det blev förvirrande.

**Varför:** Radens höjd begränsades till en fjärdedel av vyhöjden (maxRow = max(minRow*4, height/4) i app/utils/calendarLayout.ts). I en dygnsvy på 748 px blev taket 187 px, så en åtta timmars händelse ritades 187 px hög - den såg ut att vara sex timmar lång. En tolftimmars händelse blev lika stor, så längden gick inte att läsa av alls.

**Vad som gäller framöver:**
- Radens höjd är exakt proportionell mot tidsaxeln: höjd = längd * (vyhöjd / tidsfönster). Ingen övre gräns.
- Nedre gränsen kvar: MIN_ROW_PX = 16 px, så att en kort händelse i en utzoomed vy fortfarande har en läsbar titel.
- Bara etikettbandet (16 px) reserverar plats i kolumnen, så en lång händelse kan ha andra rader ovanpå sin tomma yta. Raderna senare i tiden ritas över den långa raden, så ingen titel täcks.
- Verifierat i browsern mot dev: Sömn 8 h = 249 px, Bilbanan hämtas 1 h = 31 px, Cleos kalas 2 h = 62 px i en vy på 748 px/24 h. Inga titlar täcks.

**Fallgropar:** bun run build i src/nuxt delar .nuxt-katalogen med en körande nuxt dev - det får devservern att visa 'dist directory has been removed. Restarting Nuxt...' och svara 503. Bygg i en ren katalog, eller starta om devservern efter bygget.
