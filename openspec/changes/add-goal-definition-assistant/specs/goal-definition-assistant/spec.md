## ADDED Requirements

### Requirement: Definitionspanel

Appen MUST ha en definitionspanel där användaren kan se och fylla i hur ett
mål ser ut. Panelen MUST visa målets nuvarande värden och vilka fält som
saknas, och den MUST öppnas från inboxen och från målsidan.

Panelen MUST visa, per fält, målets nuvarande värde och - när ett förslag
finns - föreslaget värde, som `från` och `till`.

Panelen MUST kunna öppnas med tangenten `Space+f` på målsidan.

#### Scenario: Öppna panelen från målsidan

- **WHEN** användaren står på en målsida och trycker `Space` och sedan `f`
- **THEN** definitionspanelen öppnas
- **AND** den visar målets titel, beskrivning, deadline, plats, upprepning,
  tidsåtgång, taggar och verifieringskriterium

#### Scenario: Öppna panelen från inboxen

- **WHEN** användaren trycker på "Bearbeta" på ett mål i inboxen
- **THEN** definitionspanelen öppnas för det målet

#### Scenario: Panelen visar saknade fält

- **GIVEN** ett mål som saknar deadline, plats och verifieringskriterium
- **WHEN** användaren öppnar definitionspanelen
- **THEN** de tre fälten markeras som saknade
- **AND** övriga fält visas med sina nuvarande värden

### Requirement: Förslaget fyller luckor men skriver aldrig över

Plan API:t MUST erbjuda `POST /api/v1/goals/:id/definition/propose` som ger ett
förslag på hur målet ska se ut. Förslaget MUST fylla ENDAST de fält som saknar
ett värde. Ett fält som redan har ett värde MUST lämnas orört, även om
förslaget har ett annat värde för det.

Förslaget MUST innehålla de föreslagna värdena, de antaganden som gjorts och
ingen annan text. Förslaget MUST inte skriva något till målet.

#### Scenario: Förslag på ett tomt mål

- **GIVEN** ett mål med titeln "Lära mig spanska" och inga andra värden
- **WHEN** användaren begär ett förslag
- **THEN** förslaget innehåller värden för beskrivning, deadline, plats,
  upprepning, tidsåtgång, taggar och verifieringskriterium
- **AND** förslaget inte innehåller något värde för titeln

#### Scenario: Befintligt värde lämnas orört

- **GIVEN** ett mål som redan har en deadline
- **WHEN** användaren begär ett förslag
- **THEN** förslaget inte innehåller någon deadline
- **AND** panelen visar att fältet redan har ett värde

#### Scenario: Antaganden redovisas

- **WHEN** förslaget bygger på antaganden om användaren
- **THEN** förslaget innehåller anteckningarna om dem
- **AND** panelen visar anteckningarna

#### Scenario: Ingen API-nyckel

- **GIVEN** att Plan API:t saknar LLM-nyckel
- **WHEN** användaren begär ett förslag
- **THEN** API:t svarar med ett fel som säger att förslag inte kan göras
- **AND** panelen visar fortfarande vilka fält som saknas
- **AND** användaren kan fylla i fälten själv

### Requirement: Godkännande sker fält för fält

Varje föreslaget fält MUST kunna accepteras, ändras eller avvisas separat.
Ett godkännande av ett fält MUST NOT ändra något annat fält.

Användaren MUST godkänna via sin egen session. Förslaget får inte appliceras
utan godkännande.

#### Scenario: Godkänn ett enskilt fält

- **GIVEN** en panel med förslag på beskrivning, deadline och plats
- **WHEN** användaren godkänner bara deadline
- **THEN** målets deadline får föreslaget värde
- **AND** målets beskrivning och plats är oförändrade

#### Scenario: Avvisa ett fält

- **GIVEN** en panel med förslag på upprepning
- **WHEN** användaren avvisar fältet
- **THEN** målets upprepning är oförändrad
- **AND** inget annat fält ändras

#### Scenario: Ändra ett förslag innan det godkänns

- **GIVEN** en panel med förslag på deadline
- **WHEN** användaren ändrar värdet och godkänner
- **THEN** målets deadline får det ändrade värdet
- **AND** inte det föreslagna

#### Scenario: Server applicerar exakt det godkända

- **WHEN** klienten skickar en begäran om att applicera två godkända fält
- **THEN** API:t skriver exakt de två värdena
- **AND** inget annat värde på målet ändras

### Requirement: Appliceringen loggas

Plan API:t MUST logga varje applicerad definition i `goal_definition_applies`
med målet, fältet, värdet från, värdet till och tidpunkten. Loggen MUST
användarens session som aktör.

#### Scenario: Applicering hamnar i loggen

- **WHEN** användaren godkänner en deadline
- **THEN** loggen får en rad med målet, fältet "deadline", det gamla värdet,
  det nya värdet och tidpunkten

#### Scenario: Loggen finns utan agenttjänsten

- **GIVEN** att agent-API:t är nere
- **WHEN** loggen läses
- **THEN** de applicerade definitionerna finns kvar

### Requirement: Verifieringskriterium på målet

Ett mål MUST kunna ha ett verifieringskriterium: en fritext som anger vad som
ska stämma för att målet ska räknas som uppnått. För ett kodmål ska det kunna
vara ett acceptanskriterium, till exempel att ett visst test ska vara grönt.

Kriteriet MUST vara fritt. Ett mål utan kriterium MUST inte blockeras från att
gå vidare till nästa fas. Kriteriet MUST gå att redigera när som helst.

#### Scenario: Sätta ett kriterium

- **WHEN** användaren sätter kriteriet "testet är grönt" på målet
- **THEN** målet har kriteriet
- **AND** kriteriet syns på målsidan

#### Scenario: Kriterium från ett kodmål

- **GIVEN** ett kodmål
- **WHEN** definitionen föreslår kriteriet "npm test passerar utan fel"
- **THEN** kriteriet kan godkännas
- **AND** det sparas på målet

#### Scenario: Mål utan kriterium

- **GIVEN** ett mål som inte har något verifieringskriterium
- **WHEN** användaren byter målets fas från `define` till `plan`
- **THEN** fasen ändras
- **AND** panelen visar en varning om att kriteriet saknas

#### Scenario: Tomt kriterium

- **WHEN** användaren rensar kriteriet
- **THEN** målet har inget verifieringskriterium
- **AND** inget annat påverkas

### Requirement: Definitionen blockerar inte fasbyten

Att en definition är ofullständig MUST NOT hindra användaren från att byta
målets fas. Gränssnittet MUST i stället visa en varning som räknar hur många
fält som saknas och om verifieringskriteriet saknas.

#### Scenario: Varning i stället för hinder

- **GIVEN** ett mål i fasen `define` utan deadline och utan kriterium
- **WHEN** användaren byter fas till `plan`
- **THEN** fasen blir `plan`
- **AND** målsidan visar varningen att fält saknas

#### Scenario: Varningen försvinner när definitionen är komplett

- **WHEN** alla fält har värden och kriteriet är satt
- **THEN** ingen varning visas för målet
