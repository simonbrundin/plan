---
title: NeuroQuant-kontot låst av för många inloggningsförsök
type: note
tags: [bug, session, neuro, projekt:neuro]
created: 2026-11-02
updated: 2026-11-03
status: active
related: [notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering, notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]
---

## Vad som hände

NeuroQuant-kontot låstes ("Account Locked", 100 misslyckade försök) 2026-11-02.
Rotorsak: lösenordet i .env på pxe var fel (NeuroQuant svarar "Lösenordet är felaktigt").
Cron körde var 5:e minut (288/dygn) och försökte logga in varje gång - 1628 fel i rad
sedan 2026-10-27 23:20.

## Varför det var farligt

Inloggningsförsöken är dyra: varje försök mot ett låst konto kan förlänga låsningen.
Ett trasigt lösenord + aggressivt schema = låst konto och en timmes automatisk låsning
i taget.

## Rotorsak i detalj (2026-11-02)

- Clerk (clerk.neuroquant.ai) svarar 422 form_password_incorrect - formfyllningen är alltså
  korrekt, det är själva lösenordet som avvisas.
- Samma (gamla) lösenord finns både i .env på pxe och i 1Password ("NeuroQuant", ändrad
  för 8 månader sedan) - de är identiska. Alltså har lösenordet ändrats på NeuroQuant-sidan,
  sannolikt 2026-10-27 (precis då monitorn började falla).
- Det går inte att läsa brevet med lösenordsåterställning: kontot är michael.brundinab@gmail.com
  medan SMTP/app-lösenordet bara ger åtkomst till simonbrundin@gmail.com.
- Ingen Google-session finns i workstationens Chromium, så "Fortsätt med Google" kan inte
  användas för att logga in programmatiskt.

## Fixen (2026-11-02)

- cron på pxe: */5 -> */15 (96 körningar/dygn)
- ny run_guard.py: daglig gräns (MAX_RUNS_PER_DAY, default 96), paus efter upprepade fel
  (eskalerande: 60 min * antal fel), längre paus vid låst konto, 24 h paus vid felaktigt
  lösenord. State i runs.json (gitignoreras).
- scraper.py: känner igen "Account Locked" och "Lösenordet är felaktigt", och fyller inte
  i formuläret alls när kontot är låst.
- monitor.py: larmar via e-post (max en gång per dygn) och pausar i stället för att försöka igen.
- ./run.sh --reset nollställer räkningen (t.ex. efter att lösenordet fixats).

## Kvar att göra

NEURO_PASSWORD i .env på pxe är fortfarande fel. Fixa lösenordet (eller nollställ via
"Glömt lösenordet?" på app.neuroquant.ai), kör ./run.sh --reset på pxe - då bör
monitorn fungera igen inom 24 h (eller direkt om guarden nollställs).

## Relaterade

- [[notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]]
- [[notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering]]
