# Kunskaps-wikin - schema

> Det här är **lagret 3** i Karpathys LLM-wiki-modell: schemat som gör
> wikin till en disciplinerad kunskapsgraf istället för en generisk
> chatthistorik. Modellen i sin helhet är sammanfattad i
> [[summaries/llm-wiki-karpathy]].

## Vad wikin är

En **persistent, sammanhängande kunskapsgraf** som underhålls av
agenterna. Den är global och delas av alla repor.

Skillnaden mot RAG: kunskapen härleds inte om vid varje fråga. Den
kompileras en gång, hålls aktuell, och ackumuleras. Varje ny källa
eller fråga som ger ett värdefullt svar lämnar wikin rikare än den
fann.

## De tre lagren

| Lager | Katalog | Innehåll | Ägare |
|-------|---------|----------|-------|
| **Råkällor** | `raw/` | Immutabla källdokument. Läses, skrivs aldrig. | Människan |
| **Wikin** | `wiki/` | Genererade markdown-sidor: entities, concepts, summaries, notes. | Agenten |
| **Schemat** | `AGENTS.md` (denna fil) | Struktur, konventioner, workflows. | Människa + agent |

Flödet: **råkälla → summary → entity/concept-sidor → index → logg**.

## Mål

1. **Kunskap överlever** sessioner, modellbyten och kontextfönster.
2. **Varje sida är spårbar** till sin källa (`sources:`) eller till
   händelsen som skapade den (`notes/`).
3. **Ingenting blir en ödelapp.** Varje sida har minst en inkommande
   länk.
4. **Föråldrad kunskap uppdateras**, inte lämnas liggande. Ett påstående
   som en ny källa motsäger ska ändras eller markeras `status: deprecated`.
5. **Läsaren hittar rätt** med `index.md` och `wikictl search`.

## Sidtyper

Alla sidor ligger i `wiki/` och har YAML-frontmatter med `type`.

### `entity` - `wiki/entities/`

Något som finns: en tjänst, ett system, en plattform, en nod.

```markdown
---
type: entity
summary: Enradig beskrivning - visas i indexet.
tags: [tagg, tagg]
sources: [raw/2026-01-01-kalla.md]
created: YYYY-MM-DD
updated: YYYY-MM-DD
---

# Namn

## Sammanfattning

[Vad är detta? 2-3 meningar]

## Detaljer

[Fördjupning]

## Relaterade

- [[id/en-annan-sida]]
```

### `concept` - `wiki/concepts/`

Ett begrepp, en idé, en pattern, en fälla.

```markdown
---
type: concept
summary: ...
tags: [...]
sources: [...]
created: YYYY-MM-DD
updated: YYYY-MM-DD
---

# Konceptnamn

## Definition

[Vad betyder det?]

## Nyckelpunkter

- ...

## Kopplingar

- [[id/relaterad-sida]]
```

### `summary` - `wiki/summaries/`

Syntesen av en råkälla. En källa - en summary.

```markdown
---
type: summary
summary: ...
tags: [...]
sources: [raw/2026-01-01-kalla.md]
created: YYYY-MM-DD
updated: YYYY-MM-DD
---

# Sammanfattning: Titel

## Källa

[Vilken källa, varifrån]

## Huvudpunkter

1. ...

## Kopplingar

- [[id/annan-sida]]
```

### `note` - `wiki/notes/`

Arbetsnoter. Korta, situationsbundna, ofta om ett specifikt problem.
Noter taggas automatiskt med `projekt:<repo>`.

```markdown
---
type: note
title: Kort rubrik
summary: ...
tags: [bug, session, projekt:typetrainer]
created: YYYY-MM-DD
updated: YYYY-MM-DD
status: active
related: [notes/annan-not]
---

# Kort rubrik

Vad som hände, varför, och vad som gäller framöver.
```

### `synthesis` - `wiki/synthesis/`

Större sammanhängande sammanställning över flera källor. Används sparsamt.

## Frontmatter-regler

Varje sida **måste** ha:

