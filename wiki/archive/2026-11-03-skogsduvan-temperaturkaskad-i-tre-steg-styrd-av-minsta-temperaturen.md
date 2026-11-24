---
title: Skogsduvan: temperaturkaskad i tre steg, styrd av minsta temperaturen
type: note
tags: [home-assistant, skogsduvan, automatisering, session, projekt:simon]
---

Kaskad som byggdes 2026-11-02 (kvall) i Skogsduvans HA, styrd av EN sanningskalla: template-sensorn sensor.hemmet_kallaste_rum (minsta temperaturen bland de sensorer som svarar).

STEG (alla pa minsta temperaturen):
1. min >= 5: pump 7 grader (script.7_grader_via_zha_ir), elpannor av. Triggas av 'Tillbaka till 7 grader' (min over 5) + condition input_select.heating_mode == Kallt.
2. min < 5: pump 8 grader (script.8_grader_via_zha_ir_duplicera_duplicera) + takflakt i badrummet AV (switch.flakt). Automation '5 grader - varmeflakt av' (entity_id automation.under_5_grader - UI-omdopning).
3. min < 3: pump 8 grader + ALLA elpannor PA (minielement x2, varmekabel, golvvarme) + push-notis till bada telefoner. Automation 'Kriskallt'.

Varfor delningen: pumpens 7->8-graderssteg och flaktens AV/Pa styrs av MIN-sensorn (robust - fungerar med trasiga sensorer), medan ELPANNORNA AV fortfarande kräver att ALLA rum har varde (binary_sensor.alla_rum_over_5_grader, fail-closed). Motivet: att stanga av elpannorna nar vi inte vet att ett rum med trasig sensor inte är kallt kan ge frost; att sanka pumpen 8->7 nar det kallaste rummet ar over 5 ar bara en energifrakga. Flaktens 'tillbaka'-steg har villkoret switch.flakt == off, sa en manuellt avstangd flakt startar inte av sig sjalv och en offline flakt ger inget fel.

VIGTIGT - HA:s numeric_state triggar BARA VID KORSNING (verifierat mot kalla 2026.10.x): koden har en armed_entities-mangd - vid start markeras en entity som 'armad' bara om den INTE redan matchar kriteriet. En automation 'above: 5' pa en sensor som redan star over 5 triggar alltsa INTE vid omstart - forst nar vardet korsar ner under 5 och sedan upp over 5 igen. Detta galler alla tre stegen (och forklarar varfor 'Kallt lage' + omstart inte satte pumpen till 7 direkt). Har man vill ha 'satta alltid vid start' maste man anvanda en template-trigger eller ett homeassistant-start-event.

LAVST: att lasa HA:s kalla pa GitHub for att forsta triggermallsemantik; kalla: homeassistant/components/homeassistant/triggers/numeric_state.py (armed_entities + 'Each entity that starts outside the range is already armed').

KOMMANDE TIP: 'ha backups new --name X' i add-on-containern skapar en FULL backup (utan --app/--folders). Supervisorns REST-svarar 405 pa POST /backups/full och /backups/new - anvand CLI:t. Full backup 166 MB finns nu (2026-11-03T00:06Z). For att redigera .storage/core.config_entries maste HA vara STOPPAD (ha core stop) annars skriver HA tillbaka filen med den in-memory-versionen.
