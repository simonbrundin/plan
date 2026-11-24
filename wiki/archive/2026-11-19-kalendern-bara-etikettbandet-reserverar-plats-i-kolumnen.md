---
title: Kalendern: bara etikettbandet reserverar plats i kolumnen
type: note
tags: [kalender, bug, layout, session, projekt:plan]
---

Användaren rapporterade att bara en av flera händelser syntes 'i toppen'. Orsaken: packningen i en kolumn lade varje rad efter FÖREGÅENDE rads fulla höjd. En lång händelse (t.ex. Arbete 21:00-04:00) tog då 209 px av kolumnen, och alla andra händelser skjuts nedanför den - de hamnade långt under sina egna tider.

**Fix:** bara etikettbandet (de översta 16 px där titeln står) reserverar plats. Radens höjd följer fortfarande händelsens längd, men bara bandet kan kollidera. En kort händelse inne i en lång händelse behåller därför sin egen tid och ritas ovanpå den långa radens tomma yta.

**Följder som behövdes:**
- Radens innehåll måste ligga i överkant (align-items: flex-start), annars hamnade titeln mitt i den långa raden och kunde bli överklappad.
- Rader måste ritas i tidsordning (z-index = radens topp), annars kunde den långa raden måla över de korta.
- Bakgrunden måste vara sluten (color-mix mot --color-surface, inte transparent), annars blev texten över en annan rad svårläst.
- 'clipped' betyder nu att ETIKETTbandet inte ryms, inte att hela raden sticker ut. En lång händelse nära botten ska fortfarande visas.

**Lärdom:** När flera element ska dela en yta, packa på den del som faktiskt bär informationen (titelbandet), inte på hela elementet. En stor yta som inte bär text ska inte skjuta undan saker som ligger i den.
