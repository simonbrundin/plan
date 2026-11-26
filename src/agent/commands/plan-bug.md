# Bugplanering

Skapa en ny plan i `.agent/plans/*.md` för att lösa **bugen** med det exakt
angivna markdown-**planformatet**.  
Följ **instruktionerna** för att skapa planen och använd **relevanta filer** för
att fokusera på rätt delar av kodbasen.

## Instruktioner

- Du skriver en plan för att lösa en bugg. Den ska vara **grundlig och precis**
  så att vi åtgärdar **grundorsaken** och förhindrar framtida regressioner.
- Skapa planen i en fil under `specs/*.md`. Ge filen ett bra namn baserat på
  buggen.
- Använd exakt det **planformat** som anges nedan.
- Undersök kodbasen för att förstå buggen, kunna reproducera den och skapa en
  bra lösningsplan.
- **VIKTIGT:** Ersätt alla `<placeholder>` i planformatet med faktiska värden.
  Lägg till så mycket detaljer som behövs för att verkligen lösa buggen.
- Använd din resonemangsmodell: **TÄNK HÅRT** på buggen, dess grundorsak och hur
  man löser den på rätt sätt.
- **VIKTIGT:** Var kirurgisk i din buggfix – lös **exakt** den bugg som
  rapporterats, spåra inte ur.
- **VIKTIGT:** Sträva efter **minsta möjliga antal ändringar** som löser
  problemet helt.
- Använd **inga dekoratorer**. Håll det enkelt.
- Behöver du ett nytt bibliotek? Använd `uv add` och dokumentera det i avsnittet
  **Notes**.
- Respektera de filer som listas under **Relevanta filer**.
- Börja alltid din research genom att läsa `README.md`.

## Relevanta filer

Fokusera enbart på följande filer/kataloger:

- `README.md`  
  → Innehåller projektöversikt och viktiga instruktioner
- Läs in alla filer jälva kodbasen (både klient och server)
- `scripts/**`  
  → Skript för att starta/stoppa server och klient

**Ignorera alla andra filer i kodbasen.**

## Planformat

```markdown
# Bug: <bugnamn>

## Bugbeskrivning

Beskriv buggen i detalj, inklusive symptom och skillnaden mellan förväntat och
faktiskt beteende.

## Problemformulering

Definiera tydligt det specifika problemet som ska lösas.

## Lösningsformulering

Beskriv den föreslagna lösningsmetoden för att åtgärda buggen.

## Steg för att reproducera

1. ...
2. ...
3. ...

## Grundorsaksanalys

Analys och förklaring av varför buggen uppstår.

## Relevanta filer

Använd dessa filer för att åtgärda buggen:

- `sökväg/till/fil.py`  
  → förklaring varför den är relevant
- `annan/viktig/fil.js`  
  → förklaring

### Nya filer (om några behövs)

- `specs/nya-filer-som-skapas.md`  
  → kort beskrivning av syftet

## Steg-för-steg-uppgifter

**VIKTIGT:** Utför stegen i ordning uppifrån och ner.

### 1. Förberedande ändringar / gemensamma grunder

- ...

### 2. ...

- ...

### 3. Tester som validerar fixen

- ...

### 4. Slutlig validering

- Kör alla valideringskommandon nedan

## Valideringskommandon

Kör dessa kommandon för att med 100 % säkerhet bekräfta att buggen är löst och
att inga regressioner uppstått.

- Innan fix: reproducera buggen med ...
- Efter fix:
  - `cd app/server && uv run pytest`
  - `scripts/start-all.sh` && testa manuellt i webbläsaren ...
  - `uv run pytest --lf` (endast de senast misslyckade testerna)

## Anteckningar

- Eventuella ytterligare kommentarer, beslut, saker att tänka på, tekniska
  begränsningar, etc.
```