| Fält | Krav |
|------|------|
| `type` | `entity` \| `concept` \| `summary` \| `note` \| `synthesis` |
| `summary` | En rad. Visas i indexet - gör den begriplig utan att sidan öppnas. |
| `tags` | Lista. Inkl. `projekt:<repo>` för noter. |
| `created` | ISO-datum. |
| `updated` | ISO-datum. Uppdateras vid varje ändring. |

Valfritt: `sources` (varifrån kunskapen kommer), `related` (länkar),
`status` (`active` \| `open` \| `deprecated`).

## Namngivning

- Sidor: **kebab-case**. `concepts/oidc.md`, inte `Concepts/OIDC.md`.
- Id:t är sökvägen utan ändelse: `concepts/oidc`, `notes/arkitektur`.
- Länkar skrivs `[[koncept/oidc]]` - alltid med hela id:t, aldrig bara
  rubriken. Det gör länken upplösbar utan att man gissar.
- En ny sida får aldrig samma id som en befintlig. Dubbletter slås ihop
  manuellt - se nedan.

## Operations

### Ingest (ny källa)

```bash
wikictl source add --file /tmp/kalla.md --title "Rubrik" --origin <url>
wikictl page --type summary --title "Rubrik" --sources raw/... --body "..."
wikictl link --from summaries/rubrik --to concepts/den-koncept-sidan
wikictl index
```

Flödet:

1. Lägg källan i `raw/` med `wikictl source add` - **aldrig** skriv i
   `raw/` för hand.
2. Läs källan. Diskutera huvudpunkterna med användaren om den är
   kontroversiell.
3. Skriv en `summary` som refererar källan i `sources:`.
4. Uppdatera eller skapa berörda `entity`/`concept`-sidor. Om en ny
   källa motsäger en befintlig sida, **ändra den** och sätt `updated`.
5. Länka sidorna med `wikictl link` - båda hållen.
6. Bygg om index: `wikictl index`.
7. Logga: `wikictl ingest` appenderar till `log.md`.

En källa kan röra 10-15 sidor. Det är normalt.

### Query (fråga)

1. `wikictl search "<fråga>"` - eller läs `index.md` om frågan är
   bred.
2. Läs de identifierade sidorna **i sin helhet**, inte bara träffraden.
   Det är oftast där "varför" står.
3. Syntetisera ett svar med citat (`[[id]]`).
4. **Om svaret är värdefullt: fila det tillbaka** som en ny sida. Ett bra
   svar som försvinner i chatthistoriken är kunskap som gått förlorad.

### Lint (hälsa)

```bash
wikictl lint
```

Saker att leta efter:

- [ ] Motsägelser mellan sidor
- [ ] Föråldrade påståenden som nyare källor har ersatt
- [ ] Ödelappar - sidor utan inkommande länkar
- [ ] Viktiga koncept som nämns men saknar egen sida
- [ ] Saknade korsreferenser
- [ ] Råkällor som inte är integrerade i någon sida
- [ ] Dubbletter - sidor om samma sak
- [ ] Dubblerade sektioner inom en sida (t.ex. "Kopplingar" två gånger)
- [ ] Avklippt brödtext - en sektion som slutar mitt i en mening

`wikictl lint` hittar trasiga länkar, ödelappar, dubblerade sidor och
sektioner, sidor utan sammanfattning, föråldrade sidor och
ointegrerade råkällor. Använd `--stale-days N` för att dessutom
flagga sidor som inte uppdaterats på N dagar.

## Dubbletter

Slås ihop manuellt, aldrig automatiskt:

1. Välj den sidan som har bäst innehåll - oftast den med fler källor.
2. Merga innehållet, komplettera frontmatter.
3. Skriv om alla länkar som pekar på den gamla id:n.
4. Ta bort den gamla filen.
5. `wikictl index` och `wikictl lint`.

## Navigering

| Fil | Typ | Innehåll |
|-----|-----|----------|
| `index.md` | Genererad | Katalog över allt, per sidtyp. **Läs den först.** |
| `catalog.md` | Genererad | Fullständig lista med metadata. |
| `graph.md` | Genererad | Länkgraph i Mermaid. |
| `log.md` | Append-only | Kronologi. `## [YYYY-MM-DD] operation \| titel`. |

