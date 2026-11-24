---
title: Skogsduvan: Elementet styrs av minsta temperaturen (fail-open)
type: note
tags: [home-assistant, skogsduvan, automatisering, session, projekt:simon]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]
---

2026-11-03: Användaren ville att Elementet (Shelly "Element", switch.element) ska slås av NÄR det kallaste värdet bland de sensorer som svarar går över 5 grader - inte kräva att ALLA rum har värde. Det var ett medvetet val bort från min fail-closed-design.

VAD SOM ÄNDRADES I automations.yaml:
- "Över 5 grader överallt" heter nu "Över 5 grader - elpannorna av" (entity_id oförändrad: automation.over_5_grader_overallt). Triggas nu pa sensor.hemmet_kallaste_rum via template-trigger (> 5) i stallet for binary_sensor.alla_rum_over_5_grader.
- switch.element lades till i BADE listorna: av-listan (over 5) och pa-listan i "Kriskallt" (min < 3). Elementet hade aldrig varit styrd av nagon automation for - det fanns bara som temperatursensor i gruppen.
- Villkoret input_select.heating_mode == Kallt ligger kvar (i "Kallt" lage styr varmen manuellt).

FALLET MED PRISET (fail-open):
- Ett rum med trasig sensor far INTE varma sig nar de fungerande rummen ar varma. Det galler just nu Elementrummet, vars egen sensor (sensor.element_temperatur) ar dod - dess element styrs alltsa av de andra rummens temperatur.
- Den fail-closed binary-sensorn (binary_sensor.alla_rum_over_5_grader) finns kvar for overvakning men styr inte langre nagon avstangning. Den ar unavailable sa lange Element + Tvaattstall ar doda.
- switch.element ar offline (Shelly 192.168.189.131 svarar inte), sa av-kommandot misslyckas tills enheten ar tillbaka - det loggas som "device communication error", det ar forvantat.
- Elementet och Tvaattstall ar de enda tva temperatursensorer i gruppen som INTE svarar just nu (2026-11-03). Tvaattstalls beacon har varit dod sedan 2026-10-05, Element-Shellyn har inte svarat sedan den upptäcktes.

## Relaterade

- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]

