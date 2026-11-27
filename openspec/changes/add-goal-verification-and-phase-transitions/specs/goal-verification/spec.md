## ADDED Requirements

### Requirement: Verifieringsplan är valfri

Ett mål MUST kunna ha en verifieringsplan. En verifieringsplan består av en typ, ett
villkor och en styrka.

Typerna är `checklista`, `delmal`, `kalender`, `bevis`, `matvarde` och `kod`.
Styrkan är `stark` eller `svag`.

Ett mål utan verifieringsplan MUST kunna avslutas av en människa utan
hindrande villkor, och MUST kunna föreslås avslutat av agenten.

En verifieringsplan får ändras av användaren när som helst. När den ändras
MUST ett tidigare verifieringsomdöme för målet ogiltigförklaras.

#### Scenario: Mål utan verifieringsplan

- **WHEN** ett mål inte har någon verifieringsplan
- **THEN** målet har ingen verifieringsvy i gränssnittet
- **AND** målet kan avslutas av användaren med en enda handling

#### Scenario: Agenten föreslår en verifieringsplan

- **WHEN** agenten föreslår en definition för ett nytt mål
- **THEN** förslaget innehåller en verifieringsplan med typ, villkor och styrka
- **AND** inget bevis eller omdöme föreslås i definitionen

#### Scenario: Användaren skärper verifieringen

- **WHEN** användaren byter ett måls verifiering från `bevis`/`svag` till
  `kod`/`stark`
- **THEN** verifieringsplanen sparas med den nya typen och styrkan
- **AND** ett tidigare verifieringsomdöme för målet markeras ogiltigt

### Requirement: Bevis sparas på målet

Bevis MUST sparas i Plan och knytas till målet och dess verifiering.
Varje bevis har typ (`text`, `fil`, `varde`, `kalenderhändelse`, `kodstatus`),
ett värde, en tidpunkt och en källa.

Beviset MUST överleva att agenttjänsten är nere, eftersom Plan äger
databasen.

#### Scenario: Användaren lämnar in ett bevis

- **WHEN** användaren skickar ett bevis för mål 449
- **THEN** beviset sparas med typ, värde, tidpunkt och källa
- **AND** beviset är kopplat till målets verifiering

#### Scenario: Beviset finns utan agenttjänsten

- **WHEN** agent-API:t är nere
- **THEN** API:t listar fortfarande bevisen för målet

### Requirement: Verifieringsomdöme

En verifiering MUST ge ett omdöme med utfallet `passed`, `failed` eller
`pending`. Ett omdöme `passed` kräver att alla villkor i verifieringsplanen
är uppfyllda.

Omdömet MUST godkännas av en människa innan det får räknas som verifierat.
Agenten får skapa och revidera omdömen men får inte godkänna dem.

#### Scenario: Verifieringen körs och faller

- **WHEN** en verifiering körs på ett mål vars villkor inte är uppfyllda
- **THEN** omdömet får utfallet `failed`
- **AND** omdömet är inte godkänt och räknas inte som verifiering

#### Scenario: Användaren godkänner ett omdöme

- **WHEN** användaren godkänner ett omdöme med utfallet `passed`
- **THEN** omdömet markeras som godkänt
- **AND** målet får en verifierad status

#### Scenario: Agenten kan inte godkänna ett omdöme

- **WHEN** agenten försöker godkänna ett verifieringsomdöme med agent-scope
- **THEN** API:t svarar med 403
- **AND** omdömet förblir ogodkänt

### Requirement: Verifieringen är ett villkor för agentens avslut

När ett mål har en verifieringsplan MUST agentens förslag att avsluta målet
kräva att planen gått igenom med utfallet `passed`.

Om villkoret inte är uppfyllt MUST agentens förslag avvisas, och API:t MUST
ange vilket villkor som saknas.

Ett mål utan verifieringsplan får föreslås avslutat av agenten. Ett sådant
förslag MUST markeras som ett förslag utan bevis.

#### Scenario: Agentens förslag avvisas när verifieringen inte gått igenom

- **WHEN** agenten försöker avsluta ett mål vars verifiering har utfallet
  `failed`
- **THEN** API:t avvisar förslaget
- **AND** svaret anger att verifieringen inte är godkänd

#### Scenario: Agentens förslag utan verifieringsplan

- **WHEN** agenten föreslår att avsluta ett mål utan verifieringsplan
- **THEN** förslaget skapas
- **AND** förslaget är markerat som utan bevis
