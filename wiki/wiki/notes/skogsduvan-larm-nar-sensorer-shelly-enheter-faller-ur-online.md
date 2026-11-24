---
title: Skogsduvan: larm när sensorer/Shelly-enheter faller ur online
type: note
tags: [home-assistant, skogsduvan, automatisering, session, projekt:simon]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]
---

2026-11-03: Anvandaren bad om push-larm nar en enhet i stugan tyst forsvinner.lost igang, sa att en dod sensor inte upptacktes forst nar nagon kommer till stugan.

VAD SOM LAGDES TILL:
- Ny template binary_sensor.stugans_shelly_enheter_online (configuration.yaml): 'on' nar alla sex Shelly-enheter svarar (switch.element, switch.flakt, switch.golvvarme, switch.handdukstork, switch.luftvarmepump, switch.varmvattenberedare), 'off' annars. Namnet ger entity-id binary_sensor.stugans_shelly_enheter_online.
- Ny automation "Utanfor online - larm" (automation.utanfor_online_larm) med fyra state-triggers:
  1. binary_sensor.alla_rum_over_5_grader fran on/off till unavailable -> "en eller flera temperatursensorer svarar inte"
  2. samma sensor fran unavailable till on/off -> "alla temperatursensorer svarar igen"
  3. binary_sensor.stugans_shelly_enheter_online fran on till off -> "en eller flera Shelly-enheter svarar inte"
  4. samma sensor fran off till on -> "alla Shelly-enheter svarar igen"
- Notis till bade telefoner (notify.mobile_app_michaels_iphone / evas_iphone).

VARFORE INTE "homeassistant"-startvag HÄR: larmet ska bara skickas vid byte, inte vid varje omstart. Vid omstart gar enheterna unknown -> off/unavailable, vilket INTE ar bytena on/off -> unavailable, sa starten spampar inte. Kontrollerat: automationens last_triggered var tomt efter omstart.

FALLGROP: Jinja-scoping. I en for-loop gor "{% set missing = missing + [d] %}" INTE variabeln synlig utanfor loopen - mallen returnerade alltsa alltid 'on'. Losningen ar {% set ns = namespace(missing=[]) %} + {% set ns.missing = ns.missing + [d] %}. Detta ar samma farlighet som den med rejectattr() som behovde | list for att materialiseras.

LAGE JUST NU (2026-11-03): binary_sensor.stugans_shelly_enheter_online = off (Element, Flakt, Luftvarmepump, Varmvattenberedare svarar inte) och binary_sensor.alla_rum_over_5_grader = unavailable (Element + Tvaattstall saknas). Larmet gor alltsa inget tills nagot kommer tillbaka - DAB larmar nar Elementet eller Tvaattstall-sensorn kommer igang.

## Relaterade

- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]

