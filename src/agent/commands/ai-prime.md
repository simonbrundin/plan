---
description: Bygg snabbt en korrekt mental modell av kodbasen innan du börjar arbeta med den
---

# Prime

Bygg en snabb och strukturerad förståelse av kodbasen innan du börjar arbeta med den.

Prime är en **utforskningsskill**, inte en implementeringsskill. Målet är att förstå hur repot är uppbyggt, hur de viktigaste delarna hänger ihop, hur projektet körs och testas samt var den relevanta koden finns.

Läs **inte hela kodbasen**. Börja med billig strukturell information och gå sedan stegvis djupare där det ger mest förståelse.

## 1. Etablera repot

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

## 2. Hitta projektinstruktioner

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

Läs alla relevanta instruktionsfiler.

Var särskilt uppmärksam på:

- arkitektur
- build-kommandon
- testkommandon
- utvecklingskommandon
- kodkonventioner
- förbjudna tillvägagångssätt
- regler för specifika kataloger
- kända problem och fallgropar

Mer specifika instruktioner gäller före mer övergripande instruktioner när de står i konflikt.

## 3. Kartlägg repot

Kör:

```bash
git ls-files
```

Hämta radantal för källfiler (använd `cloc` om tillgängligt, annars `wc -l`):

```bash
# Försök med cloc för detaljerad statistik
if command -v cloc &> /dev/null; then
  cloc . --by-file --csv --quiet 2>/dev/null || true
else
  # Fallback till wc -l för vanliga källfiler
  find . -type f \
    \( -name '*.go' -o -name '*.ts' -o -name '*.tsx' -o -name '*.js' -o -name '*.jsx' \
       -o -name '*.py' -o -name '*.rs' -o -name '*.java' -o -name '*.cs' -o -name '*.cpp' \
       -o -name '*.c' -o -name '*.h' -o -name '*.hpp' -o -name '*.swift' -o -name '*.kt' \
       -o -name '*.yaml' -o -name '*.yml' -o -name '*.json' -o -name '*.toml' \
       -o -name '*.md' -o -name '*.sh' -o -name '*.sql' \
    \) \
    -not -path './.git/*' \
    -not -path './node_modules/*' \
    -not -path './vendor/*' \
    -not -path '*/.git/*' \
    -not -path '*/node_modules/*' \
    -not -path '*/vendor/*' \
    -not -path '*/dist/*' \
    -not -path '*/build/*' \
    -not -path '*/target/*' \
    -not -path '*/.next/*' \
    -not -path '*/coverage/*' \
    -not -path '*/__pycache__/*' \
    -not -path '*/.cache/*' \
    -not -path '*/.venv/*' \
    -not -path '*/venv/*' \
    | xargs wc -l 2>/dev/null \
    | sort -rn | head -50
fi
```

Inspektera även den övergripande katalogstrukturen:

```bash
find . -maxdepth 2 -type d \
  -not -path './.git*' \
  -not -path './node_modules*' \
  -not -path './vendor*' | sort
```

Identifiera:

- applikationer
- bibliotek/paket
- tjänster
- infrastruktur
- tester
- scripts
- dokumentation
- konfiguration
- genererad kod
- databasmigreringar
- deployment-konfiguration

Anta inte att katalogstrukturen beskriver arkitekturen. Verifiera viktiga antaganden genom att läsa relevanta filer.

## 4. Identifiera teknik och körmodell

Leta efter projektets viktigaste manifest och konfigurationsfiler.

Exempel:

- `go.mod`
- `package.json`
- `Cargo.toml`
- `pyproject.toml`
- `requirements.txt`
- `pom.xml`
- `build.gradle`
- `*.csproj`
- `Dockerfile`
- `docker-compose.yml`
- Kubernetes-manifest
- Helm charts
- CI-konfiguration

Läs de viktigaste filerna.

Ta reda på:

- huvudsakliga programmeringsspråk
- ramverk
- pakethanterare
- build-system
- runtime
- viktiga externa beroenden
- databaser
- meddelandeköer
- API:er
- deploymentmodell

## 5. Hitta entrypoints

Identifiera var applikationen faktiskt börjar köras.

Leta efter:

- `main`
- bootstrap-kod
- HTTP/server-initiering
- CLI-entrypoints
- workers
- consumers
- schemalagda jobb
- frontendens root
- dependency injection-konfiguration

Exempel:

```bash
find . -type f \
  \( -name 'main.go' -o -name 'main.py' -o -name 'main.ts' \
     -o -name 'index.ts' -o -name 'index.js' \
     -o -name 'App.tsx' -o -name 'Program.cs' \) \
  -not -path './.git/*' \
  -not -path './node_modules/*'
```

Läs relevanta entrypoints.

Målet är att kunna svara på:

> Om jag startar applikationen, var börjar exekveringen och vad händer därefter?

## 6. Bygg en strukturell kodkarta

Läs inte varje källfil.

Identifiera istället viktiga symboler och relationer.

Använd projektets befintliga verktyg om sådana finns, exempelvis:

- language server / LSP
- tree-sitter
- `ctags`
- compiler metadata
- IDE-index

Om `ctags` finns tillgängligt kan exempelvis följande användas:

```bash
ctags -R \
  --exclude=.git \
  --exclude=node_modules \
  --exclude=vendor \
  .
```

Om lämpliga verktyg saknas, identifiera manuellt viktiga:

