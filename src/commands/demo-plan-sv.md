# Bug-planering

Skapa en ny plan för att lösa `Buggen` med hjälp av det exakt angivna
markdown-formatet `Planformat`. Följ `Instruktionerna` för att skapa planen och
använd `Relevanta filer` för att fokusera på rätt filer.

## Variabler

issue_number: $1 adw_id: $2 issue_json: $3

## Instruktioner

- VIKTIGT: Du skriver en plan för att lösa en bugg baserat på `Buggen` som kommer
  att tillföra värde till applikationen.
- VIKTIGT: `Buggen` beskriver buggen som ska lösas, men kom ihåg att vi inte
  löser buggen — vi skapar planen som kommer att användas för att lösa buggen
  baserat på `Planformatet` nedan.
- Du skriver en plan för att lösa en bugg, den ska vara grundlig och precis så
  att vi åtgärdar grundorsaken och förhindrar regressioner.
- Skapa planen i `specs/`-katalogen med filnamnet:
  `issue-{issue_number}-adw-{adw_id}-sdlc_planner_{deskriptivt-namn}.md`
  - Ersätt `{deskriptivt-namn}` med ett kort, beskrivande namn baserat på buggen
    (t.ex. "fix-login-fel", "lösa-timeout", "lappa-minnesläcka")
- Använd planformatet nedan för att skapa planen.
- Undersök kodbasen för att förstå buggen, återskapa den och sammanställ en plan
  för att fixa den.
- VIKTIGT: Ersätt varje <placeholder> i `Planformatet` med det begärda värdet.
  Lägg till så mycket detaljer som behövs för att åtgärda buggen.
- Använd din resonemangsmodell: TÄNK NOGGRANT om buggen, dess grundorsak och
  stegen för att åtgärda den ordentligt.
- VIKTIGT: Var kirurgisk med din buggfix, lös den aktuella buggen och fall inte
  av spåret.
- VIKTIGT: Vi vill ha minimalt antal ändringar som kommer att lösa och adressera
  buggen.
- Använd inte dekoratörer. Håll det enkelt.
- Om du behöver ett nytt bibliotek, använd `uv add` och se till att rapportera det
  i `Notes`-avsnittet i `Planformatet`.
- VIKTIGT: Om buggen påverkar användargränssnittet eller användarinteraktioner:
  - Lägg till en uppgift i avsnittet `Steg-för-steg-uppgifter` för att skapa en
    separat E2E-testfil i `.claude/commands/e2e/test_<deskriptivt_namn>.md` baserat
    på exempel i den katalogen
  - Lägg till E2E-testvalidering i ditt Valideringskommandon-avsnitt
  - VIKTIGT: När du fyller i `Planformat: Relevanta filer`-avsnittet, lägg till
    en instruktion att läsa `.claude/commands/test_e2e.md` och
    `.claude/commands/e2e/test_basic_query.md` för att förstå hur man skapar en
    E2E-testfil. Lista din nya E2E-testfil i avsnittet `Planformat: Nya filer`.
  - För att vara tydlig: vi skapar inte en ny E2E-testfil, vi skapar en uppgift
    att skapa en ny E2E-testfil i `Planformatet` nedan
- Respektera begärda filer i avsnittet `Relevanta filer`.
- Börja din undersökning med att läsa `README.md`-filen.

## Relevanta filer

Fokusera på följande filer:

- `README.md` - Innehåller projektöversikt och instruktioner.
- `app/server/**` - Innehåller kodbasens server.
- `app/client/**` - Innehåller kodbasens klient.
- `scripts/**` - Innehåller skript för att starta och stoppa servern + klienten.
- `adws/**` - Innehåller AI Developer Workflow (ADW)-skript.

- Läs `.claude/commands/conditional_docs.md` för att kontrollera om din uppgift
  kräver ytterligare dokumentation
