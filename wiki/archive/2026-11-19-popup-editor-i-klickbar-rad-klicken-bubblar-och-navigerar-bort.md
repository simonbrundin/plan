---
title: Popup-editor i klickbar rad: klicken bubblar och navigerar bort
type: note
tags: [bug, ui, nuxt, goals, projekt:plan]
---

Att rendera en editor/popup INUTI en klickbar rad gör att klick i popupen bubblar upp till radens klick-handler. I /goals navigerar raden till målsidan, så att trycka 'Spara' i upprepningseditorn hoppade ur vyn - det såg ut som att 'inget hände'.

FALLER: en Vue-komponent i en lista-rad ärver inte radens stopp. Lösningen är en wrapper-div med @click.stop (och @keydown.stop) runt komponenten - inte på komponenten själv.

FÄLLA 2: appens tangent-commandon följer det MARKERADE målet. På målsidan (/goal/:id) är isGoalSelected false tills man markerar huvudmålet, så 'r' gör ingenting så länge målet inte är markerat (samma gäller 'd', 'e', 'c'). Det är appens konvention - inte ett fel - men en förvirring när man testar.

FÄLLA 3: /priority filtrerar på 'påbörjade' (started) som default, så ett nytt testmål syns inte där förrän det fått en starttid. och PATCH goals/:id/status kräver status_id - starttid sätts med PATCH goals/:id { started }.

TESTA MED HJÄLP: appen körs i kind-klustret (tilt up --file environments/dev/Tiltfile) på http://localhost:3100 (nuxt) och 4100 (go-api). /auth/dev-login ger en session-cookie när DEV_AUTH_ENABLED=true (satt i dev-deploymenten). Med playwright-core + /usr/bin/chromium kan man logga in, öppna en vy och klicka - och lyssna på 'framenavigated' för att se när en vy hoppar. Använd localhost (inte 127.0.0.1) så att cookien matchar domänen.
