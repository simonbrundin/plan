# Mål att nå innan jag går över till Plan 1.0

## Utvecklingsmiljö

- [ ] `dev`
  - [ ] Kör utvecklingsmiljö i kubernetes
    - [ ] DevContainer körs i kubernetes
      - [x] Installera Kubernetes provider i Devpod
      - [ ] devpod up github.com/simonbrundin/plan
      - [ ] devpod up github.com/ditt-användarnamn/ditt-repo@min-coola-branch
    - [ ] Notis i telefonen som öppnar en länk där jag kan se min applikation i
          Utvecklingsmiljön.
    - [ ] NeoVim används i DevContainer
      - [ ] Förstå hur det fungerar när man använder NeoVim
      - [ ] Använd AI för att förså hur officiella dokumentationen föreslår att
            man ansluter med NeoVim
    - [ ] DevContainer kör Tilt
- [ ] Skriv om cli så min update talos config inte sparar generated som en fil
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
  - [ ] Lycktan moterad
    - [ ] Kilen passar
    - [ ] Rör nedkortat
    - [ ] Kablarna på utsidan skalade
  - [ ] Lampknappar monterade
  - [ ] Eluttag monterat
  - [ ] Inkopplat i central
  - [ ] Lock på plats
- [ ] Renault igång
- [ ] Dockhus klart

### Talos Cluster

- [ ] auth.simonbrundin.com fungerar
  - [ ] Få alla mina noder att fungera
  - [ ] Uppgradera alla till senaste talos och Kubernetes
- [ ] Installera Orange Pi:s

#### Senare

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
