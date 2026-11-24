---
title: Skogsduvans HA: temperaturlogik ombyggd, fyra döda sensorer
type: note
tags: [home-assistant, skogsduvan, session, automatisering, projekt:simon]
created: 2026-11-02
updated: 2026-11-03
status: active
related: [notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet, notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]
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
- automations.yaml (första omgången, senare rättad - se nedan): '5 grader - varmeflakt av' triggar pa min-sensorn; 'Over 5 grader overallt' triggar pa binary_sensorn (state to on); 'Kriskallt' triggar pa min-sensorn below 3.
- Jinja-fallsticker jag faktiskt ramade in: rejectattr() ar en generator - att anvanda den tvA ganger (map + count) gav count 0 -> villkoret blev falskt. Måste | list materialiseras. Testa därfor alltid template lokalt med mockade states.
- Backup av originalen: /homeassistant/config_backups/20261102-223946/

KLART (2026-11-02, kväll):
- HA Core omstartad -> group.hemtemperaturer, sensor.hemmet_kallaste_rum (16.1-16.3 °C) och binary_sensor.alla_rum_over_5_grader (unavailable, fail-closed) finns i recordern.
- min_max 'Snittemp - Inne': sensor.tado_temperatur borttagen genom redigering av .storage/core.config_entries (sed pa JSON-raden) + omstart -> snittet räknas nu på 3 sensorer (utbyggnad unavailable räknas inte in).
- HA Core uppdaterad 2026.10.1 -> 2026.10.4 ('ha core update' i add-on-containern). Automations, template-sensor och min_max fungerar efter uppdateringen.

FALLET (viktigt, kostade tid):
- Template-entitetens entity_id genereras från NAMNET, inte från unique_id. "Alla rum över 5 grader" blev därför binary_sensor.alla_rum_over_5_grader (_grader-suffixet) - min automationen refererade till binary_sensor.alla_rum_over_5 hade den ALDRIG triggat. Läs alltid det faktiska entity_id:t ur .storage/core.entity_registry (jq) efter att en template lagts till.
- En sak som inte fungerade: att kopiera home-assistant_v2.db + -wal och sedan kolla i SQLite. En kopieringsrace mot WAL-checkpoint gjorde att de nya entiteterna saknades i states_meta och jag jagade flera timmar fel spår (trodde att template-integrationen var trasig). Testa alltid mot en FRISK kopia och kom ihåg att wal + db ska kopieras atomiskt (t.ex. via sqlite3 .backup eller pa en stillastående fil).
- Det går inte att läsa HA:s live-state via API från add-on-containern: supervisor-token -> 401 pa /core/api/*, och .storage/auth-token är hashade. Använd istället recordern (eller en shell_command-automation) för att verifiera.

NYTT FALLET (2026-11-02 kvall, rättad efter granskning av originalet):
- Aliaset '5 grader - varmeflakt av' var MISSLYDANDE. Originalets handling var INTE att stanga av en flakt utan att kora script.8_grader_via_zha_ir_duplicera_duplicera, dvs. satta VAERMEPUMPEN (ZHA TS0001, ieee e4:56:ac:ff:fe:76:6f:49) till 8 grader. Jag hade i forra omgangen gjort om den till 'stang av varmeflakt_koksbank under 5 grader' - det var bakvant (stanger AV en varmekalla nar huset kyls) och krockade med Kriskallt (<3) som tander samma switch. RATTAT 2026-11-02: handlingen ar nu varmepumpen till 8 grader + switch.flakt av.
- IDENTITET: switch.varmeflakt_koksbank tillhor ZHA-enheten 'Varmekabel - Kosbank' (TS011F) - det ar en VARMEKABEL i diskbanken, inte en flakt. Bevis: den stangs av i 'Over 5 grader overallt' ihop med minielement_kosbank, minielement_garderob och golvvarme (alla elpannor), och tands ihop i Kriskallt (<3). Den riktiga flakten ar switch.flakt = Shelly 'Flakt' (Zbeacon? nej: Shelly 1, config entry '4372', 192.168.189.106) = takflakt i badrummet som drar luft fran loftet for att cirkulera luften i huset. Den har varit OFFLINE i hela recorder-fonstret.
- Automationens entity_id heter automation.under_5_grader (UI-omdopning) - slaget pA aliaset ('5 grader - varmeflakt av') hittar ingenting i registret.
- Varfor stugans varmekassa ska ga 7 -> 8 grader: kaskaden ar minsta temperatur -> >=5: pump 7 grader (script.7_grader_via_zha_ir), alla elpannor av; <5: pump 8 grader + flakt av; <3: pump 8 grader + ALLA elpannor pa (minielement x2, varmekabel, golvvarme) + larm. Sista steget 'Over 5 grader overallt' kraver binary_sensorn (fail-closed) och mode Kallt, sa det kan inte flaxa mot <5-steget (min<5 och alla>5 ar mutually exclusive).
- Bevis att action-formen script.8_grader_via_zha_ir_duplicera_duplicera fungerar: Kriskallt triggade 2026-11-01 04:11 med exakt den actionen (script-namn == YAML-nyckel, oavsett UI-entity_id).
 `ha` finns i Terminal&SSH-add-on-containern: ha core check | restart | update | info | logs. Det gör hela deploy-cykeln (validera -> starta om -> verifiera i recordern) utan att röra UI:t.

## Relaterade

- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]
- [[notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet]]
