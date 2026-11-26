---
title: Kunskaps-wikin
type: note
summary: Så fungerar den globala kunskaps-wikin, vilka lager den har och hur agenter använder verktyget.
tags: [wiki, minne, agent, verktyg]
created: 2026-10-21
updated: 2026-11-17
status: active
related: [notes/globala-agentinstruktioner-deployas-via-dotfiles, notes/infrastructure-agents-md-uppdaterad-tre-fallor-for-agenter, notes/konventioner, notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin, notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin, notes/utvecklingsflode, summaries/llm-wiki-karpathy]
---

Wikin är **global** och ligger i `~/repos/agents/wiki` (symlänkad som
`~/wiki`). Den är byggd efter Karpathys LLM-wiki-idé: atomiska
markdown-anteckningar, länkade till varandra, som en agent kan söka i och
skriva till - och som överlever både kontextfönster och modellbyten.

## De två reglerna

De ligger i `AGENTS.md` i varje repos rot och i den **globala**
`AGENTS.md` (`~/AGENTS.md`) som gäller i alla repor:

1. **Läs wikin innan du frågar användaren.** Svaret finns ofta redan
   dokumenterat.
2. **Skriv en not i inboxen när något stort händer** - incident, beslut,
   ny feature eller en icke-obvious bugg.

## Global användning

Wikin bor i `agents`-repot men används av alla repor:

- `~/wiki` är en symlänk hit, så sökvägen är densamma överallt.
- `wikictl` är installerat i `~/.local/bin` och hittar wikin automatiskt:
  `--wiki` → `$WIKI_DIR` → `wiki/` i en katalog ovanför → `~/wiki`.
- Anteckningar taggas automatiskt med `projekt:<repo>` utifrån vilket
  repo kommandot körs i. `wikictl search "fel" --project plan` hittar
  rätt.
- `wikictl init <repo>` skriver in wikinreglerna i ett annat repos
  `AGENTS.md`. Idempotent - ett äldre block ersätts om texten ändrats.
- Den globala `AGENTS.md` ligger i dotfiles-repot
  (`dotfiles/pi/.pi/agent/AGENTS.md`) och deployas med stow till
  `~/.pi/agent/AGENTS.md` och `~/AGENTS.md`.

## Verktyget

`src/agent/cmd/wikictl` i plan-repot (paketet `internal/wiki`). Kommandot är
`wikictl` - installerat i `~/.local/bin`:

| Kommando | Gör |
|---|---|
| `list` | Lista alla sidor. `--project` begränsar. |
| `search <fråga>` | Sök i titlar, taggar och brödtext. `--tag`, `--project`. |
| `show <id>` | Visa en sida. |
| `graph` | Visa kunskapsgrafen. `--mermaid` ger bara diagrammet. |
| `inbox` | Visa vad som väntar på ingestion. |
| `note --title ... --tags ... --body ...` | Skriv en not till inboxen. |
| `page --type ... --title ...` | Skriv en sida direkt (note/entity/concept/summary/synthesis). |
| `link --from ... --to ...` | Länka två sidor. `--remove` tar bort länken. |
| `source add --file ...` | Lägg in en immutabel råkälla. `source list` listar dem. |
| `ingest` | Ingestera inboxen. `--dry-run` visar vad som skulle ske. |
| `index` | Bygg om `index.md`, `catalog.md` och `graph.md`. |
| `lint` | Hälsa: trasiga länkar, ödelappar, dubbler, gammal info, ointegrerade råkällor. |
| `init <repo>` | Ge ett annat repo wikinreglerna. |

## Modellen

- Sidor ligger i `wiki/` - `entities/`, `concepts/`, `summaries/` för
  kuraterat innehåll och `notes/` för arbetsnoter från alla repor.
- Id:t är sökvägen utan ändelse, t.ex. `concepts/oidc` och
  `notes/arkitektur`. Länkar pekar på id:t.
- Metadata i YAML-frontmatter: `title`, `type`, `summary`, `tags`,
  `created`, `updated`, `related`, `status`, `sources`.
- Taggar: `projekt:<repo>` sätts automatiskt, plus fria taggar som
  `bug`, `arkitektur`, `session`.
- Länkar skrivs antingen i `related:` eller som `[[id/sida]]` i
  brödtexten - `AllRelated()` slår ihop dem.
- `index.md` är **genererad** från sidornas frontmatter och byggs om av
  `wikictl index` och vid varje ingestion. `catalog.md` och `graph.md`
  är också genererade - skriv inte i dem för hand.
- `log.md` appenderas vid varje ingestion, i formatet
  `## [YYYY-MM-DD] operation | titel`.

## Ingestion

`wikictl ingest`:

1. Inlägg i `inbox/` blir sidor i `wiki/notes/`. Finns det redan en sida
   med samma id blir inlägget ett tillägg under `## Uppdatering <datum>`.
2. Råmaterialet arkiveras i `archive/` med dagens datum.
3. `catalog.md`, `graph.md` och `log.md` byggs om.

Sammanslagningar görs **inte** automatiskt. `ingest` föreslår istället
sidor som delar två eller fler taggar - en agent som förstår innebörden
avgör sedan.

## Råkällor

`raw/` innehåller immutabla källdokument - artiklar, gists, utdrag ur
repo. De ändras aldrig efter att de lagts in. `wikictl source add`
lägger in en fil och skriver frontmatter med `origin` och `added`.
Linten flaggar råkällor som inte är integrerade i någon sida.

## Varför den ligger i agents-repot

- Wikin versionshålls i git - en anteckning om ett beslut lever längre
  än beslutet.
- Den är läsbar utan verktyg: `index.md` och `graph.md` ger en komplett
  överblick över vad som är känt.
- Ingen extern tjänst, ingen API-nyckel, ingen databas. Bara filer.

Se [[notes/utvecklingsflode]] för kommandona och
[[notes/konventioner]] för skrivreglerna. Mönstret bakom designen är
beskrivet i [[summaries/llm-wiki-karpathy]].

## Relaterade

- [[notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin]]
- [[notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin]]
- [[notes/infrastructure-agents-md-uppdaterad-tre-fallor-for-agenter]]
- [[notes/globala-agentinstruktioner-deployas-via-dotfiles]]
- [[summaries/llm-wiki-karpathy]]
- [[notes/konventioner]]
- [[notes/utvecklingsflode]]
