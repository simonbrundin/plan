---
name: prime
description: |
  Kartlägger snabbt en okänd kodbas innan arbete påbörjas. Används i början
  av varje nytt projekt eller när agenten pekas mot ett nytt repository.

  Utlösare: "prime", "utforska", "kartlägg", "förstå projektet", "vad är det här för projekt", "explore", "understand the codebase"
---

# Prime Skill

Bygg en snabb och strukturerad förståelse av kodbasen innan du börjar arbeta med den.

Prime är en **utforskningsprotokoll**, inte en implementationsprotokoll. Målet är att förstå hur repot är uppbyggt, hur de viktigaste delarna hänger ihop, hur projektet körs och testas samt var den relevanta koden finns.

## Protokoll

### Steg 1: Etablera repot

Kör:

```bash
pwd
git rev-parse --show-toplevel 2>/dev/null || true
git status --short
git branch --show-current
```

Ta reda på:
- repots rot
- aktuell arbetskatalog
- aktuell branch
- om det finns lokala ändringar
- om det är ett monorepo eller innehåller flera applikationer/paket

### Steg 2: Hitta projektinstruktioner

Sök efter instruktioner:

```bash
find . \
  \( -name AGENTS.md -o -name AGENTS.override.md \
     -o -name CLAUDE.md -o -name CLAUDE.local.md \
     -o -name .cursorrules \
     -o -path '*/.cursor/rules/*' \
     -o -name copilot-instructions.md \
  \) \
  -not -path './.git/*' \
  -not -path './node_modules/*' \
  -not -path './vendor/*'
```

Läs alla relevanta instruktionsfiler. Var särskilt uppmärksam på:
- arkitektur
- build-kommandon
- testkommandon
- kodkonventioner
- förbjudna tillvägagångssätt

Mer specifika instruktioner gäller före mer övergripande instruktioner vid konflikt.

### Steg 3: Kartlägg repot

```bash
git ls-files | head -100
```

Identifiera:
- applikationer
- bibliotek/paket
- tjänster
- infrastruktur
- tester
- scripts
- dokumentation

### Steg 4: Identifiera teknik och körmodell

Leta efter:
- `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`
- `Dockerfile`, `docker-compose.yml`
- Kubernetes-manifest, Helm charts, CI-konfiguration

Ta reda på:
- huvudsakliga programmeringsspråk
- ramverk
- build-system
- runtime
- viktiga externa beroenden
- databaser

### Steg 5: Hitta entrypoints

```bash
find . -type f \
  \( -name 'main.go' -o -name 'main.py' -o -name 'main.ts' \
     -o -name 'index.ts' -o -name 'index.js' \
     -o -name 'App.tsx' -o -name 'Program.cs' \) \
  -not -path './.git/*' -not -path './node_modules/*'
```

Målet är att kunna svara på:
> Om jag startar applikationen, var börjar exekveringen och vad händer därefter?

### Steg 6: Bygg en strukturell kodkarta

Identifiera viktiga:
- paket/moduler
- klasser, structs, interfaces
- funktioner, services, handlers
- domänobjekt

Prioritera symboler som är:
1. entrypoints
2. centrala för många andra komponenter
3. interfaces/abstraktioner
4. centrala domäntyper
5. gränser mot infrastruktur

Målet är en **karta över systemet**, inte en komplett läsning av implementationen.

### Steg 7: Spåra huvudarkitekturen

Börja vid entrypoints och följ de viktigaste exekveringsvägarna.

```text
entrypoint
    ↓
initiering
    ↓
transport / interface
    ↓
applikationslogik
    ↓
domän
    ↓
infrastruktur
    ↓
databas / externa system
```

Beskriv den faktiska arkitekturen som finns i repot — tvinga inte in den i en modell.

### Steg 8: Undersök konfiguration och integrationsgränser

- miljövariabler
- databasanslutningar
- API-klienter
- autentisering
- externa tjänster

### Steg 9: Undersök tester och verifiering

```bash
find . -type f \
  \( -name '*_test.go' -o -name '*.test.ts' -o -name '*.spec.ts' \
     -o -name 'test_*.py' -o -path '*/tests/*' \) \
  -not -path './.git/*' -not -path './node_modules/*'
```

Ta reda på:
- hur testerna är organiserade
- vilka kommandon som ska köras efter förändringar

### Steg 10: Använd Git-historiken vid behov

```bash
git log --oneline -20
```

Använd selektivt för att förstå arkitektoniska beslut som inte är uppenbara.

### Steg 11–12: Verifiera och rapportera

Kontrollera innan rapporten skrivs:
- Kan jag förklara hur applikationen startar?
- Kan jag förklara de viktigaste komponenterna?
- Kan jag förklara hur data flödar mellan dem?
- Vet jag hur projektet byggs och testas?
- Finns det viktiga delar som fortfarande är oklara?

## Rapportformat

### Projekt
- syfte
- huvudsakligt språk/ramverk
- typ av repository
- viktigaste applikationerna/paketen

### Arkitektur
De viktigaste komponenterna och hur de hänger ihop.

### Exekveringsflöde
Huvudflödet från entrypoint till viktiga subsystem.

### Viktiga komponenter
De viktigaste modulerna/paketen/klasserna och varför de är viktiga.

### Data och integrationer
- databaser
- API:er
- externa tjänster

### Build och test
Relevanta kommandon för utveckling, build och test.

### Projektkonventioner
Viktiga instruktioner och konventioner.

### Viktiga filer
De få filer som ger mest arkitektonisk förståelse.

### Osäkerheter
Var tydlig med sådant som inte kunde fastställas.

## Grundprincip

Optimera för **maximal arkitektonisk förståelse per använd token**.

Målet är inte "läs kodbasen". Målet är "bygg den minsta korrekta mentala modellen som gör att agenten kan navigera och förändra kodbasen intelligent."

## Autostart-regler

Prime körs **automatiskt** när:
1. Agenten pekas mot ett nytt projekt (nytt working directory)
2. Användaren säger "vad är det här för projekt" eller liknande
3. En konversation startas med ett okänt projekt

Prime körs **ej** när:
1. Samma projekt som i föregående konversation
2. Användaren redan har kört Prime nyligen (cacha resultatet)

## Caching

Spara Prime-rapporten till `/tmp/prime-{project-hash}.md` för att undvika onödig omkartläggning.
