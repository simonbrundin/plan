## Context

Användaren har ett brett registersystem: mål från hela livet plus många
kodprojekt, av många olika slag. Allt hanteras som *mål* - även det som ligger
under ett annat mål. Han vill att en agent ska kunna ta ett mål från
definition till slutfört och verifierat, och han vill godkänna det agenten
gör.

Det som finns i dag:

- Målmodell med fem faser (`statuses`-tabellen), godkännanden
  (`agent_approvals`), körningar (`agent_runs`), händelser
  (`agent_run_events`), kostnad och budget.
- Agenten kör i klustret som ett Kubernetes Job per körning
  (`src/agent/internal/jobcontroller`), med egna profiler
  (`src/agent/internal/profile`).
- Human gate är strukturellt: agentens API-nyckel har scope `agent` och kan
  inte besluta om ett godkännande.
- Taggar på mål finns redan (`tags`, `goal_tags`) och används för att
  beskriva *där* något hör hemma ("ute", "hem", "vinter").

Begränsningar som formatet anger:

- Plan äger målen. Agenten äger förslagen.
- Faser, tillstånd och anrop ligger i det versionerade kontraktet
  `src/contracts/plan-agent.yaml` och testas från båda sidor.
- Hemligheter krypteras med SOPS + age. Ingen plaintext i git.

## Goals / Non-Goals

Goals:

- Fasövergångar ska vara deterministiska kod med villkor, inte modelltext.
- Ett mål ska kunna avslutas av människan med en enda handling, utan att
  agenten eller en verifieringsplan är inblandad.
- När en verifieringsplan finns ska den ligga bakom agentens förslag, så att
  "verifierat" är bevis, inte en känsla.
- Agentens rättigheter ska följa uppgiften, inte målets ämne.
- Allt agenten skapar ska vara mål - samma begrepp användaren redan använder.

Non-Goals:

- Inga sensorer eller Home Assistant. Bevisen kommer från användaren,
  kalendern, incheckningar och - för kodmål - repot och CI.
- Inga autonoma köp, bokningar eller utskick. De kräver godkännande
  oavsett inställning.
- Ingen egen agentdatabas för mål. Bara konversationer ligger i agentens DB.
- Inte den dagliga pulsen, kvällscheck-in eller notiserna. De är egna
  changes som bygger på den här.

## Decisions

### Decision: Fasmaskinen ligger i Plan API, inte i agenten

Övergången `POST /api/v1/goals/:id/phase` kontrollerar förutsättningarna i
Go-kod och loggar i `status_updates`.

Alternativ: agenten sätter status via den befintliga statusendpointen.

Varför: villkor som "kräver att verifieringen gick igenom" eller "kräver att
alla undermål är avslutade" kan inte garanteras av en modell. En modell kan
också föreslå `completed` - men API:t vägrar om bevis saknas. Det gör
systemet begripligt: statusen säger vad som är sant, inte vad agenten tyckte.

### Decision: Avslutandet har två vägar, och den enkla vägen är förstahands

**Väg A - människan avslutar.** `POST /api/v1/goals/:id/complete` (eller
samma fasanrop med `to=completed` och `by=human`) avslutar målet med en
handling. Ingen verifieringsplan behövs, ingen agent inblandad. Exempel:
"Byt glödlampa" - användaren byter lampan och trycker `d`, så är den klar.

**Väg B - agenten föreslår.** Agenten får föreslå `completed`. Om målet har
en verifieringsplan måste planen gå igenom först och omdömet visas i
godkännandet. Om målet inte har någon plan får förslaget ändå fram, men
gränssnittet visar tydligt att det är ett förslag utan bevis.

Varför två vägar: verifieringsmaskineri som är obligatoriskt blir nonsens för
enklare mål, och det är just de enkla målen som är flest. Att göra
verifieringen till ett valfritt extra gör att systemet fungerar från dag ett
och att tyngden läggs på de mål där den faktiskt ger något.

### Decision: Verifieringsplanen är deklarativ, valfri och har en styrka

Typer: `checklista`, `delmal`, `kalender`, `bevis`, `matvarde`, `kod`.
Varje typ har ett villkor (t.ex. "3 av 4 moment", "alla undermål klara",
"ett kalenderevent har passerat", "ett värde finns", "grön CI-körning").

Styrka: `stark` (beviset är objektivt) eller `svag` (användarens intyg).

Varför två styrkor: verifiability varierar enormt mellan mål. Ett kodmål har
en objektiv grön CI. "Kött en ny hobby" har det inte. Att kräva samma
hårdhet av alla mål skulle antingen blockera de mjuka målen helt eller göra
det svaga målen lika mycket vilda som idag. Användaren väljer per mål - och
det är valfritt, så ett mål utan plan fungerar precis som idag.

### Decision: Agentens rättigheter följer uppgiften, inte målets ämne

Det första förslaget var `goals.execution_kind` med `kod` och `liv`.
Användaren avvisade det: målen har många olika typer, och gränsen mellan
"kodmål" och "livsmål" går inte att dra rent - ett livsmål kan kräva kod, ett
kodmål kan kräva research.

I stället bär **körningen** en åtgärdsprofil:

| Profil | Vad agenten får |
| --- | --- |
| `planering` (standard) | Prata, bryta ned, föreslå, fråga. Skriver ingenting. |
| `research` | Läsa och samla underlag (webb, repos, dokument). Skriver ingenting. |
| `arbete` | Arbeta i en isolerad kopia, skriva kod, föreslå ändringar via branch och PR. |

Profilen föreslås av agenten i definitionen och godkänns av användaren.
Målets ämne är en **tagg** - den finns redan i Plan (`tags`, `goal_tags`) och
behöver ingen ny modell.

