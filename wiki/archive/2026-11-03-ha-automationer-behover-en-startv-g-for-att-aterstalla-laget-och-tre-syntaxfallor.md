---
title: HA: automationer behöver en startvág för att återställa läget (och tre syntaxfällor)
type: note
tags: [home-assistant, session, projekt:simon]
---

Problem som upptäcktes 2026-11-03 i Skogsduvans HA: en temperatureautomation som styrs av minsta temperaturen ska "äga" rätt tillstånd även efter omstart - men varken numeric_state eller template-trigger triggar vid start.

VARFÖR INTE:
1. numeric_state: HA markerar en entity som "armad" bara om den INTE redan matchar kriteriet vid start (armed_entities i homeassistant/components/homeassistant/triggers/numeric_state.py). Ett 'above 5' triggar da aldrig nar startvardet redan ligger over 5.
2. template-trigger: triggerns template renderas NER en gang nar automationen satts upp - och da har HA redan atterstatt sensorns varden. Startlagen ar alltsa redan "sant", sa det sker aldrig någon flipp och triggern gar aldrig.

LOSNING (verifierad fungerande 2026-11-03 10:52 UTC):
- Varje temperatursteg har TVAI triggare: en template-trigger (id: temp) for korsningar, och homeassistant-start-eventet (id: start).
- Startvagen gor INTE villkorskontrollen direkt - den vantar 90 sekunder (delay i en choose som matchar trigger.id == start) sa att alla entities ar atterstallda, och lopar sedan samma villkor inne i aktionen (choose) med de riktiga vardena.
- MAL: gora inte villkor pa automationsniva for temperaturberoende villkor - de evaluated for tidigt. Lag dem i aktionens choose.
- Resultat: efter varje omstart skickas 7/8-graderskommandot och elpannorna stangs/satts pa utifran det faktiska laget, istallet for att ligga kvar i det gamla.

TRE SYNTAXFALLER SOM KOSTADE TID:
1. homeassistant-triggern vill ha "event: start" (en strang). "event: {event_type: start}" ger felet "expected 'start' or 'shutdown' at 'event'. Got None".
2. Villkoret "condition: trigger" med id-lista maste indenteras DJUPARE under condition - annars blir "id" ett syskon till "condition" och felet blir "not a valid option at ...choose[0].id".
3. 'ha core check' FANGAR INTE schema-fel i automations.yaml - det returnerar exit 0 medan automationerna disables i loggen. Kolla alltid loggen pa "failed to setup triggers" / "could not be validated" efter en konfig-andring.

OPS: deploya automations.yaml atomiskt - skriv till automations.yaml.new och mv over originalet. HA lampar annars filen medan den skrivs och alla automationer blir unavailable (hade detta 2026-11-03). Kopiera av home-assistant_v2.db + -wal kan ge "database disk image is malformed" - forsok igen da det intrarffar.