Ändra inte i de genererade filerna - de skrivs över av
`wikictl index`.

## Verktyget: `wikictl`

Källa: `src/agent/cmd/wikictl` (paketet `internal/wiki`).
Installerat i `~/.local/bin`.

```bash
wikictl list                        # alla sidor
wikictl search "flux"               # sök i titlar, taggar, brödtext
wikictl search "fel" --project plan # sök inom ett projekt
wikictl show concepts/oidc          # visa en sida
wikictl graph --mermaid             # länkgraphen

wikictl inbox                       # vad väntar på att ingesteras?
wikictl note --title "..." --body "..."          # not till inboxen
wikictl page --type concept --title "..."        # skriv en sida
wikictl link --from concepts/a --to concepts/b    # korslänka
wikictl source add --file ...                    # råkälla
wikictl source list                              # råkällor + status

wikictl ingest --dry-run           # se vad som skulle hända
wikictl ingest                     # gör det
wikictl index                      # bygg om index/catalog/graph
wikictl lint                       # hälsa
wikictl init /path/to/repo         # ge ett annat repo wikinreglerna
```

`wikictl` hittar wikin automatiskt: `--wiki` → `$WIKI_DIR` → `wiki/`
i en katalog ovanför → `~/wiki`.

## Arbetsflöde för agenter

**Innan du frågar användaren** - sök i wikin. Svaret finns ofta redan
dokumenterat.

**Under arbetet** - fånga det som är värdefullt. Det som är svårt att
härleda ur koden: incidenter, beslut, icke-obvious buggar, fällor,
förklaringar till varför något ser konstigt ut.

**Efteråt** - skriv en not per idé:

```bash
wikictl note --title "Rubrik" --tags "bug, session" --body "Vad som hände, varför, och vad som gäller framöver."
wikictl ingest
```

En idé per not. Två saker = två noter.

## Regler

1. **Skriv aldrig direkt i `raw/`** - råkällor är immutable.
2. **Frontmatter på allt** - utan `type` och `summary` blir wikin
   osökbar.
3. **Länka alltid** - `[[id/sida]]` i `related:` och i brödtexten.
4. **Datum på allt** - `created` och `updated`.
5. **Logga allt** - varje operation dokumenteras i `log.md`.
6. **Syntetisera, summera inte** - wikins värde är i syntesen, inte i
   repetitionen.
7. **En idé per not.**
8. **Länka hellre än att duplicera.**
9. **Inga hemligheter** - inga credentials, tokens eller personuppgifter.
   Noter och commits versionshålls.
10. **Skriv varför, inte bara vad** - koden visar vad, noten visar varför.

## Filstruktur

```
wiki/
├── raw/                    # Immutabla råkällor (genereras av wikictl source add)
├── wiki/                   # Genererade sidor
│   ├── entities/           # Entiteter (system, tjänster, plattformar)
│   ├── concepts/           # Begrepp, patterns, fällor
│   ├── summaries/          # Synteser av råkällor
│   ├── synthesis/          # Större sammanställningar
│   └── notes/              # Arbetsnoter från alla repor
├── inbox/                  # Inkorg - agenter skriver härifrån
├── archive/                # Processat inbox-material
├── sandbox/                # Utkast och exploration
├── index.md                # GENERERAD - masterkatalog, läs först
├── catalog.md              # GENERERAD - fullständig lista
├── graph.md                # GENERERAD - länkgraph
├── log.md                  # Append-only operationslogg
└── AGENTS.md               # Detta schema
```

## Global användning

Wikin bor i `agents`-repot men används av alla repor. `~/wiki` är en
symlänk hit.

- `wikictl init <repo>` skriver in wikinreglerna i ett annat repos
  `AGENTS.md`.
- Den globala `AGENTS.md` ligger i dotfiles-repot och deployas med stow
  till `~/.pi/agent/AGENTS.md` och `~/AGENTS.md`.
