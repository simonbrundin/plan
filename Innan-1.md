# Mål att nå innan jag går över till Plan 1.0

## Övrigt

- [ ] Komma igång med Farcic's ai workflow
  - [ ] Hur fan vill jag att ett workflow ska se ut?
    - [ ] Potentiella steg
      - [ ] Definera problemet
- [ ] Lägga in alla produktet till Daniel
- [ ] Skriva till Kristian och tacka för förra helgen

## Back on Kubernetes-track

- [ ] Installera Talos på datorer
  - [ ] Mac Mini
  - [ ] HP
- [ ] Alla Talos noder fungerar och är inkopplade
  - [ ] Ping av pxe och talos noder fungerar
- [ ] Ping av pxe och talos noder fungerar
- [ ] ArgoCD fungerar
- [ ] Teleport fungerar
- [ ] Vault fungerar
- [ ] auth.simonbrundin.com fungerar
- [ ] Installera
      [Kubernetes Dashboard](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/)
- [ ] Installera Renovate i klustret så jag får PRs om uppdateringar

## OpenClaw

- [x] Installera OpenClaw på en Raspberry Pi 5

## Begbot

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
- [ ] Värderingsfunktionen fungerar och värderingen stämmer överens med
      delvärderingarna
- [ ] Skapa en separat sida för search terms och en för scrapings där man lista
      alla gånger den försökt scrapa
- [ ] Olika tabbar i /ads där en är Alla och en är Prisvärda
- [ ] Avbryt scraping knapp och cli kommando
- [ ] Produkter visar i /products
- [ ] Ikoner bredvid alla meny-items från https://nuxt.com/modules/icon
- [ ] Inloggning med Supabase Auth
- [ ] Skicka mail ifall jag vinsten är över mina begränsningar i trading_rules
- [ ] Lista skickade mail i frontend
- [ ] Budfunktion
- [ ] Skicka meddelande till säljare funktion
- [ ] Scrapingkörningar - Visa en lista med alla gånger scraping funktionen körs
      och visa information hur många nya annonser den hittat, hur många som var
      köpvärda mm

## Agent

- [ ] Notifikationer från OpenCode när den är klar eller vill ha input
- [ ] /free-model skapa ett kommando som kollar i
      https://openrouter.ai/models?q=free och ser vilka som har aktiva
      providers. sedan rekommenderar du en lista på topp tre baserat på de som
      är bäst för kodning och driftning av infrastruktur
- [ ] Kanban board för agentos specs
- [ ] Karpathy i agents.md
  - [ ] https://x.com/godofprompt/status/2018482335130296381
- [ ] Fråga om push och commit när en spec är färdig

### Infrastructure

- Skapa en ett kommando som heter `simon kubernetes health` som kollar så alla
  noder i mitt taloskluster ser hälsosamma ut. Och som visar att kubectl
  fungerar. jag behöver kunna köra både talosctl och kubectl och kubectl
  fungerar. jag behöver kunna köra både talosctl och kubectl och jag vill att
  det ska säkerställas att det går i början på skripet.

#### Kubernetes Frontend

- [ ] Kunna se alla noder både från talos och kubectl
  - [ ] Typ liknade Omni från Sidero
  - [ ] Jag vill kunna se versioner för både talos och kubernetes på varje nod
        och ha möjligheten att uppdatera dom med ett enkelt knappklick
  - [ ] Hur uppdaterar jag Talos på ett enklare sätt?
    - [ ] [Tuppr](https://github.com/home-operations/tuppr)
- [ ] https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/

## Plan 1.0

- [ ] plan.simonbrundin.com fungerar
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

- [ ] Dashboard

- [ ] auth.simonbrundin.com fungerar
  - [ ] Få alla mina noder att fungera
  - [ ] Uppgradera alla till senaste talos och Kubernetes
- [ ] Installera Orange Pi:s

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
  - [x] NeoVim används i DevContainer
    - [x] Förstå hur det fungerar när man använder NeoVim
    - [x] Använd AI för att förså hur officiella dokumentationen föreslår att
          man ansluter med NeoVim
- [x] Skriv om cli så min update talos config inte sparar generated som en fil
      för i den finns känslig data

### Vad är jag ute efter?

- Mobilikon med Hot-Reload
  - När ska dev synas? Ska min dev env alltid köras? Belastar det inte mitt
    kluster i onödan?
    - När och hur ska min dev env stängas?
    - Ska jag använda vCluster?
  - Vilken url vill jag kunna se applikationen ifrån?
    - dev-simonbrundin.appdomain.com?
    - Ska den vara publik?
    - Hur skapas den?
- Avlasta min laptop
  - Måste något köras lokalt? Vad är problemen med att ha allt remote alltså i
    mitt Taloskluster?
- Miljöerna lika varandra
- Nyckelhantering - så inga hemlisar läcker ut
  - [Git Credentials Manager](https://www.vcluster.com/blog/using-devpod-with-private-git-repositories)
  * Hur ansluter jag till mitt kluster utan att behöva kopiera min kubectl
    config överallt?
    - vCluster?
- Utvecklingsverktyg och IDE projektanpassat
  - Hur får jag det att fungera med NeoVim eller andra IDE än VS Code?
  - Hur installerar jag paket i en DevContainer enklast?
    - DevBox?
- Enkelt för framtida utvecklare att börja utveckla
  - Installationskommando
    - curl -sL cli.simonbrundin.com | bash
  - dev commando
    - dev up - lista där jag kan välja projekt
- Enkelt Git-workflow
  - PR skapar ny testmiljö med länk

### Badrum

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