Varför: rättigheter är det som ska vara hårda, och de har ett enda sätt att
vara fel - att ge en agent mer än den behöver. En profil per uppgift ger
precis det, och den växer bara när du godkänner en större profil. Underliggande
kodsiffrar `profile.Coding` och `profile.Life` blir en implementationdetail:
`arbete` mappar mot kodprofilens rättigheter, `planering` och `research` mot
livsprofilens.

### Decision: Godkännande har en huvudregel och en valfri spets

**Huvudregeln:** agenten får inte agera på ett mål förrän du godkänt dess
förslag. När du godkänner förslaget får agenten arbeta inom det - skapa
undermål, justera datum, uppskatta tid.

**Undantaget som aldrig går att stänga av:** allt som lämnar datorn - köpa,
boka, skicka, ringa - kräver ett eget godkännande, varje gång.

**Valfri spets:** `goals.approval_mode` med `plan` (standard) eller
`varje_atgard`. I läget `varje_atgard` frågar agenten om varje enskild
åtgärd - det passar för mål där du vill ha full kontroll.

Varför: skillnaden mellan lägena är helt enkelt "måste du godkänna agentens
plan en gång, eller varje enskild åtgärd?". Standard är `plan`, eftersom
"godkänn varje delmål" blir administration. Det som aldrig får ske utan
godkännande är de externa handlingarna, och det ligger utanför läget.

### Decision: Agenten kör i klustret, bevisen kommer från användaren

Ingen Home Assistant, inga sensorer. Beviskällor:

- `bevis` - användaren laddar upp eller skriver in ett bevis (text, bild,
  fil, länk).
- `checklista` - checkmoment avbockas av användaren.
- `delmal` - undermål som avslutas.
- `kalender` - händelser kopplade till målet som passerat.
- `matvarde` - ett uppmätt värde (vikt, antal, tid).
- `kod` - branch/PR/CI-status för kodmål.

Varför: agenten behöver inga externa integrationer för att verifiera, och
användaren slipper dela hemligheter med hemlighetshantering i klustret. En
sensorbaserad verifierare kan läggas till senare som en ny typ - modellen
bär den kostnaden inte.

### Decision: Två API:er, en förslag - en beslut

Agenten får skapa och uppdatera verifieringsplanar, bevis och
verifieringsomdömen, men kan inte besluta ett godkännande (det redan
regleras av `scope = agent` i
`src/api/internal/middleware/agent_scope.go`).

Varför: samma princip som för godkännanden - beslut fattas alltid av en
människa i Plan.

### Decision: Människan får hoppa över faser

En människa får avsluta ett mål direkt från vilken fas som helst - även från
`define`. Faserna är ett stöd för agentens arbete, inte en spärr mot
användaren.

Varför: genvägen ska kännas som en genväg. Ett mål som varar en minut ska inte
gå igenom fyra faser för att avslutas, och en människa vet bäst när något är
klart. Agenten däremot får aldrig hoppa över faser - för agenten är
faserna ett kontrakt.

### Decision: Kodmålens verifierare definieras senare

Vad som ska räcka som bevis för ett kodmål - om grön CI räcker, eller om koden
också måste ha granskats av en människa - är inte avgjort. Det definieras när
kodagenten byggs ut på riktigt.

Varför: det är ett separat beslut med separata konsekvenser, och det ska inte
blockera att grundmekaniken (fasmaskin, valfri verifiering, profiler) kan
byggas. Tills dess stödjer modellen båda: `kod`-verifieraren kan kräva CI
och mänsklig granskning.

## Risks / Trade-offs

- **Fasövergångar kan kännas motstridiga** om villkoren är för strikta.
  → Mitigation: villkoren är synliga i gränssnittet ("varför kan inte det här
  gå vidare?"), och Väg A gör att en människa alltid kan avsluta direkt.
- **Två vägar att avsluta kan ge två sanningar** - ett mål som avslutas av
  människan utan verifiering, och ett som verifierats.
  → Mitigation: gränssnittet visar skillnaden ("deklarerat" mot "verifierat"),
  och `completed` bär alltid *hur* det avslutades.
- **Bevislager kan växa och bli brus** → Mitigation: bevis hör till ett
  verifieringsomdöme och ett mål; gamla omdömen arkiveras, och API:t listar
  bara det senaste omdömet per mål.
- **Agenten kan föreslå svaga verifierare** för att slippa undan bevis.
  → Mitigation: `svag` är synlig i gränssnittet, och en måltyp med objektiva
  bevis (kod) godkänner inte `svag` som enda verifierare.
- **Profilen `arbete` ger agenten skrivrättigheter** - den största risken i
  systemet.
  → Mitigation: profilen är explicit, godkännd och synlig; jobbet körs
  isolerat utan Kubernetes-token; kodändringar landar i en branch och en PR,
  aldrig direkt i main.

## Migration Plan

1. Nya tabeller `goal_verifications`, `goal_evidence`,
   `goal_verification_results` samt `goals.approval_mode` med default `plan`.
   Tomma - inget historiskt att migrera.
2. Den befintliga statusendpointen (`PATCH /api/v1/goals/:id/status`) behålls
   för människans manuella fasbyten, men går via samma maskin internt så att
   villkoren gäller även där. Ett manuellt byte till `completed` utan bevis är
   därmed alltid tillåtet - det är Väg A.
3. Rollback: kolumnen och tabellerna kan ligga kvar; endpointen stängs av med
   en feature toggle. Ingen data går förlorad.

## Open Questions

Inga blockerande frågor kvar för den här ändringen. Kommande förslag (agent-
roller, morgonpuls, kvällscheck-in, notiser) får sina egna frågor när de
tas fram.
