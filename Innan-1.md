# Mål att nå innan jag går över till Plan 1.0

- [x] Testa secrets skill
- [x] Installera zitadel i kluster
- [x] Läsa in mig på skillnaden på oidc och idp. ifall jag kanske ska avnända
- [x] Skydda longhorn url
- [x] Säkerställ inte vem som helst kan skapa ett konto och logga in. Behövs
      något speciellt scope?
- [x] Uppdatera flux
- [x] Skydda rutter
- [x] Ta bort url skydd på flux tillfälligt tills zitadel är tillbaka
- [x] auth.simonbrundin.com fungerar [ ] Dubbelkolla om någon ha arbete måste
      göras i zitadel https://zitadel.com/docs/concepts/architecture/solution
- [x] Flux Webhook fungerar
- [x] Lös detta i Plan: Kustomization Failed 4m ago flux-system/plan health
      check failed after 4m30.058886584s: timeout waiting for:
      [Job/plan/grant-user-permissions status: 'InProgress']
- [x] longhorn.simonbrundin.com använder oidc
- [x] Uptime Kuma fungerade
- [x] Få flux ui att bli skyddad av zitadel igen genom att lägga till security
- [x] Säkerställ att webhook i flux fungerar och startar synkning policy i
      kustomization.yaml igen
