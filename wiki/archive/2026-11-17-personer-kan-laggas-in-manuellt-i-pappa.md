---
title: Personer kan läggas in manuellt i pappa
type: note
tags: [pappa, feature, databas, gdpr, rekrytering, projekt:pappa]
---

Masterdatabasen fylls normalt från YAML-importen, men ibland vet vi om någon innan hen finns i någon fil - en kandidat från ett samtal, en namnkunnig som tipsar om en kollega. Nu finns knappen '＋ Lägg till person' på /personer (2026-11-17).

BESLUT:
- Endast namnet är obligatoriskt. Titel, organisation, ort, senioritet, kontaktuppgifter, LinkedIn, kompetenser och anteckningar fylls i när vi vet mer - en halvfärdig person är bättre än ingen.
- Skrivning kräver adminsession, precis som parkering och dubblettsammanslagning.
- Dubblettkoll FÖRE insparning: ett namn som nästan stämmer (similarity >= 0.92) avvisas med 409, och appen erbjuder att öppna personen som redan finns. Vi skapar hellre inga dubbletter än att låta dubblettpanelen hantera dem.
- Organisationen återanvänds om den finns (normalized_name är unik), annars skapas den.
- En aktuell anställningsrad läggs in så att CV:t inte är tomt - konfidens 'medium', eftersom uppgifterna kommer från en människa och inte från en källa.
- Källan blir manual://person/<id> med kind 'manual', och person.extra får added_manually = true - så går det att se vilka personer som lagts in manuellt (viktigt för GDPR-spårbarhet: var kommer uppgifterna ifrån).

VAR: API:t är POST /api/people (src/nuxt/server/api/people.post.ts), reningen av inmatningen ligger i src/nuxt/shared/utils/candidate.ts så att samma regler gäller för webbläsaren och CLI:er. Dialogen är AddPersonDialog.vue.
