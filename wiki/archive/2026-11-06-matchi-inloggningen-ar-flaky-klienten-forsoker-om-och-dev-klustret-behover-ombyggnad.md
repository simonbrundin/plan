---
title: MATCHi-inloggningen är flaky - klienten försöker om, och dev-klustret behöver ombyggnad
type: note
tags: [bug, matchi, keycloak, stk, lesson, deploy, projekt:stk]
---

Symtom: efter att frontend-funktionen 'markera tider som är bokade i MATCHi' deployats till dev-klustret syntes inga gröna rader. Två samverkande orsaker:

1) Utanför klustret fungerade allt (lokalt: 24 av 39 bokningar matchade), men i dev-podden misslyckades varje inloggning mot Keycloak med 'kunde inte hitta Keycloak-parametrar i inloggningssidan'. Nätverket i klustret var bevisat OK (curl från en debug-pod på samma nod fick Keycloaks sida med session_code/execution/tab_id).

Fix: inloggningen i matchi.Client körs nu upp till 3 gånger (loginOnce + login) med ökande väntetid, eftersom Keycloak ibland svarar med en sida utan sessionsparametrarna. Detta är samma flakiness som beskrivs i notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass.

2) Dev-klustret kör de images som Tilt byggt. Utan en ombyggnad av images syns inte frontend-ändringar i klustret. Tilt bygger om lokalt till localhost:5000/<image>:tilt-<hash> och sätter image på deploymenten; man kan också bygga manuellt (docker build + push + kubectl set image deploy/<name> <container>=<image>).

Verifiering: mot tilt-portforwarden (nuxt på 127.0.0.1:3005) gav /bookings 95 gröna rader av 150 via playwright.
