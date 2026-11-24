---
title: Skogsduvans HA: temperaturlogik ombyggd, fyra döda sensorer
type: note
tags: [home-assistant, skogsduvan, session, automatisering, projekt:simon]
---

Kontext: HA Green (2026.10.1) i stugan, näs via 'ssh skogsduvan' -> Terminal&SSH-addon-container (core-ssh), config i /homeassistant (/config). Config-filer: configuration.yaml (har nu template:-sektion), groups.yaml (ny), automations.yaml, scripts.yaml.

FOUND (2026-11-02):
- 4 av 8 temperatursensorer i 'alla rum'-listan var doda: sensor.tado_temperatur finns INTE (Tado-integrationen har config entry men 0 devices/entities), sensor.termometer_utbyggnad_temperatur unavailable sedan 2026-11-01 (Zbeacon a4:c1:38:33:c9:02:dc:b3, batteri), sensor.termometer_tvattstall_temperatur Zbeacon a4:c1:38:30:4e:d4:97:46 sedan 2026-10-05, sensor.element_temperatur (Shelly 'Element' E8DB84D529BA) rapporterade aldrig.
- Automation '5 grader - varmeflakt av' hade 'attribute: setting' i numeric_state-trigger -> triggar pa ett attribut ingen temperatursensor har -> ALDRIG triggade.
- Automation 'Over 5 grader overallt' anvande en template-trigger med 'float' UTAN default pa unavailable -> Template error -> kunde aldrig bli sann (bara manuellt utlost en gang, 2026-11-02 20:48).
- min_max-hjalparen 'Snittemp - Inne' innehaller sensor.tado_temperatur (finns inte) -> snittet raknades pa 2 av 4.

INSIGHT (viktigt, misstag jag gjorde först):
- HA:s script-integration: entity_id == YAML-nyckeln (entity.py: self.entity_id = ENTITY_ID_FORMAT.format(key)), och tjänstenamnet == unique_id == nyckeln. Men entity_id kan DÖPAS OM i UI:t (t.ex. script.7_grader_via_zha_ir) medan nyckeln ligger kvar (script.8_grader_via_zha_ir_duplicera_3). Därför skiljer sig entity-registret fran nycklarna utan att nagon fil var trasig. Verifierade IR-koderna byte-for-byte mot loggade script_started/call_service-event: 19 koder pa disk == det som kor. Automations anropar ibland nyckeln (script.<key>) och ibland entity_id:n (target:), och BADA fungerar.
- Bevis att ingen reload skett sedan 30 sep: registry modified_at for de omdopta skripten ligger kvar pa 29 sep, och .HA_VERSION skrevs 10 okt utan att entity_id:n backades -> registret vinner over koden explicita entity_id.

LOSNING (2026-11-02, deployad):
- Ny grupp group.hemtemperaturer i groups.yaml (enda listan; bastu + tado medvetet utelamnade).
- configuration.yaml: template-sensor 'Hemmet - kallaste rum' (min over giltiga sensorer, availability = minst en giltig -> frysrisklarmet fungerar trots doda beacons) + binary_sensor 'Alla rum over 5 grader' (fail-closed: availability kräver att ALLA gruppmedlemmar har varde).
- automations.yaml: '5 grader - varmeflakt av' triggar pa min-sensorn och slar bara av switch.varmeflakt_koksbank; 'Over 5 grader overallt' triggar pa binary_sensorn (state to on); 'Kriskallt' triggar pa min-sensorn below 3.
- Jinja-fallsticker jag faktiskt ramade in: rejectattr() ar en generator - att anvanda den tvA ganger (map + count) gav count 0 -> villkoret blev falskt. Måste | list materialiseras. Testa därfor alltid template lokalt med mockade states.
- Backup av originalen: /homeassistant/config_backups/20261102-223946/

KVAR ATT GORA (2026-11-02): ladda om group/template/automation (eller starta om HA Core) sa att group.hemtemperaturer och sensor.hemmet_kallaste_rum finns; ta bort sensor.tado_temperatur ur min_max 'Snittemp - Inne' (UI); fysisk tillsyn av Shelly Element + batteri-beacons for Utbyggnad/Tvattstall. HA 2026.10.4 finns att uppdatera till.
