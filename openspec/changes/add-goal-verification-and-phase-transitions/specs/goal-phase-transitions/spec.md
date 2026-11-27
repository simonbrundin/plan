## ADDED Requirements

### Requirement: Fasövergångar styrs av villkor

Plan API:t MUST erbjuda en fasövergång
`POST /api/v1/goals/:id/phase` som flyttar ett mål mellan faserna `define`,
`plan`, `execute`, `verify` och `completed`.

Övergången MUST kontrollera övergångens förutsättningar innan fasen ändras.
Om förutsättningarna inte är uppfyllda MUST API:t svara med 409 och ange
vilket villkor som saknas. Målets fas MUST då vara oförändrad.

En fasövergång MUST loggas i status-historiken med både från- och tillfas,
och med vem eller vad som utförde den.

#### Scenario: Giltig övergång

- **WHEN** ett mål i fasen `define` med en godkänd definition flyttas till
  `plan`
- **THEN** fasen blir `plan`
- **AND** status-historiken får en rad med `define` och `plan`

#### Scenario: Villkor inte uppfyllt

- **WHEN** ett mål i fasen `define` utan godkänd definition försöks flyttas
  till `plan`
- **THEN** API:t svarar med 409
- **AND** svaret anger att definitionen inte är godkänd
- **AND** målets fas ändras inte

#### Scenario: Okänd fas avvisas

- **WHEN** en övergång begärs till fasen `klar`
- **THEN** API:t svarar med 400
- **AND** målets fas ändras inte

### Requirement: Varje övergång har ett angivet villkor

Varje övergång MUST ha exakt ett angivet villkor. Övergångarna och deras villkor är:

- `define` → `plan`: målet har en godkänd agentdefinition.
- `plan` → `execute`: målet har minst ett undermål med datum eller
  uppskattad tid.
- `execute` → `verify`: varje undermål har ett resultat - avslutat,
  uppskjutet eller hoppat.
- `verify` → `completed`: se kravet om avslutande.

En människa får med en explicit override-parameter tvinga en övergång. En
sådan override MUST loggas som ett avvikelseevent med angiven orsak.

#### Scenario: Övergång till execute kräver delmål

- **WHEN** ett mål i fasen `plan` utan undermål försöks flyttas till `execute`
- **THEN** API:t svarar med 409
- **AND** svaret anger att undermål saknas

#### Scenario: Override loggas

- **WHEN** en människa tvingar en övergång med orsaken "vill hinna ikapp"
- **THEN** övergången genomförs
- **AND** ett avvikelseevent med orsaken sparas för målet

### Requirement: Människan kan avsluta ett mål direkt

En människa MUST kunna avsluta ett mål med en enda handling, oavsett vilken
fas målet befinner sig i - även från `define` - och oavsett om målet har en
verifieringsplan.

Ett sådant avslutande MUST registrera hur målet avslutades - `deklarerat` om
ingen verifiering gått igenom, `verifierat` om målet har en verifieringsplan
med ett godkänt omdöme.

Om målet har en verifieringsplan men ingen godkänd verifiering MUST
avslutandet ändå genomföras och markeras som `deklarerat`.

#### Scenario: Enkelt mål avslutas av användaren

- **WHEN** användaren avslutar målet "Byt glödlampa" direkt från fasen `define`
- **THEN** målet blir `completed`
- **AND** avslutandet registreras som `deklarerat`
- **AND** inget verifieringskrav hindrar avslutandet

#### Scenario: Agenten får inte hoppa över faser

- **WHEN** agenten försöker avsluta ett mål som står i fasen `define`
- **THEN** API:t avvisar förfrågan
- **AND** målets fas ändras inte

#### Scenario: Avslutande med godkänd verifiering

- **WHEN** användaren avslutar ett mål med en verifieringsplan och ett godkänt
  omdöme med utfallet `passed`
- **THEN** målet blir `completed`
- **AND** avslutandet registreras som `verifierat`

#### Scenario: Agenten kan inte avsluta utan godkännande

- **WHEN** agenten försöker avsluta ett mål direkt, utan godkänt förslag
- **THEN** API:t avvisar förfrågan
- **AND** målets fas ändras inte

### Requirement: Fasen går inte bakåt utan människa

En fasövergång MUST gå framåt i fasordningen. Att gå bakåt MUST kräva en
människa och MUST loggas som en avvikelse.

Agenten MUST inte kunna flytta ett mål bakåt i fas.

#### Scenario: Agenten försöker gå bakåt

- **WHEN** agenten försöker flytta ett mål från `execute` tillbaka till `plan`
- **THEN** API:t avvisar förfrågan
- **AND** målets fas ändras inte

#### Scenario: Människa öppnar upp målet igen

- **WHEN** användaren flyttar ett mål från `execute` tillbaka till `plan`
- **THEN** fasen blir `plan`
- **AND** övergången syns i status-historiken

### Requirement: Agenten kan inte skriva fas fritt

Målets fas MUST ändras via fasövergången. Agent-scope MUST kunna begära
övergångar, men alla villkor MUST kontrolleras likadant för agent och
människa.

Målets fas MUST inte kunna sättas till `completed` av agenten utan ett
godkänt förslag från människa, och - om målet har en verifieringsplan - utan
att verifieringen gått igenom.

#### Scenario: Agentens statusanrop går via maskinen

- **WHEN** agenten sätter målets status med agent-scope
- **THEN** villkoren för övergången kontrolleras
- **AND** ett mål med en verifieringsplan vars verifiering inte gått igenom
  inte kan bli `completed`
