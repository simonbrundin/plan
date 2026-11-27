---
title: Skogsduvan: värmepumpen styrs från högra sidofältet (IR + trappa 5-7-17-20-24)
type: note
tags: [home-assistant, skogsduvan, session, automatisering, projekt:simon]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen, notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]
---

2026-11-27: Byggde ett värmepump-kort i högra sidofältet ("Vår vy" -> vyn "Standard") i Skogsduvans HA.

VAD SOM LAGDES TILL
- configuration.yaml: input_select.varmepump_temperatur (options 5,7,17,20,21,22,23,24) och
  input_select.varmepump_flakt (Låg/Medium/Hög).
- scripts.yaml: varmepump_ir_kod (räknar ut IR-koden ur temp+läge och kör den),
  varmepump_hoj / varmepump_sank (ett steg upp/ner i trappan), varmepump_flakt_lag/_medium/_hog.
- .storage/lovelace.var_vy: fyra kort med view_layout position=sidebar (display, -/+ , lägesknappar, av-knapp).

VAD JAG LÄRDE MIG (viktigt)
- Script-tjänsten heter YAML-NYCKELN, inte entity_id:t. Registret har entity_id:n script.5_grader_via_zha_ir
  men nyckeln är 8_grader_via_zha_ir_duplicera_2. Därför anropas skript via generic script.turn_on
  med entity_id i target - det fungerar och är lättare att lasa.
- 22 grader hog heter script.22_grader_high (inte script.22_high) - enda avvikelsen i namnschemat.
- IR-koderna ar enkelriktade: HA vet aldrig vilket setpoint pumpen star pa. input_selecten visar
  darfor det SISTA varde vi skickat, inte pumpens faktiska lage.
- entity_id for input_select kommer av NAMNET, inte nyckeln: "Varmepump flaktlage" ->
  input_select.varmepump_flakt (inte ..._flaktlage). Verifiera alltid entity_id:t via /api/states.
- Template-mal i ett scripts target.entity_id funkar (testat), men via REST-API:t pa
  /api/services/<dom>/<service> renderas mallar INTE (400 Bad Request) - mal maste renderas i
  config/automation, inte i API-anropet.
- persistent_notification.create skapar INGA entities i HA 2026.11 (saknar i /api/states och i
  registret) - anvand input_text for att verifiera via API.

MISSLYDANDE KOD I KASKADEN (hittade, inte fixat)
- Automationen "5 grader - flakten av" anropar script.8_grader_via_zha_ir_duplicera_duplicera.
  Den nyckeln tillhord script.17_grader_via_zha_ir -> steget satter 17 GRADER, inte 8 som
  beskrivningen sager. Det finns ingen 8-graderskod; koderna ar 5, 7, 17 och 20-24.
  Innan jag laste registret trodde jag att koden var trasig - det var bara jag som missade skillnaden
  mellan nyckel och entity_id.

VERIFIERING (utan att skicka IR)
- Matrisen 8 temperaturer x 3 flaktlagen (24 kombinationer) raknades ut via ett temporart
  test-script och en input_text - alla 24 landade i ratt befintligt skript.
- Trappan upp/ner testad for varje steg (5->7->17->20->21->22->23->24).
- Inget IR-signal har skickats under arbetet - anvandaren ska gora det forsta riktiga trycket.

Relaterat: [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]],
[[notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]]
