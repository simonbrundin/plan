---
allowed-tools: Bash
description:
  Analyserar ändringar, grupperar per Conventional Commits och skapar separata
  commits per kategori
---

# Push

## Syfte

Analyserar git-ändringar, kategoriserar dem per Conventional Commits-typ (feat,
fix, refactor, docs, chore, style), skapar separata commits med lämpliga
meddelanden och pushar till remote.

## Variabler

Inga dynamiska variabler - kommandot hanterar allt internt.

## Workflow

### Steg 1: Analysera ändringar

Kör för att se vilka filer som ändrats:

```bash
git status --short
git diff --name-only
```

### Steg 2: Kategorisera per Conventional Commits-typ

Agenten analyserar filer och grupperar dem:

| Typ          | Filer/Mönster                                                     |
| ------------ | ----------------------------------------------------------------- |
| **feat**     | Nya filer, `.vue`, `.ts`, `.tsx`, nya komponenter, nya funktioner |
| **fix**      | `.py`, `.ts` med "fix", "bug", "error", patch-filer               |
| **refactor** | Omstrukturerade filer, filer med "refactor" i kontexten           |
| **docs**     | `.md`, README, dokumentation                                      |
| **chore**    | Config: `.json`, `.yaml`, `.toml`, `.env`, `.gitignore`           |
| **style**    | `.css`, `.scss`, styling-filer                                    |
| **test**     | Testfiler: `*.test.*`, `*.spec.*`, `test/`, `tests/`              |

### Steg 3: Gruppera och commita per typ

För varje kategori med ändringar:

```bash
git add <filerna>
git commit -m "<typ>: <sammanfattning>"
```

Agenten skapar en beskrivande sammanfattning baserat på filernas ändringar.

### Steg 4: Pusha

Efter alla commits:

```bash
git push
```

### Fallback

Om endast en typ av ändring → gör en enda commit (befintligt beteende).

## Relevanta filer

- `.git/` (vid behov för att förstå konfiguration)

## Rapport

Kommandot ska svara med:

- Sammanfattning av ändringar som hittades (antal filer per kategori)
- Antal commits som skapades
- Commit-meddelanden som användes
- Bekräftelse på att push lyckades
- Varning om inga ändringar hittades