- paket/moduler
- klasser
- structs
- interfaces
- funktioner
- services
- handlers
- repositories
- controllers
- domänobjekt

Prioritera symboler som är:

1. entrypoints
2. centrala för många andra komponenter
3. interfaces/abstraktioner
4. centrala domäntyper
5. gränser mot infrastruktur
6. publika API:er

Målet är en **karta över systemet**, inte en komplett läsning av implementationen.

## 7. Spåra huvudarkitekturen

Börja vid entrypoints och följ de viktigaste exekveringsvägarna.

För varje större applikation eller tjänst, ta reda på hur flödet går genom systemet.

Exempel:

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

Detta är endast ett exempel. Beskriv den faktiska arkitekturen som finns i repot istället för att tvinga in den i en viss arkitekturmodell.

Läs representativa filer från varje viktig del.

## 8. Undersök konfiguration och integrationsgränser

Identifiera konfiguration som påverkar arkitekturen.

Leta efter:

- miljövariabler
- konfigurationsfiler
- dependency injection
- databasanslutningar
- API-klienter
- autentisering
- auktorisering
- köer
- cache
- externa tjänster
- feature flags
- Kubernetes-konfiguration
- Docker-konfiguration

Ta reda på hur komponenterna kommunicerar.

Var särskilt uppmärksam på gränser där en förändring i en komponent kan påverka andra komponenter.

## 9. Undersök tester och verifiering

Hitta teststrukturen:

```bash
find . -type f \
  \( -name '*_test.go' -o -name '*.test.ts' -o -name '*.spec.ts' \
     -o -name 'test_*.py' -o -path '*/tests/*' \) \
  -not -path './.git/*' \
  -not -path './node_modules/*'
```

Läs representativa tester.

Ta reda på:

- hur testerna är organiserade
- vad som enhetstestas
- vad som integrationstestas
- hur applikationen verifieras
- vilka kommandon som ska köras efter förändringar

Kör inte dyra tester under Prime om det inte behövs för att förstå projektet.

## 10. Använd Git-historiken vid behov

Använd Git-historiken för att förstå arkitektoniska beslut som inte är uppenbara i den nuvarande koden.

Kör:

```bash
git log --oneline -20
```

Om en viss komponent verkar särskilt viktig eller svårbegriplig:

```bash
git log --oneline -- <path>
```

Använd `git blame` selektivt när det behövs.

Försök inte rekonstruera hela projektets historia.

Målet är att kunna svara på:

> Varför ser koden ut så här?

när den frågan inte går att besvara genom att bara läsa den nuvarande koden.

## 11. Identifiera de viktigaste filerna

Skapa en mental prioritering av repots filer.

### Kärna

Filer som definierar arkitekturen, domänen, gränssnitten eller exekveringsflödet.

### Stöd

Filer som implementerar viktig men mer lokal funktionalitet.

### Periferi

Genererad kod, boilerplate, statiska resurser, migreringar, vendor-kod och annat som normalt inte behövs för att förstå arkitekturen.

Lägg inte lika mycket kontextbudget på alla tre kategorier.

## 12. Verifiera din mentala modell

Innan rapporten skrivs, kontrollera att förståelsen är sammanhängande.

Fråga dig:

- Kan jag förklara hur applikationen startar?
- Kan jag förklara de viktigaste komponenterna?
- Kan jag förklara hur data flödar mellan dem?
- Kan jag identifiera de viktigaste domänbegreppen?
- Kan jag identifiera viktiga externa gränser?
- Vet jag hur projektet byggs och testas?
- Vet jag var ny funktionalitet sannolikt ska placeras?
- Vet jag vilka projektspecifika instruktioner som begränsar förändringar?
- Finns det viktiga delar av arkitekturen som fortfarande är oklara?

Om något viktigt fortfarande är oklart: gör **riktad utforskning** av just den delen.

Gör inte om hela kartläggningen.

## 13. Rapportera

Skapa en kort men användbar karta över kodbasen.

### Projekt

- syfte
- huvudsakligt språk/ramverk
- typ av repository
- viktigaste applikationerna/paketen

### Arkitektur

Förklara de viktigaste komponenterna och hur de hänger ihop.

### Exekveringsflöde

Beskriv huvudflödet från entrypoint till viktiga subsystem.

### Viktiga komponenter

Lista de viktigaste modulerna/paketen/klasserna/funktionerna och varför de är viktiga.

### Data och integrationer

Beskriv:

- databaser
- API:er
- köer
- externa tjänster
- viktiga dataflöden

### Build och test

Ange relevanta kommandon för:

- utveckling
- build
- test
- lint/typecheck om det används

### Projektkonventioner

Sammanfatta viktiga instruktioner och konventioner som hittats i projektets dokumentation.

### Viktiga filer

Lista de få filer som ger mest arkitektonisk förståelse.

### Git-kontext

Nämn historik eller tidigare beslut som är relevanta för att förstå dagens arkitektur.

### Osäkerheter

Var tydlig med sådant som inte kunde fastställas.

Påstå aldrig att du förstår något som inte har verifierats.

## Grundprincip

Optimera för **maximal arkitektonisk förståelse per använd token**.

Målet är inte:

> Läs kodbasen.

Målet är:

> Bygg den minsta korrekta mentala modellen som gör att agenten kan navigera och förändra kodbasen intelligent.
