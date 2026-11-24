---
title: Skogsduvan: temperaturkaskad i tre steg, styrd av minsta temperaturen
type: note
tags: [home-assistant, skogsduvan, automatisering, session, projekt:simon]
created: 2026-11-03
updated: 2026-11-17
status: active
related: [notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online, notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02, notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]
---

Kaskad som byggdes 2026-11-02 (kvall) i Skogsduvans HA, styrd av EN sanningskalla: template-sensorn sensor.hemmet_kallaste_rum (minsta temperaturen bland de sensorer som svarar).

ENTITETERNAS FYSISKA TILLHORSIGHET (2026-11-03, viktigt - flera entity-id:n ar relikter):
- switch.varmeflakt_koksbank = ZHA-pluggen som nu heter 'Varmekabel - Kosbank (gammal varmeflakt)'. Entity-id:t kommer fran da pluggen var en varmeflakt i koken - ZHA byter aldrig entity-id vid omdopning. Anvandaren bytte ut varmeflakten mot ett minielement under diskbanken.
- switch.minielement_koksbank = minielementet under diskbanken (ZHA 'Minielement - Kosbank')
- switch.minielement_garderob = minielementet i garderoben
- switch.flakt = Shelly 'Flakt' (badrum takflakt, cirkulerar luft fran loftet)
- switch.element = Shelly 'Element' (elementrummet)
Aliaset '5 grader - varmeflakt av' bytte namn till '5 grader - flakten av' 2026-11-03 (UI-namnet var redan 'Under 5 grader nagonstans').

STEG (alla pa minsta temperaturen):
1. min >= 5: pump 7 grader (script.7_grader_via_zha_ir), elpannor av. Triggas av 'Tillbaka till 7 grader' (min over 5) + condition input_select.heating_mode == Kallt.
2. min < 5: pump 8 grader (script.8_grader_via_zha_ir_duplicera_duplicera) + takflakt i badrummet AV (switch.flakt). Automation '5 grader - varmeflakt av' (entity_id automation.under_5_grader - UI-omdopning).
3. min < 3: pump 8 grader + ALLA elpannor PA (minielement x2, varmekabel, golvvarme) + push-notis till bada telefoner. Automation 'Kriskallt'.

Varfor delningen: pumpens 7->8-graderssteg och flaktens AV/Pa styrs av MIN-sensorn (robust - fungerar med trasiga sensorer). ELPANNORNAS AV styrdes till en borjan av binary_sensor.alla_rum_over_5_grader (fail-closed: alla rum maste ha varde), men 2026-11-03 bytte anvandaren om till MIN-sensorn (fail-open) och Elementet lades in i bade av- och pa-listan - se [[notes/skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open]]. Priset ar att ett rum med trasig sensor inte far varmare nar de fungerande rummen ar varma. binary-sensorn finns kvar for overvakning men styr inte langre nagon avstangning. Flaktens 'tillbaka'-steg har villkoret switch.flakt == off, sa en manuellt avstangd flakt startar inte av sig sjalv och en offline flakt ger inget fel.

STARTVAG: 2026-11-03 fick varje temperatursteg en extra vag - Home Assistants start-event - sa att laget aterstalls efter en omstart eller stromavbrott (90 s fordropning sa att alla entities hinner aterstallas). Se [[notes/ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor]].

VIGTIGT - HA:s numeric_state triggar BARA VID KORSNING (verifierat mot kalla 2026.10.x): koden har en armed_entities-mangd - vid start markeras en entity som 'armad' bara om den INTE redan matchar kriteriet. En automation 'above: 5' pa en sensor som redan star over 5 triggar alltsa INTE vid omstart - forst nar vardet korsar ner under 5 och sedan upp over 5 igen. Detta galler alla tre stegen (och forklarar varfor 'Kallt lage' + omstart inte satte pumpen till 7 direkt). Samma sak galler template-triggern: dess template renderas redan 'sant' nar automationen satts upp, sa ingen flipp sker - losningen var start-eventet + fordropning (se startvagen ovan).

LAVST: att lasa HA:s kalla pa GitHub for att forsta triggermallsemantik; kalla: homeassistant/components/homeassistant/triggers/numeric_state.py (armed_entities + 'Each entity that starts outside the range is already armed').

KOMMANDE TIP: 'ha backups new --name X' i add-on-containern skapar en FULL backup (utan --app/--folders). Supervisorns REST-svarar 405 pa POST /backups/full och /backups/new - anvand CLI:t. Full backup 166 MB finns nu (2026-11-03T00:06Z). For att redigera .storage/core.config_entries maste HA vara STOPPAD (ha core stop) annars skriver HA tillbaka filen med den in-memory-versionen.

## Relaterade

- [[notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02]]
- [[notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online]]
- [[notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]]
- [[notes/skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open]]
- [[notes/ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor]]
- [[notes/skogsduvans-ha-utsatts-for-automatiserade-skannningar-fran-internet]]
