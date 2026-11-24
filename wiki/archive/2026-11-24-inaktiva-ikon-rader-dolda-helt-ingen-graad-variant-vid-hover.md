---
title: Inaktiva ikon-rader dolda helt - ingen gråad variant vid hover
type: note
tags: [ui, ikon, css, projekt:plan]
---

Symtom: på vissa mål syntes ikoner (t.ex. start-ikonen) trots att funktionen inte var aktiv. De syntes som en gråad/outdimad variant när raden svävas över - CSS-regeln '.opt-row:hover .opt-icon { display: inline-flex }' i src/nuxt/app/assets/css/main.css.

Beslut: en ikon visas bara om funktionen faktiskt är på. Hover-regeln tas bort - en inaktiv ikon är helt dold, aldrig gråad. Funktionerna läggs till med tangenterna (c deadline, r upprepning, b startdatum, m tidsåtgång, p plats, a agent, s påbörjat, e evigt, leader + t taggar) i stället för med musen.

Extra fix: start-ikonen ('påbörjad') räknas nu som aktiv bara om målet är started OCH inte finished - i databasen fanns 72 mål med både started och finished satt, så klara mål visade start-ikonen som på. Ändrad i goal/[id].vue (huvudmål + undermål), goals/index.vue och now.vue.

Supersederar: notes/dolda-ikoner-i-barnmalsrader-pa-malsidan-visa-vid-hover (där beslutet var att dolda ikoner skulle synas vid hover).

Verifierat headless (chromium, Nuxt-build mot Go-API mot Neon dev): på /goals och /goal/39 är listan över synliga ikoner identisk i vila och efter hover - bara aktiva ikoner (evigt, agent, deadline, upprepning, start) syns.
