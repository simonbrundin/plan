# Mål att nå innan jag går över till Plan 1.0

## Utvecklingsmiljö

- [ ] Tilt startar i mitt kluster
  - [ ] Hur ansluter jag mot mitt kluster hemma?
    - [ ] Twingate fungerar i mitt kluster och jag kan ansluta mot det från min
          dator och min iPhone
      - [ ] Youtubea videor hur Twingate fungerar och specifikt om kubernetes
            operator för Twingate
- [ ] Välj verktyg och rita en överblick
  - [ ] Tilt
  - [ ]
  - [ ] DevContainer
  - [ ] Git Credentials Manager
- [ ] Tiltfile
  - [ ] Skapar en http-route
- [ ] Mobilikoner för dev och prod fungerar
  - [x] Skapa appikon för bla PWA
  - [x] Prod-ikon
  - [ ] Ikon på iPhone där jag kan se hur det ser ut just nu i dev
    - [ ] Hur får jag urlen?
    - [ ] Startsätt?
      - [ ] devpod up github.com/simonbrundin/plan
      - [ ] devpod up github.com/ditt-användarnamn/ditt-repo@min-coola-branch
    - [x] Hur laddar jag ner den?
      - [x] Enklast kanske är att skapa en länk som jag kan lägga till som en
            PWA
    - [ ] Hur löser jag branching?
    - [ ] NeoVim används i DevContainer
      - [ ] Förstå hur det fungerar när man använder NeoVim
      - [ ] Använd AI för att förså hur officiella dokumentationen föreslår att
            man ansluter med NeoVim
    - [ ] DevContainer kör Tilt
    - [ ] Installera Kubernetes provider i Devpod

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
  - Git Credentials Manager
    - https://www.vcluster.com/blog/using-devpod-with-private-git-repositories
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

## Laptop

- [x] Få kopiera och klista in att fungera
- [ ] Fundera igen om det inte är bättre att köra chezmoi iställer för stow.
      Fördelen jag ser är att jag med ett kommando kan installera allt jag
      behöver på datorn utan att den har något installlerat på sig.

## Höstmixen

- [ ] Fixa tjej till Viktor
- [x] Få in Emma och Backlund
- [x] Ring Häggman

### Badrum

### Hemfix

- [ ] Renault igång
- [ ] Dockhus klart

### Talos Cluster

- [ ] auth.simonbrundin.com fungerar
  - [ ] Få alla mina noder att fungera
- [ ] Installera Orange Pi:s

#### Senare

- [ ] Robotdammsugare fungerar
- [ ] Standard extensions i Chromium
  - [ ] Vimium
  - [ ] 1Password
  - [ ] Video Speed Controller
  - [ ] Video Speed Controller
- [ ] Linkding för att hantera bokmärken i min Homelab så jag kan byta browser
      utan att all mina bokmärken försvinner utan att all mina bokmärken
      försvinner

#### Senare -----------------------------------------------------------------

- [ ] Robotdammsugare fungerar
- [ ] Standard extensions i Chromium
  - [ ] Vimium
  - [ ] 1Password
  - [ ] Video Speed Controller
  - [ ] Video Speed Controller
- [ ] Linkding för att hantera bokmärken i min Homelab så jag kan byta browser
      utan att all mina bokmärken försvinner utan att all mina bokmärken
      försvinner

## Klart --------------------------------------------------------------------

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

## Dennis

- [x] Gör löner
- [x] Betala ut friskvård
