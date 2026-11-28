---
title: Mål och datamodell
description: Så ser databasen ut - mål, hierarki, beroenden, upprepningar, taggar, platser, kalender och agentkörningar.
order: 3
icon: i-lucide-database
updated: 2026-11-28
group: Användaren
---

# Mål och datamodell

All data ligger i PostgreSQL. Det kanoniska schemat ligger i
`environments/base/db/schema.yaml` - en ConfigMap med SQLen under nyckeln
`schema.sql` - och appliceras av Atlas Operator. Det är alltså DDL, ingen
seed. Seeden (statuser och rotmålet) sker i applikationen vid start, i
`internal/database/seed.go`.

## Tabellerna

### Kärnan

| Tabell | Innehåll |
| --- | --- |
| `users` | Användare (`sub` från Zitadel, e-post, namn) |
| `sessions` | Sessioner: token, användare, utgångsdatum |
| `statuses` | Målets status: `define`, `plan`, `execute`, `verify`, `completed` |
| `goals` | Målen - titel, beskrivning, ikon, status, datum, tid, `agent`, `now`, `order` |
| `user_goals` | Koppling mellan användare och mål |
| `status_updates` | Historik över statusbyten |

`goals` innehåller de egenskaper som avgör hur en rad ritas i
gränssnittet:

| Kolumn | Betydelse |
| --- | --- |
| `status_id` | Var i arbetsflödet målet befinner sig |
| `inbox` | Sant = tanke som inte är ett mål än |
| `now` | Sant = ska göras nu |
| `agent` | Sant = målet är markerat för agentarbete |
| `eternal` | Sant = målet har ingen slutdatum, det pågår |
| `due` / `start_date` | Deadline och startdatum |
| `estimated_minutes` | Uppskattad tidsåtgång |
| `recurrence_rule` | RRULE för återkommande mål |
| `recurrence_mode` | `schedule` eller `push` |
| `recurrence_anchor` / `recurrence_next` | Referensdatum och nästa förekomst |
| `order` / `priority_order` | Sortering i mållistan respektive prioritetslistan |
| `completion_kind` | Hur målet avslutades: `deklarerat` eller `verifierat` |
| `approval_mode` | `plan` (huvudregeln) eller `varje_atgard` |

### Relationer

| Tabell | Innehåll |
| --- | --- |
| `goal_relations` | Förälder/undermål (`parent_id`, `child_id`) samt `order` och `weight` |
| `goal_dependencies` | "X måste vara klart innan Y" (`goal_id`, `depends_on_id`) |

Mål kan alltså bäde *innehålla* andra mål (hierarki) och *vara beroende
av* andra mål (flöde). Beroendet påverkar vilka mål som går att göra nu
- ett mål med ett ouppfyllt beroende lyfts inte fram i Gör nu.

### Etiketter och platser

| Tabell | Innehåll |
| --- | --- |
| `tags` | Användarens etiketter - normaliserade till gemener |
| `goal_tags` | Kopplingar mellan mål och etiketter |
| `places` | Platser med lat/long och radie i meter |
| `goal_places` | Kopplingar mellan mål och platser |
| `calendar_event_places` | Kopplingar mellan händelser och platser |
| `users.current_place_id` | Var användaren befinner sig just nu |

Ett mål är slutfört när användaren befinner sig inom radien av någon
av sina platser. Semantiken är **eller** - det räcker med en av
platserna.

### Kalender

| Tabell | Innehåll |
| --- | --- |
| `calendars` | Kalendrar med färg, ikon och standardflagga |
| `calendar_events` | Händelser med `start_at`/`end_at`, tidszon, hela dagar, status |
| `goal_completions` | Slutförda förekomster av återkommande mål |

`start_at` är inklusiv och `end_at` exklusiv, som i SQL. Tidszonen
lagras som en IANA-zon (`Europe/Stockholm`) medan `timestamptz`
kolumnerna bevarar den absoluta tidpunkten. Återkommande händelser
lagras som en serie (`recurrence_rule`) med undantag i
`parent_event_id` + `original_start_at`.

### Fasmaskin och verifiering

| Tabell | Innehåll |
| --- | --- |
| `goal_verifications` | Verifieringsplanen: typ, villkor, styrka - **valfri** |
| `goal_evidence` | Bevisen: text, källa, tidsstämpel |
| `goal_verification_results` | Omdömena: `passed`, `failed`, `pending`, `avslagen` |
| `goal_phase_events` | Fasloggen: från, till, vem, tvingad, orsak |

Verifieringsplanen är valfri. Ett mål utan plan fungerar precis som ett
mål med plan - skillnaden är bara att det då aldrig kan räknas som
"verifierat", bara som "deklarerat".

Ett mål avslutas på två sätt:

- **Deklarerat** - en människa säger att målet är klart.
- **Verifierat** - en verifiering gått igenom och en människa godkänt
  omdömet.

Byter du verifieringsplanen blir tidigare omdömen ogiltiga - de gällde
den gamla planen.

### Agentkörningar

| Tabell | Innehåll |
| --- | --- |
| `agent_runs` | Körningen: mål, fas, status, trigger, profil, modell, kostnad, reserverad kostnad |
| `agent_run_events` | Kronologisk händelsehistorik per körning (JSONB-payload) |
| `agent_approvals` | Godkännanden med beslut, feedback, konsekvens och verifieringsplan |
| `agent_run_usage` | Tokenanvändning per körning, med prislistversion |
| `agent_budgets` | Budgetgränser per period, agent eller alla agenter |
| `api_keys` | API-nycklar: hash, scope (`full` / `agent`), senaste användning |

Tabellerna `agent_runs`, `agent_run_events` och `agent_approvals` ägs av
Plan - inte av agenten. Det är därför gränssnittet kan visa en
pågående körning även när agenttjänsten är nere, och varför beslutet
om ett godkännande finns kvar efter att agenten startat om.

## Kostnad

Priset räknas fram ur en versionsstyrd prislista i
`src/api/internal/pricing`. Körningen rapporterar bara tokenantal, så
kostnaden går att reproducera: samma modell, samma tokenantal och samma
prislistversion ger samma summa.

Budgetar gränsar köningen. En budget gäller en period (`day`, `week`,
`month`) och kan gälla en enskild agent eller alla. Pågående körningar
reserverar kostnad (`reserved_cost_usd`), så att två samtidiga körningar
inte tillsammans passerar gränsen.

## Begränsningar i databasen

| Constraint | Skyddar mot |
| --- | --- |
| `goals_valid_recurrence_mode` | Okänt upprepningsläge |
| `goals_valid_phase` / `goals_valid_status` | Okända faser eller tillstånd |
| `goals_valid_trigger` / `goals_valid_profile` | Okänd trigger eller profil |
| `agent_approvals_valid_status` | Okänt beslutstillstånd |
| `no_self_reference` | Ett mål som är beroende av sig självt |
| `goal_verifications_valid_type` / `goal_verifications_valid_strength` | Okänd verifieringstyp eller styrka |
| `goal_verification_results_valid_outcome` | Okänt omdöme |
| `goals_valid_completion_kind` / `goals_valid_approval_mode` | Okänt sätt att avsluta eller okänt godkännandeläge |
| `calendars_user_default_idx` | Två standardkalendrar per användare |

## Utveckla schemat

1. Ändra SQLen i `environments/base/db/schema.yaml` (nyckeln `schema.sql`).
2. Testa i dev och PR-miljön.
3. Låt Flux applicera den deklarativa schemat i produktion.

Eftersom Atlas applicerar DDL, körs inga `INSERT` från schemat - seeden
ligger i `internal/database/seed.go` istället.
