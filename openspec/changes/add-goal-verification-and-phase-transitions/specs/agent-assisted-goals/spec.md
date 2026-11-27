## ADDED Requirements

### Requirement: Agentens åtgärdsprofil följer uppgiften

En agentkörning MUST ha en åtgärdsprofil: `planering`, `research` eller
`arbete`.

- `planering` får prata med användaren, bryta ned målet och föreslå, men
  skriver inte till något annat än Plan.
- `research` får läsa och samla underlag, men skriver inte till något annat
  än Plan.
- `arbete` får arbeta i en isolerad kopia och föreslå ändringar via branch och
  PR.

Profilen MUST föreslås av agenten i definitionen och MUST godkännas av
användaren innan körningen startar. En körning får aldrig ha en vidare profil
än den godkända.

Målets ämne - kod, hälsa, ekonomi - är en tagg, inte en typ. En etikett för
målets ämne eller område MUST kunna sättas utan att påverka agentens rättigheter.

#### Scenario: Standardprofil

- **WHEN** en definition föreslås utan profil
- **THEN** profilen är `planering`
- **AND** körningen får inte klona repo, skriva filer eller använda
  git-hemlighet

#### Scenario: Koduppdrag föreslår arbetsprofil

- **WHEN** agenten föreslår en definition för ett mål som kräver en
  kodändring
- **THEN** förslaget innehåller profilen `arbete`
- **AND** profilen är synlig i godkännandet

#### Scenario: Godkänd profil begränsar körningen

- **WHEN** en körning startar med profilen `research`
- **THEN** jobbet har ingen repo-klon och ingen git-hemlighet
- **AND** en skrivning utanför Plan avvisas av profilens policy

#### Scenario: Ämnet är en tagg

- **WHEN** användaren sätter taggen "kod" på ett mål
- **THEN** taggen syns på målet
- **AND** agentens åtgärdsprofil påverkas inte av taggen

### Requirement: Godkännande före agenten agerar

Agenten MUST inte agera på ett mål förrän användaren godkänt agentens förslag
för det målet.

Ett godkänt förslag ger agenten rätt att arbeta inom förslaget - skapa
undermål, justera datum, uppskatta tid och föreslå verifieringar.

Åtgärder som lämnar användarens dator - att köpa, boka, skicka eller ringa -
MUST kräva ett eget godkännande varje gång, oavsett målets inställning.

Varje mål får en `approval_mode` som är `plan` eller `varje_atgard`. I läget
`varje_atgard` kräver varje enskild åtgärd ett eget godkännande.

#### Scenario: Standardläge

- **WHEN** ett mål markeras för agentarbete utan angivet läge
- **THEN** målets läge blir `plan`
- **AND** gränssnittet visar att förslaget ska godkännas

#### Scenario: Arbete inom en godkänd plan

- **WHEN** agenten jobbar på ett mål med läge `plan` och ett godkänt förslag
- **THEN** agenten får skapa undermål och justera datum utan nytt godkännande
- **AND** agenten får inte köpa, boka eller skicka utan godkännande

#### Scenario: Varje åtgärd godkänns

- **WHEN** målets läge är `varje_atgard` och agenten föreslår att justera ett
  datum
- **THEN** ett godkännande skapas för just den åtgärden

#### Scenario: Utanförstående åtgärd kräver godkännande

- **WHEN** agenten föreslår att köpa något för ett mål med läge `plan`
- **THEN** ett godkännande skapas
- **AND** åtgärden utförs inte förrän en människa godkänt det

### Requirement: Enhetlig målterminologi

Allt agenten skapar för ett mål MUST vara mål. Agentens nedbrytning skapar
undermål genom målrelationer, inte en separat checkpoint-modell.

Ett undermål är den minsta enheten agenten arbetar med och har samma fält
som ett mål - datum, uppskattad tid, ordning, status och verifiering.

#### Scenario: Agentens plan blir ett träd av mål

- **WHEN** agenten bryter ned ett mål i steg
- **THEN** stegen skapas som undermål i målrelationen
- **AND** inga andra enheter skapas i målmodellen

#### Scenario: Undermål verifieras på samma sätt

- **WHEN** ett undermål skapas av agenten
- **THEN** det har samma fält och samma fasflöde som ett mål på toppnivå