- [x] Sätta upp så man inte kan pusha till main utan PR
- [x] Push till plan kör CI-workflow från Deployment Pipeline Repot
- [x] Göra så tester körs så fort en fil sars
- [x] Fundera på hur jag bygger flera micro services med ett ci workflow
- [x] Få spc.simonbrundin.com att fungera
- [x] Skapa pr workflow för plan
- [x] Kunna logga in med Zitadel på Plan
- [ ] Byt ut graphql mot rest
- [ ] Mål syns i prod
- [x] Postgraphile istället för hasura
- [ ] Kolla så min app är upplagd efter
      [länk](https://nuxt.com/docs/4.x/directory-structure)
- [ ] Installera Immich i Homelab
- [x] iptv fungerar
- [ ] Dashboard för allt. En del ska vara dora metrics för mina olika projekt
- [ ] development mode i zitadel. vad betyder det och ska jag ha ktvå appar?
- [ ] Installer Kyverno Chainsaw
- [x] Zitadel nås via auth.simonbrundin.com
- [ ] Jellyfin fungerande
- [ ] Hårddisken i worker-2 fungerar
- [ ] Installera sista datorn i klustret
- [x] Installera Orange Pi:s
- [ ] Plan går att använda
- [ ] Prod av Plan fungerar
- [ ] Pipeline för Plan fungerar
- [x] Få igång plan i prod något enklare än authentik. och hur användarhantering
      egentligen fungerar i en modern webbapp
- [x] skill - skapa encrypted secret

## Cluster

- [x] Lägg till några Kyverno policies
- [x] Push triggar direktsynkronisering
- [x] Longhorn fungerar
- [x] Installera Kyverno
- [x] Få urls att fungera till klustret. grafana.simonbrundin.com
  - [x] Vänta på svar från Körsbär - Annars byt internetleverantör
- [x] Inga fel i flux
- [x] Authentik fungerar
- [x] Vissa routes säkra
- [x] Installera KubeScape
- [x] Föra över de sista komponenterna från ArgoCD till Flux
- [x] Lägga till en Kyverno-policy som gör det omöjligt att radera
      longhorn-system namespacet
- [x] auth.simonbrundin.com fungerar

---

### Meddelande till agent > PR med länk

#### Kan skicka meddelande

- [ ] Jag har en input-ruta i produktion som kan skicka anrop till ett backend
  - [ ] Klustret 100 %
  - [x] Kan skicka meddelanden i dev till backend dev
    - [x] Backend loggar att den tagit emot ett meddelande
    - [x] Inputruta finns i dev
    - [x] Jag kan skapa en ny konversation
      - [x] I drizzle går det att se en tabell med samma schema som i schema.sql
      - [x] Databas med messagestabell finns och gå att se i drizzle
      - [x] Konversationer och meddelanden från dev databasen syns i frontend
      - [x] Lägg till DrawDB i tilt
      - [x] Visa spinner medans konversationer laddas

#### Agent kan hantera meddelande

- [ ] Backend körs i produktion
- [ ] Agenten kan spara det inkommna meddelandet i en databas

#### Agent kan åtgärda problembeskrivning i meddelande i kodbasen

#### Få notiser om nya PRs

Länken ska öppna en ephemeral environment där jag kan preview förändringen

### PITER

#### Prompt Input

- Håller på att implementera
  - Eget GUI med inputruta

- Kandidater
  - GitHub Issues
  - Jira
  - Mitt egna planeringssystem

#### Trigger

- Kandidater
  - Webhook
    - Ska verkligen alla issues trigga en agent körning?

#### Environment

- Kandidater
  - Kubernetes
  - Pi Server

#### Review System

### Dashboard

### Övrigt

- [ ] Skapa tabell för meddelanden
- [ ] Skapa tabell för konversationer
- [ ] Skapa tabell för planer
- [x] Jag vill använda AtlasOperator i produktion för att hantera min databas.
      Hur ska schema förändringar ta sig till produktion? Jag älskar tanken med
      att ha en schema.sql för produktion, likt den här
      /home/simon/repos/plan/environments/kubernetes/overlays/prod/db/schema.sql.
- [ ] Göra så messengekommando tar json som input vilket gör att den kan hantera
      både repo-url och conversation_id
- [ ] Workflow behöver veta vilket repo meddelandet refererar till
- [ ] Hur ger jag agent tillgång till repo?
- [ ] Planeringskommando
- [ ] Skapa kommando som värderar ifall det behövs skapas en feature eller
      scenario
- [ ] Dashboard
  - [x] Konversationer
    - [x] Inputruta där jag kan starta konversationer likt i tex messenger med
          ett sidofält till vänster med alla konversationer
  - [ ] KPIs ska synas
- [ ] Välj terminalagent. Goose, Pi, Opencode mm.
- [ ] Kolla klart på TAC

### Slutfört

- [x] Testa köra dev env
  - [x] Få dev.nu att fungera
  - [x] Ta bort dev.nu
  - [x] Lägg till dev kommando i package.json istället
- [x] Skapa meddelandehanteringsprocess i README
- [x] Skapa .devcontainer
- [x] Skapa databas för agent
- [x] Gör så den kan köra kommandon från commands-mappen i agents-repot
- [x] Trigger - skapa python kommando som jag kan köra med typ
      `message.py "meddelande"` som sedan kör `agent "meddelande"` (just nu
      opencode "meddelande" tills jag bestämt mig för en annan agent)

## Mjukvarumall

- [ ] Skapa ett repo för en mall som går att använda för att komma igång med ett
      nytt kodprojekt snabbt
- [ ] Komma på ett sätt hur jag använder mallen för att starta ett nytt
      kodprojekt
- [ ] Hur håller jag mina projekt uppdaterade med de nya sakerna som kommer in i
      mallen

## Övrigt

- [ ] IR Kontroll
  - [ ] https://smarthomescene.com/reviews/tuya-zigbee-infrared-ir-remote-zs06-review/
  - [ ] https://www.youtube.com/watch?v=d1kkJpolE1w&t=3s
  - [ ] PÅ
    - [ ] B2ERYRHxAYwG4AUDAVkCgANAF8AL4AMHQBdAA0ATwAdAC+AHA0Ab4BMD4Acv4AsPQD9AF0AHQANAC+AHA+ATF+APG+ADQ0AjwAPgBxdADwc1FmERYRHxAcAjgAcFWQJZAvEBQAMDjAbxAUAH4AcDQBdAA0AXwAdAC+AHA0Ab4BMD4Acv4AsPQD9AF0AHQANAC+AHA+ATF+APG+ADQ0AjwAPgBxcDWQLxAQ==
  - [ ] Av
    - [ ] B2YRZhEDAocG4AUDQAEDWwIDAkAXBVsC5gFbAkAHQAOAC0AXQANAD8AHQAsBWwJAEwADYAEDWwLmAUAb4BMDAVsC4AUDgAEB5gFAAwBbIAdAA0A
    - [ ] B1sRWxHxAYoG4AUDAVkCgAMHigZZAlkC8QHgBwMBigaAA0AXQAtAB+ALA0Ab4BMD4Asz4AcTQD9AE0AHQANAC+AHA+ALF0ATwAPgBy/gBw8FigZZAooGQAcD8QFZAoADQAtAA0APB1sWWxFbEfEBQA/gAwNAG0ADQBNAB+AHA0AXQANAF8AHQAvgBwNAG+ATA+AHL+ALD0A/QBdAB0ADQAvgBwPgCxdAE8AD4Acv4AcPwCdAB8AbC4oG8QGKBvEBWQLxAQ==
  - [ ] 8 grader
    - [ ] B2MRYxHyAYYG4AUDAVkCgANAF8AL4AMHQBdAA0ATwAdAC+ADA+ADF0AL4AEDCVkChgbyAVkC8gFAB0ADQAvgAwPgAxPAC0AHCYYGWQKGBvIBWQLgBQMHhgZZAlkC8gHgBwMBhgaAA+AHF+AbD0A7wCdAC+APA0AjB1YWYxFjEfIB4AcjQBtAA0AXQAfgBwNAF0ADQBfAB0AL4AMD4AMXQAvgBwNAK8ATQAvgAwNAF8APwAtAB8ADwBvAB0AXwAvgAwdAF0AD4AMT4B8L4AM3QAvgDwMDWQLyAQ==

## Agent Dashboard

### Infrastructure

- [x] Skapa en ett kommando som heter `simon kubernetes health` som kollar så
      alla noder i mitt taloskluster ser hälsosamma ut. Och som visar att
      kubectl fungerar. jag behöver kunna köra både talosctl och kubectl och
      kubectl fungerar. jag behöver kunna köra både talosctl och kubectl och jag
      vill att det ska säkerställas att det går i början på skripet.

#### Kubernetes Frontend

- [ ] Kunna se alla noder både från talos och kubectl
  - [ ] Typ liknade Omni från Sidero
  - [ ] Jag vill kunna se versioner för både talos och kubernetes på varje nod
        och ha möjligheten att uppdatera dom med ett enkelt knappklick
  - [ ] Hur uppdaterar jag Talos på ett enklare sätt?
    - [ ] [Tuppr](https://github.com/home-operations/tuppr)
- [ ] https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/
- [ ] Kunna se senaste versionerna av talos och kubernetes samt vad som körs i
      mitt kluster

## Plan 1.0

- [x] plan.simonbrundin.com fungerar
- [ ] Inkorg
  - Inmatningsprocess
    1. Skriv text
    - Saker man vill ha in om varje mål
      - Titel
        - Tänk på att det är skillnad på det man ofiltrerat skriver ner så fort
          man öppnar appen och vad som hade varit den optimala titeln
        - Kina-AI ändrar titeln och vill man ändra så kan man göra det senare
      - Ikon
      - Anteckningar
      - Tidsåtgång
      - Föräldrar
- [ ] Skapa appikon för bla PWA

### Sophies önskemål

## Software Agent

- [ ] Dashboard

- [x] auth.simonbrundin.com fungerar
  - [x] Få alla mina noder att fungera
  - [x] Uppgradera alla till senaste talos och Kubernetes

---

### Hemfix

- [ ] Städa
- [ ] Uttag och 18 v batteriladdare uppsatt i hörnskåpet
- [ ] Handfat uppsatt
  - [ ] Packningar täta
    - [x] Köp Packningar
- [x] Inneverktyg
- [ ] Elmätare på badet
- [ ] Termometer på badet
- [ ] Renault igång
- [ ] Dockhus klart

---

- [ ] DevContainers
  - [ ] Simon CLI finns tillgängligt i min DevContainer
  - [ ] NeoVim används i DevContainer
    - [ ] Förstå hur det fungerar när man använder NeoVim
    - [ ] Använd AI för att förså hur officiella dokumentationen föreslår att
          man ansluter med NeoVim
- [ ] Mend Revonate Bot installerat för att uppdatera depencencies
- [x] Prod-ikon
- [x] Robotdammsugare fungerar
- [x] Standard extensions i Chromium
  - [x] Vimium
  - [x] 1Password
  - [x] Video Speed Controller
  - [x] Video Speed Controller
- [x] Linkding för att hantera bokmärken i min Homelab så jag kan byta browser
      utan att all mina bokmärken försvinner utan att all mina bokmärken
      försvinner
- [x] Kunna spara idéer länkar med mera i Notion i en inkorg, både från
      telefonen, datorn och automatiskt från min kindle för anteckningar och
      hightlights
- [ ] Dot AI fungerar

## Klart --------------------------------------------------------------------

### Hemfix

- [ ] Farstulycktan lyser
  - [x] Lycktan moterad
    - [x] Kilen passar
    - [x] Rör nedkortat
    - [x] Kablarna på utsidan skalade
  - [ ] Köp
    - [ ] Lock
    - [ ] Brun sladd
    - [ ] 4 längre skruvar
  - [ ] Lampknappar monterade
  - [ ] Eluttag monterat
  - [ ] Inkopplat i central
  - [ ] Lock på plats
- [ ] Renault igång
- [ ] Dockhus klart

### Talos Cluster

- [x] Prod-ikon
- [ ] Skapa appikon för bla PWA
- [ ] Installera
      [Kubernetes Dashboard](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/)
- [ ] Robotdammsugare fungerar
- [ ] Standard extensions i Chromium
  - [ ] Vimium
  - [ ] 1Password
  - [ ] Video Speed Controller
  - [ ] Video Speed Controller
- [ ] Linkding för att hantera bokmärken i min Homelab så jag kan byta browser
      utan att all mina bokmärken försvinner utan att all mina bokmärken
      försvinner
- [ ] Kunna spara idéer länkar med mera i Notion i en inkorg, både från
      telefonen, datorn och automatiskt från min kindle för anteckningar och
      hightlights
- [ ] Installera Renovate i klustret så jag får PRs om uppdateringar

## Klart --------------------------------------------------------------------

- [x] Få kopiera och klista in att fungera
- [x] Fundera igen om det inte är bättre att köra chezmoi iställer för stow.
      Fördelen jag ser är att jag med ett kommando kan installera allt jag
      behöver på datorn utan att den har något installlerat på sig.
- [x] Uppdatera talos controlplanes med api server oidc settings via
      `simon talos update config`
- [x] `ping` 10.10.10.11 fungerar [x] Kan logga in mot Unifi Gatewayens VPN
      server

- [x] Boka södra och tutte bene
- [x] Vill jag kunna se dev hela tiden?
  - [x] Hur får jag urlen?
  - [x] Startsätt?
  - [x] Hur laddar jag ner den?
    - [x] Enklast kanske är att skapa en länk som jag kan lägga till som en PWA
- [x] Hur vill jag starta min Utvecklingsmiljö?
  - [x] Kör `simon dev`
    - [x] Den låter en välja från en lista med repon via fzf
      - [x] Föreslå först repot man befinner sig i
      - [x] Är det overkill räcker det med att man får välja mellan repon i en
            mapp på datorn?
- [x] Hur ansluter jag mot mitt kluster hemma?
  - [x] Twingate fungerar i mitt kluster och jag kan ansluta mot det från min
        dator och min iPhone
    - [x] Youtubea videor hur Twingate fungerar och specifikt om kubernetes
          operator för Twingate
- [x] Nycklar till Talos och K8s i 1Password
- [x] Ta bort kubeconfig, talosconfig och secrets.yaml från min githistorik
- [x] Balders - ta med
  - [x] Sparkcykel
  - [x] Välling
  - [x] USB-C laddare
  - [x] USB-C hub
  - [x] HDMI-kabel
  - [x] Laddocka i sovrummet
  - [x] Matlådor
  - [x] Tennisgrejer till Noomi
  - [x] Tennisgrejer till mig
- [x] Blåsa rent öron

- [x] Tilt startar på min dator
  - [x] Testa `tilt up`
    - [x] Göra så Tilt använder kubernetes istället för docker-compose.
    - [x] Göra om från docker-compose till manifests
  - [x] I vilket namespace körs tilt-manifests?
- [x] Be pappa betala för 1Password
- [x] Gör löner
- [x] Betala ut friskvård
- [x] Beställ genomskinligt rör från Temu
- [x] Kan logga in på kubernetes via Teleport
  - [x] Använda Claude för setup
- [x] DevContainer körs i kubernetes
  - [x] Installera Kubernetes provider i Devpod
  - [x] devpod up github.com/simonbrundin/plan
  - [x] devpod up github.com/ditt-användarnamn/ditt-repo@min-coola-branch
- [x] Köp zigbee repeater ute
- [x] Köp zigbee termoemeter till badet
- [x] Hur ska länken skapas?
- [x] Hur ska länken skapas?
- [x] Hur ska länken skapas?
- [x] Hur ska länken skapas?
- [x] Tilt fungerar
  - [x] Säkerställ att det går att bygga image från Dockerfile.dev och köra med
        docker
  - [x] Det går att köra `tilt up` via `dev`
    - [x] Den fastnar i byggprocessen pga av extra paketen den behöver
          installera. Kan man göra om så den använder en annan image som redan
          innehåller paketen eller kan man använda mise?
  - [x] Red ut varför podden startar om
  - [x] Skapa en kubernetes resource med en port forward så jag kan se min nuxt
        dev server på en lokal port tex via **localhost:3000**
    - [x] Vilket namn? nuxt dev server?
- [x] Hämta paket åt mamma
- [x] Httproute till min dev server
- [x] Undersök om det är värt att bygga imagen i klustret istället för då
      slipper jag kanske använda ttl.sh
  - [x] https://github.com/tilt-dev/tilt-example-builders/blob/main/kubectl_build/Tiltfile
- [x] Skriv om cli så min update talos config inte sparar generated som en fil
      för i den finns känslig data
- [x] Hämta mammas bil
- [x] Tilt fungerar igen
- [x] Kan ansluta mot databas
  - [x] Få igång Drizzle Studio så jag ser att databasen fungerar
- [x] Tilt Resources
  - [x] Postgres-databas
- [x] Farstulycktan lyser
  - [x] Lycktan moterad
    - [x] Kilen passar
    - [x] Rör nedkortat
    - [x] Kablarna på utsidan skalade
- [x] `s k login teleport` kan logga in med thetisnyckel
- [x] `dev` fungerar i plan och ger mig notis i telefonen
  - [x] Auth fungerar
  - [x] Auth fungerar
  - [x] Auth fungerar
  - [x] Auth fungerar
- [x] Köp
  - [x] Lock
  - [x] Brun sladd
  - [x] 4 längre skruvar
- [x] Lampknappar monterade
- [x] Eluttag monterat
- [x] Inkopplat i central
- [x] Lock på plats
- [x] Lager157
  - [x] Keps
  - [x] Tjocksockar
  - [x] Tjocksockar
- [x] Wayland visar målet jag jobbar på nu
- [x] Kursplan för programmering klar
- [x] Alla datorer har senaste Talos och Kubernetes
- [x] Allt i ArgoCD är grönt
- [x] Allt i ArgoCD är grönt
- [x] Allt i ArgoCD är grönt
- [x] Allt i ArgoCD är grönt
- [x] Longhorn är helt grönt
  - [x] Koppla in och ur hårddisken i worker-3
  - [x] worker-8 fungerar
  - [x] worker-4 fungerar
  - [x] worker-2 fungerar
  - [x] worker-7 fungerar
- [x] Två hårddiskar i worker-1 fungerar
- [x] Ändra storage class på plan för cnpg
- [x] https://domain-locker.com/about/self-hosting
- [x] https://domain-locker.com/about/self-hosting
- [x] https://domain-locker.com/about/self-hosting
- [x] https://domain-locker.com/about/self-hosting
- [x] https://domain-locker.com/about/self-hosting
- [x] Vaccintester
  - [x] Ta bort felmeddelande
  - [x] Logga
    - [x] Be Daniel skicka nuvarande logga
    - [x] Gör en enklare logga för Oavsett med en testikon
  - [x] Peka Oavsett till seiq.se
  - [x] Få Fortknox integration att fungera med Woocommerce
    - [x] Se om det går att få in produkterna från Fortknox
  - [x] Meny
    - [x] Butik
    - [x] Om oss
    - [x] Kontakt
    - [x] Dölj sociala medier tillsvidare
  - [x] Swish
  - [x] Dölj kontaktformulär och gör en annan kontaktruta
  - [x] Felsök network change
  - [x] Inloggning till wordpress.com
  - [x] Man behöver inte ha en inloggning för att handla
  - [x] SimplyBook behövs inte
  - [x] Få fraktalternativ att fungera
    - [x] Gör så att när man handlar över 500kr så blir det fri frakt
  - [x] Betalningsalternativ
  - [x] Produkter i ordning
    - [x] Daniel skickar mig lista på alla produkter som ska vara med
      - [x] Inkludera bilder
    - [x] Daniel skickar mig lista på alla produkter som ska vara med
      - [x] Inkludera bilder
      - [x] Inkludera bilder
      - [x] Inkludera bilder
      - [x] Inkludera bilder
      - [x] Inkludera bilder
- [x] Kolla in Mise
- [x] Komma igång med Farcic's ai workflow
  - [x] Hur fan vill jag att ett workflow ska se ut?
    - [x] Kolla med Grok
    - [x] Potentiella steg
      - Definera och spara problemet
        - Skapa Issue i GitHub
          - Ska finnas i Issue
            - Titel: Kort & självförklarande - Vad ska uppnås? (The Why + What)
            - Acceptance criteria
              - Punkter som måste vara uppfylld för att kunna stänga issuen
            - Edge cases & icke-funktionella krav
              - Felhantering: ...
              - Prestanda: ...
              - Säkerhet: ...
              - Kompatibilitet: ...
            - Teknisk kontext / begränsningar
              - Stack: FastAPI 0.115+, Python 3.12, PostgreSQL 16
              - Redan existerande: auth-router i src/api/v1
              - Får INTE: använda sessions (vi kör stateless JWT)
            - Hur jag tänkt (valfritt – bra för komplexa issues)
              - Min初步 tankegång / föreslagen approach:
              - Skapa ny modell TokenPair
              - Implementera /login → returnerar access + refresh
              - Ny endpoint /refresh
            - Relaterat
              - #17 User model & databasstruktur
              - ADR-003: Token-strategi
              - Figma: https://...
            - /label prio:medium feature backend /estimate M] Skriv tester -
            - Labels
              - time-estimate
            - Assignee?
      - TDD/BDD
      - Koda
      - Refactor - Förbättra, Förenkla, Kodprinciper
        - Clean Code Principer
      - Dokumentation - Uppdatera projektets README och agentens minne
        - AGENTS.md
      - PR - Tester ska först gått igenom
        - En beskriving hur jag verifierar att allt fungerar. Tex en länk till
          ett frontend med beskrivning över vad jag ska klicka på och se. Eller
          något jag kan köra i terminalen för att testa att en backend fungerar
          som det är tänkt, tex ett api-anrop
- [x] Visa alla Issues i mina repos och alla agent som arbetar på något
- [x] Visuellt mer likt Sidecar
- [x] Lägg till i simon cli ett kommando som är simon ai tui som öppnar mitt ai
      tui
- [x] Installera OpenClaw på en Raspberry Pi 5
- [x] Red, Green, Refactor i /implement
- [x] /work
- [x] Dark Mode
- [x] Uppdatera standard så att den kommer ihåg att all text som visas i x
      frontend ska vara på svenska
- [x] Svenska i Frontend
- [x] Bara ett backend api
- [x] Skriv det som standard att det enda backend api som ska byggas är Go och
      inte in Nuxt.
- [x] Scraping sidan fungerar
- [x] Lägg till snurra som visas medan varje sida väntar på svar från api.
- [x] Hittade annonser fungerar
- [x] Annonser visar produktnamn, pris, fraktkostnad, värdering
- [x] Okänd produkt
- [x] Hämta pris, fraktkostnad
- [x] Priser visas korrekt, nu visas 3800 sim 38.00
- [x] Fraktkostnad hämtas korrekt
- [x] Bästa gratis modellen?
- [x] Ändra /work så den inte måste läsa varje fil. den borde kunna använda rg
      för att hitta vilka specs som är färdiga och det borde gå mycket fortare
      att att behöva använda read på varje spec. hittar den några specs som inte
      är helt färdiga så kan den isf läsa bara dessa.
- [x] Beskrivning i description istället för titel
- [x] Rätt description sparas
- [x] Fraktkostnad sparas i databasen med ören istället för kr. så ifall det
      står 19,50 kr så sparas det som 1950
- [x] Blocket anropas via API och hämtar inte HTML
- [x] Värderingsfunktion
- [x] Vinstuträkning
- [x] Hämta nya annonser fungerar
- [x] Visa delvärderingarna i /ads
- [x] Dela upp product_valuations och listing_valuations. just nu är bara
      delvärderingarna kopplade till en produkt
- [x] Visa statusuppdateringar, samma sak som i loggarna när hämta annonser körs
- [x] Värderingsfunktionen fungerar och värderingen stämmer överens med
      delvärderingarna

## Begbot

## Övrigt

- [x] Lägga in alla produktet till Daniel
- [x] Minimi förändring av portfölj för att trigga email. kanske 0.2% till att
      börja med i Neuro repot.
- [x] Skriva till Kristian och tacka för förra helgen
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Visa delvärderingarna i /ads
- [x] Dela upp product_valuations och listing_valuations. just nu är bara
      delvärderingarna kopplade till en produkt
- [x] Visa statusuppdateringar, samma sak som i loggarna när hämta annonser körs
- [x] Värderingsfunktionen fungerar och värderingen stämmer överens med
      delvärderingarna

## Begbot

## Övrigt

- [x] Lägga in alla produktet till Daniel
- [x] Minimi förändring av portfölj för att trigga email. kanske 0.2% till att
      börja med i Neuro repot.
- [x] Skriva till Kristian och tacka för förra helgen
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail
- [x] Hjälpa norton med mail

## Agent

- [x] Notifikationer från OpenCode när den är klar eller vill ha input
- [x] /free-model skapa ett kommando som kollar i
      https://openrouter.ai/models?q=free och ser vilka som har aktiva
      providers. sedan rekommenderar du en lista på topp tre baserat på de som
      är bäst för kodning och driftning av infrastruktur
- [x] Kanban board för agentos specs
- [x] Karpathy i agents.md
  - [x] https://x.com/godofprompt/status/2018482335130296381
- [x] Fråga om push och commit när en spec är färdig
- [x] Fråga om push och commit när en spec är färdig
- [x] Fråga om push och commit när en spec är färdig
- [x] Fråga om push och commit när en spec är färdig
- [x] Fråga om push och commit när en spec är färdig

## Slutfört

- [x] Uppgradera alla noder
  - [x] Talos
    - [x] 1.11.6
      - [x] controlplanes
      - [x] workers
    - [x] 1.12.4
      - [x] controlplanes
      - [x] workers
- [x] Få igång worker 2
- [x] Få igång worker 9

### Begbot

- [x] Låt LLM värdera annonstexten ifall produkten har hel.
- [x] Dela upp inställningar
  - [x] Mejlinställningar
    - [x] alt 1
      - [x] Aktiv produkt
      - [x] rabatt 15 %
      - [x] vinst 200 kr
    - [x] alt 2
      - [x] Inaktiv produkt
      - [x] minst antal annonser värderingen är baserad på - 10
      - [x] Kolla på lägsta värderingsmetoden
        - [x] rabatt 15 %
        - [x] vinst 200 kr
  - [x] Sparningsinställningar
    - [x] värderingssäkerhet 0 %
    - [x] värde - 500
    - [x] minst antal annonser värderingen är baserad på - 10
  - [x] Ekonomiinställningar
    - [x] Omsättningsdagar i inställningar
  - [x] Auto Enable Produkt inställningar
    - [x] Minsta antal annonser 10
    - [x] Värderingssäkerhet 90 %
    - [x] värde 500

- [x] Skapa permanenta loggar för scrapern
  - [x] Göra så loggen är kvar på sidan när ha klickar på "Hämta annonser" och
        går
- [x] Dubbelkolla status på scraping runs ifall det blir ett error
- [x] Emailuppdateringar
  - [x] Vinst måste inkludera frakt och köpskydd
  - [x] Visa rabatt i email

- [x] Visa tilläggsdatum i /ads
- [x] Se om det går att snabba upp hämtningen av listing i /ads
- [x] Bredvid värderingen på en listing vill jag se en procentsats hur säker den
      är på värderingen är korrekt
- [x] sent_emails - Skapa en tabell med alla email som sänts och vilken listing
      id de rekommenderat

- [x] Automatiskt aktivera produkter som har minst visst många produkter den
      värderas på och som har en säkerhetsprocent över trading_rules och där
      sammanvägda värderingen är minst min_profit/min_discount
- [x] Spara bara produkter och listings som har ett värde på minst
      min_profit/min_discount
- [x] bug - just nu använder blocket värderingsmetoden bara den första kategorin
      i
- [x] Av-/aktivera search terms i /scraping/terms

- [x] Testa lägga in alla produkter från blocket som en sökterm

- [x] Ändra sparande annonser till mejlade annonser i /ads
- [x] Lägg till en sida med värderingsmetoder
- [x] Rensa /ads
  - [x] Ta bort 2000 kr ebay från varje annons i /ads
  - [x] Ta bort nypris
  - [x] Dölj beskrivning för varje annons i /ads

- [x] Scrapinghistorik
  - [x] Hittades
  - [x] Nya annonser - som inte fanns i cache
  - [x] Nya produkter
  - [x] Sparade produkter - som uppfyllde kraven för att bli sparade
  - [x] Sparade annonser - som uppfyllde kraven för att bli sparade
  - [x] Mejlade annonser - som uppfyllde emailkraven och skickades som en
        köprekommendation

- [x] Ikoner bredvid alla meny-items från https://nuxt.com/modules/icon
- [x] Skapa en separat sida för search terms och en för scrapings där man lista
      alla gånger den försökt scrapa
- [x] Olika tabbar i /ads där en är Alla och en är Prisvärda
- [x] Avbryt scraping knapp och cli kommando
- [x] Produkter visar i /products
- [x] Inloggning med Supabase Auth
- [x] Skicka mail ifall jag vinsten är över mina begränsningar i trading_rules
- [x] Lista skickade mail i frontend
- [x] Budfunktion
- [x] Scrapingkörningar - Visa en lista med alla gånger scraping funktionen körs
      och visa information hur många nya annonser den hittat, hur många som var
      köpvärda mm
- [x] BEGBOT: Mejl skickas med riktig data och inte fejkdata från mail.html
- [x] Flytta över till PXE så Begbot körs 24/7 ifrån sidan och sedan kommer
      tillbaka till samma sida
- [x] trading_rules går att ställa in i frontend
- [x] Värderingar
- [x] Delvärderingar visas för varje produkt i /products
- [x] Voice input för linux, whisper alternative
- [x] Sätta upp Tailscale eller Headscale
- [x] Först förstå skillnaden mellan typ tailscale och twingate. kolla gammal
      grok konversation
- [x] BUG: cronjobs syns inte
- [x] Göra så man kan aktivera och avaktivera värderingstyper för olika
      produkter och att den sammanvägda värderingen bara använder sig av de som
      är aktiverade
- [x] Sammanvägd värdering visas i /products per produkt. Det är bara en formel
      som tar ett medelvärde av de värderingar man har. Ha också med en
      procentsats för hur säker man är på värderingen bredvid
- [x] Värderingscellerna ska gå att klicka på och ändra

- [x] Lägga till fler söktermer
  - [x] Spelkonsoller
    - [x] PS5
  - [x] Fjällräven
  - [x] Verktyg
    - [x] Milwaukee
    - [x] Makita
    - [x] Festool
    - [x] DeWalt
  - [x] Se om det går att ta en hel kategori

- [x] Pagination på /products
- [x] bug - köpskyddet måste också dras av när vinsten beräknas
- [x] /products
  - [x] bug - går inte att ta bort produkter i /products via ta bort knappen
  - [x] fix - kunna klicka och ändra namn på en produkt i /products och då ska
        det såklart också sparas till databasen
  - [x] Sökning av produkter
  - [x] Filter

- [x] Scrapingfunktion
  - [x] Logga information om varje steg en annons går igenom
  - [x] Dubbelkolla ner sparad information
    - [x] Dubbelkolla så den bara hämtar annonser som inte finns i databasen
    - [x] Dubbelkolla det bara sparas annonser som kan skickas
    - [x] Frakt - När den hämtar frakt så blir det okänd när det står frakt från
  - [x] Värdering går att lita på
    - [x] Tradera
    - [x] Egen värdering
  - [x] Köpskydd - Spara ner pris för ev köpskydd och visa bara i frontend ifall
        det alternativet finns
- [x] Se säkerhetsprocents för värderingen på varje annons i /ads
- [x] Mail 2.0
  - [x] Säkerhetsprocent, frakt och produkt i mail
  - [x] Ta bort nypris
  - [x] Se till att inga rekommendationer skickas för produkter som är enabled
        false
  - [x] Bilder visas med hjälp av bildlänkar
  - [x] Skicka bara rekommendationer ifall säkerhetsprocenten för värderingen är
        över ett visst värde som defineras i trading_rules och går att ändra i
        /settings
- [x] Uppdatera värdering knapp i /products, en knapp för alla och en knapp per
      produkt.
- [x] Pagination
- [x] Undersöka varför så många listing blir kategoriserade som fel produkt
- [x] Få ner laddningstiden på /ads
- [x] /ads använder den sammanslagna värderingen från /products
- [x] Blocket värderingsmetod - Medianpris
- [x] Senaste värderingsdatum visas per produkt
- [x] Få Tailscale att fungera så jag kan komma åt alla mina enheter som är
      anslutna till min router när jag inte är inloggad på hemnätverket
- [x] Se ett datum på värderingen
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Värderingscellerna ska gå att klicka på och ändra

- [x] Lägga till fler söktermer
  - [x] Spelkonsoller
    - [x] PS5
  - [x] Fjällräven
  - [x] Verktyg
    - [x] Milwaukee
    - [x] Makita
    - [x] Festool
    - [x] DeWalt
  - [x] Se om det går att ta en hel kategori

- [x] Pagination på /products
- [x] bug - köpskyddet måste också dras av när vinsten beräknas
- [x] /products
  - [x] bug - går inte att ta bort produkter i /products via ta bort knappen
  - [x] fix - kunna klicka och ändra namn på en produkt i /products och då ska
        det såklart också sparas till databasen
  - [x] Sökning av produkter
  - [x] Filter

- [x] Scrapingfunktion
  - [x] Logga information om varje steg en annons går igenom
  - [x] Dubbelkolla ner sparad information
    - [x] Dubbelkolla så den bara hämtar annonser som inte finns i databasen
    - [x] Dubbelkolla det bara sparas annonser som kan skickas
    - [x] Frakt - När den hämtar frakt så blir det okänd när det står frakt från
  - [x] Värdering går att lita på
    - [x] Tradera
    - [x] Egen värdering
  - [x] Köpskydd - Spara ner pris för ev köpskydd och visa bara i frontend ifall
        det alternativet finns
- [x] Se säkerhetsprocents för värderingen på varje annons i /ads
- [x] Mail 2.0
  - [x] Säkerhetsprocent, frakt och produkt i mail
  - [x] Ta bort nypris
  - [x] Se till att inga rekommendationer skickas för produkter som är enabled
        false
  - [x] Bilder visas med hjälp av bildlänkar
  - [x] Skicka bara rekommendationer ifall säkerhetsprocenten för värderingen är
        över ett visst värde som defineras i trading_rules och går att ändra i
        /settings
- [x] Uppdatera värdering knapp i /products, en knapp för alla och en knapp per
      produkt.
- [x] Pagination
- [x] Undersöka varför så många listing blir kategoriserade som fel produkt
- [x] Få ner laddningstiden på /ads
- [x] /ads använder den sammanslagna värderingen från /products
- [x] Blocket värderingsmetod - Medianpris
- [x] Senaste värderingsdatum visas per produkt
- [x] Få Tailscale att fungera så jag kan komma åt alla mina enheter som är
      anslutna till min router när jag inte är inloggad på hemnätverket
- [x] Se ett datum på värderingen
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Cron Jobs
- [x] Cron Jobs