- Om din uppgift matchar något av de villkor som listas, inkludera dessa
  dokumentationsfiler i `Planformat: Relevanta filer`-avsnittet i din plan

Ignorera alla andra filer i kodbasen.

## Planformat

```md
# Bug: <buggnamn>

## Metadata

issue_number: `{issue_number}` adw_id: `{adw_id}` issue_json: `{issue_json}`

## Buggbeskrivning

<beskriv buggen i detalj, inklusive symtom och förväntat vs faktiskt beteende>

## Problembeskrivning

<definiera tydligt det specifika problemet som behöver lösas>

## Lösningsbeskrivning

<beskriv det föreslagna lösningssättet för att åtgärda buggen>

## Steg för att återskapa

<lista exakta steg för att återskapa buggen>

## Grundorsaksanalys

<analysera och förklara grundorsaken till buggen>

## Relevanta filer

Använd dessa filer för att åtgärda buggen:

<hitta och lista filerna som är relevanta för buggen och beskriv varför de är
relevanta i punktlistor. Om det finns nya filer som behöver skapas för att
åtgärda buggen, lista dem i ett h3-avsnitt 'Nya filer'.>

## Steg-för-steg-uppgifter

VIKTIGT: Utför varje steg i ordning, uppifrån och ner.

<lista steg-för-steg-uppgifter som h3-rubriker plus punktlistor. Använd så många
h3-rubriker som behövs för att åtgärda buggen. Ordningen spelar roll, börja med
de grundläggande delade ändringarna som krävs för att åtgärda buggen och gå
sedan vidare till de specifika ändringarna som krävs för att åtgärda buggen.
Inkludera tester som kommer att validera att buggen är åtgärdad utan
regressioner.>

<Om buggen påverkar användargränssnittet, inkludera en uppgift för att skapa en
E2E-testfil. Din uppgift ska se ut så här: "Läs `.claude/commands/e2e/test_basic_query.md`
och `.claude/commands/e2e/test_complex_query.md` och skapa en ny E2E-testfil i
`.claude/commands/e2e/test_<deskriptivt_namn>.md` som validerar att buggen är
åtgärdad, var specifik med stegen för att bevisa att buggen är åtgärdad. Vi vill
ha det minimala antalet steg för att validera att buggen är åtgärdad och
skärmdumpar för att bevisa det om möjligt.">

<Ditt sista steg ska vara att köra `Valideringskommandon` för att validera att
buggen är åtgärdad utan regressioner.>

## Valideringskommandon

Utför varje kommando för att validera att buggen är åtgärdad utan regressioner.

<lista kommandon du kommer att använda för att validera med 100% säkerhet att
buggen är åtgärdad utan regressioner. Varje kommando måste köras utan fel så var
specifik om vad du vill köra för att validera att buggen är åtgärdad utan
regressioner. Inkludera kommandon för att återskapa buggen före och efter
fixen.>

<Om du skapade ett E2E-test, inkludera följande valideringssteg: "Läs
.claude/commands/test*e2e.md`, läs sedan och kör din nya E2E
`.claude/commands/e2e/test*<deskriptivt_namn>.md`-testfil för att validera att
denna funktionalitet fungerar.">

- `cd app/server && uv run pytest` - Kör servertester för att validera att buggen
  är åtgärdad utan regressioner
- `cd app/client && bun tsc --noEmit` - Kör frontend-tester för att validera att
  buggen är åtgärdad utan regressioner
- `cd app/client && bun run build` - Kör frontend-bygge för att validera att buggen
  är åtgärdad utan regressioner

## Anteckningar

<valfritt lista ytterligare anteckningar eller kontext som är relevant för buggen
som kommer att vara hjälpsamt för utvecklaren>
```

## Bugg

Extrahera buggdetaljerna från `issue_json`-variabeln (parsa JSON:en och använd
title- och body-fälten).

## Rapport

- VIKTIGT: Returnera uteslutande sökvägen till den skapade planfilen och inget
  annat.
