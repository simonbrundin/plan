---
title: Manuell sortering av mål på priority
type: note
tags: [feature, ui, priority, go, nuxt, projekt:plan]
---

På /priority kan det markerade målet nu flyttas med ArrowUp/ArrowDown. Ordningen är separat från prioriteringsvikten och sparas i goals.priority_order via PATCH /api/goals/priority/reorder. Dolda mål behåller sin plats när ett synligt mål flyttas. Listan använder priority_order först och vikt som sekundär ordning.
