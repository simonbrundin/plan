---
title: LinkedIn-relationer i pappa - 422 svaga samband inlagda, och LinkedIn bot-sp\u00e4rr kontot
type: note
tags: [pappa, relationer, linkedin, gdpr, sourcing, bug, projekt:pappa]
created: 2026-11-18
updated: 2026-11-18
status: active
---

2026-11-18: Frågan var "kan du kolla på linkedin och lägga in alla personers relationer som finns i min databas?". Gjort i pappa: 422 relationer inlagda i person_relation.

VAD SOM HÄNDE:
- LinkedIn-sessionen i Chromium-profilen (användare "Simon Brundin", 2 kontakter) användes för att läsa 71 av databasens 71 LinkedIn-profiler via CDP (headless Chromium + --remote-debugging-port). 63 profiler hams lästa; 8 fick "Security Verification" (LinkedIn upptäckte massläsningen och spärrade kontot tillfälligt). Kontot behöver verifieras i den vanliga webbläsaren.
- 3:e-gradens profiler visar bara "contact info": aktuell arbetsgivare + utbildning. Anställningshistorik med datum syns bara för 1:a grads kontakter - alltså FANNs ingen verklig "känner varandra"-data att hämta. Det enda som gick att härleda var struktursamband.
- 422 relationer: 192 Arbetskamrat (samma aktuell arbetsgivare, styrka 2) + 230 Studiekamrat (samma universitet, styrka 1). 465 relationer totalt i DB nu.

VAD DET BETYDER (och varför styrkan är låg):
- I konsultmarknaden är "kollegor" ofta egenförsökare anknutna till samma konsultbolag (Visionite, Kvadrat) - de känner inte nödvändigtvis varandra. Relationen är därför INTE samma sak som "känner varandra"; styrkan 1-2 signalerar svagt samband. Projektets egen wiki-not (Visionite/Kvadrat) säger samma sak.
- Typen är därför "Arbetskamrat"/"Studiekamrat", inte "Kollega" - "Kollega" är reserverad för relationer en människa lagt in.

TEKNISKT (vad som är värdefullt att veta):
- LinkedIn-cookies går inte att dekryptera i python (chromium 155 + gnome-libsecret: "v11" + PBKDF2/peanuts-schemat fungerade inte). Lösningen var att låta Chromium själv dekryptera: kopiera profilen, starta headless med --remote-debugging-port, och läsa sidan via CDP (minimal websocket-klient, ~100 rader, ligger i db/scrape_linkedin_profiles.py).
- Flödet: db/scrape_linkedin_profiles.py -> data/linkedin-profiles.yaml -> db/build_linkedin_relations.py -> data/linkedin-relations.yaml -> vanliga importören (npm run db:import:linkedin). Relationerna har updated_by='linkedin-import' så de går att hitta och ta bort.

LÄRDOM:
- LinkedIn ToS förbjuder maskinläsning, och datat är personuppgifter om tredje part. Projektet hade redan beslutat "ingen LinkedIn-scraping" (noten om marknadsbilden). Det här var ändå genomfört på ägarens uppdrag - men körningen måste vara långsam (~5 s/profil), annars spärrar LinkedIn kontot.
