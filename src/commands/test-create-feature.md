---
allowed-tools: Write, Edit, Glob, Grep, Read, Bash
description: Skapa ny BDD feature i Gherkin-format
argument-hint: [feature-beskrivning]
---

# Skapa feature

## Syfte

Skapa en ny feature-fil i Gherkin-format i `.features/` eller lämplig undermapp.
Feature-filen ska innehålla välstrukturerade scenarier med
Given-When-Then-struktur som följer BDD-best-practices.

## Variabler

FEATURE: $1 <beskrivning av feature: new-feature skapar en ny feature i
.features/ eller i någon lämplig undermapp i .features/. Den är skriver i
gherkinformat med Given, When, Then och andra nyckelord som känns relevanta.

## Workflow

1. Analysera `$FEATURE` och extrahera:
   - Feature-namn och beskrivning
   - Business value/vad vi vill uppnå
   - Antal och typ av scenarier som behövs
2. Bestäm lämplig undermapp i `.features/` baserat på domän/område
3. Skapa feature-filen med:
   - Feature-titel och beskrivning (Business value)
   - Scenario med Given-When-Then struktur
   - Använd Scenario Outline för varianter
   - Lämpliga taggar (@tag)
4. Spara filen som `<feature-namn>.feature` i lämplig mapp

# Så skriver du bra features och scenarier i Gherkin med BDD

**BDD handlar primärt om samarbete och tydlig kommunikation** –
Gherkin-scenarier är levande dokumentation som alla i teamet (affär, produkt,
utveckling, test) ska kunna läsa och förstå. Automatisering är en bonus, inte
huvudsyftet.

## Tips – bästa praxis

### Grundläggande principer

- **Skriv alltid i affärsspråk** – vardagligt, domänspecifikt språk som
  produktägare, användare och affärsfolk förstår utan att fråga "vad betyder
  det?".
- Håll meningarna så korta som möjligt
- **Fokusera på beteende, inte implementation** – beskriv **vad** som ska hända
  och vad som ska observeras, aldrig **hur** det sker tekniskt.
- **Ett beteende = ett scenario** – bryt upp om det finns flera
  "Then"-påståenden eller flera orelaterade saker som testas samtidigt.
- **Håll Given-When-Then strikt men naturligt**
  - Given: kontext och förutsättningar
  - When: den handling som triggar beteendet (oftast **en** When per scenario)
  - Then: observerbara resultat (oftast **en** Then – flera är ett
    varningstecken)  
    → Flera Given/And är oftast OK.

### Skrivstil & läsbarhet

- Använd **konkreta roller eller personas** istället för "jag" när det ger
  värde  
  Ex: "som inloggad kund", "som lagerpersonal", "som Alice (ny kund)".
- Föredra **deklarativt** språk framför imperativt  
  **Dåligt**: "klickar på knappen 'Lägg i varukorg'"  
  **Bra**: "lägger till produkten 'AirPods Pro' i varukorgen"
- Håll scenarier **korta** (helst 4–12 rader) och fokuserade.
- Undvik UI-specifika detaljer i scenarier som förväntas leva längre än 6–12
  månader.

### Struktur & organisation

- Använd **Scenario Outline** + exempeltabell för variationer av samma
  beteende  
  (max ~10–15 exempel – vid fler → splittra till flera scenarier eller
  omgruppera).
- Använd **Background** sparsamt – endast för gemensam Given som verkligen
  gäller **alla** scenarier i filen.

## Vanliga anti-patterns (undvik dessa!)

1. **Tekniska detaljer i Gherkin**  
   → "API-anropet returnerar 201" / "databasen uppdateras med…"  
   → Bryter affärsspråket och gör scenarierna sköra.

2. **Vaga eller tvetydiga formuleringar**  
   → "systemet fungerar", "något händer", "rätt saker visas"  
   → Kan tolkas på många sätt → leder till diskussioner vid granskning.

3. **Flera beteenden i samma scenario**  
   → flera When-Then-par eller orelaterade Then-satser  
   → Svårt att felsöka, underhålla och förstå.

4. **För mycket eller för lite detaljer**
   - För högnivå → scenariot säger nästan ingenting konkret
   - För detaljerat → blir skört, UI-kopplat och dyrt att underhålla

5. **"Jag"-perspektiv överallt**  
   → Gör scenarierna mindre objektiva och svåra för andra roller att relatera
   till.

6. **Scenario Outline med 30+ exempelrader**  
   → Bättre att splittra till flera tydliga scenarier.

7. **Background som döljer viktig kontext**  
   → Om något är viktigt → skriv det i Given i scenariot istället.

8. **Inget samarbete – QA skriver scenarierna ensamma**  
   → Blir automatiserade testfall istället för gemensamma krav.

9. **Scenarier skrivs efter implementation (inte före)**  
   → Missar hela BDD-discovery-delen (Three Amigos, Example Mapping).

10. **Överdriven UI-detaljering i långlivade scenarier**  
    → "klickar på meny → väljer Administrera → fyller i fält X" = hög
    underhållskostnad.

## Relevanta filer

- `.features/` - katalog för feature-filer
- Befintliga `.feature`-filer för att följa konventioner

## Exempel

```gherkin
Feature: Användarinloggning
  Som registrerad användare
  Vill jag kunna logga in med mina uppgifter
  För att få tillgång till mina personliga sidor

  Scenario: Användare loggar in med giltiga uppgifter
    Given att användaren är registrerad med epost "test@example.com" och lösenord "Hemligt123"
    When användaren loggar in med epost "test@example.com" och lösenord "Hemligt123"
    Then ska användaren se sin dashboard

  Scenario Outline: Inloggning med ogiltiga uppgifter
    Given att användaren är registrerad med epost "test@example.com" och lösenord "Hemligt123"
    When användaren loggar in med epost "<epost>" och lösenord "<lösenord>"
    Then ska ett felmeddelande visas
    And ska användaren inte vara inloggad

    Examples:
      | epost            | lösenord    |
      | fel@example.com  | Hemligt123  |
      | test@example.com  | FelLösenord |
      |                  | Hemligt123  |
```

## Rapport

Presentera den skapade feature-filen med sökväg och en kort sammanfattning av
vilka scenarier som skapades.
