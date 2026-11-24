# Inbox

Här skriver agenter saker som ska in i kunskapsgrafen.

Det är den **enda** katalogen i wikin som agenter skriver i. Allt annat
är kuraterat.

## Regler

- **En idé per fil.** Två saker = två filer.
- **Länka hellre än att duplicera.** Om något redan finns, länka till sidan.
- **Skriv varför, inte bara vad.** Koden visar vad - noten ska visa varför.
- **Inga hemligheter.** Allt här versionshålls i git - inga tokens,
  credentials eller personuppgifter.
- **Skriv på svenska.**

## Format

Markdown med YAML-frontmatter:

```markdown
---
title: Kort rubrik
type: note
tags: [bug, session, projekt:typetrainer]
created: YYYY-MM-DD
updated: YYYY-MM-DD
status: active
related: [concepts/Flux]
---

# Kort rubrik

Vad som hände, varför, och vad som gäller framöver.
```

## Flöde

```bash
wikictl note --title "Rubrik" --tags "bug" --body "Vad som hände och varför."
wikictl ingest --dry-run   # se vad som skulle hända
wikictl ingest              # gör det
```

`ingest` flyttar inläggen till `wiki/notes/`, arkiverar råmaterialet i
`archive/` och bygger om `catalog.md`, `graph.md` och `log.md`.
