---
title: HA Tuya Tesvor S6: anslutningen fungerar men statusfält saknas
type: note
tags: [home-assistant, tuya, tesvor, dammsugare, felsökning, projekt:simon]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [entities/home-assistant]
---

Felsökning 2026-11-25 av Krokviken (10.20.20.20): UniFi visar Tesvor S6 som online på 10.20.20.175 (Wi-Fi IoT, MAC d8:c0:a6:75:7a:8d, senaste trafik ökar kontinuerligt). HA Tuya diagnostics för entry 01KC9B3N5WDXMTGXKSPDXNZPVS visar endpoint https://apigw.tuyaeu.com, mqtt_connected=true och Tuya-enheten bf418cae6af7e4e4a0hqwb online=true. Alltså är varken Wi-Fi, routing, MQTT eller Tuya-inloggningen grundfelet.

Enhetens rapporterade Tuya-status innehåller bara: power_go=false, pause=false, switch_charge=true, mode=chargego, seek=false. Den innehåller INTE DP-koden 'status'. Home Assistants Tuya VacuumActivityWrapper letar efter just DP-koden status (eller pause som fallback) för att kunna sätta vacuum-state. Därför blir vacuum.tesvor_s6 state 'unknown' även när enheten är online och står i chargego. Det är en begränsning i Tuya-enhetens datapunkter/mappning, inte ett anslutningsfel. supported_features=13084 innehåller ändå start, return-to-base, pause/locate m.fl.; automationens tjänster kan alltså fortfarande skickas.

HA har dessutom sex Tuya config entries för samma konto/enhet (fem user + en dhcp, flera med samma Tuya user code CxpntbH och EU-endpoint). Alla rapporterar MQTT connected och Tesvor online, men detta är en städskuld: flera parallella Tuya managers kan ge onödig polling och bör rensas genom UI/återautentisering vid ett separat, försiktigt underhållstillfälle. Ta inte bort entries blint: async_remove_entry kan återkalla Tuya-sessionen.

Konsekvens för automationen 'Dammsugare - ga tillbaka till station nar nagon kommer hem': villkoret 'not vacuum state docked/charging' passerar alltid när state är unknown. Det gör att return_to_base försöks varje gång Simon eller Emma kommer hem, vilket är säkrare än att inte skicka kommandot men bör dokumenteras. För verklig status behövs en annan integration/lokal API eller en Tuya device quirk som mappar Tesvor S6:s mode (chargego/smart/...) till vacuum-activity; att bara starta om HA eller Tuya-integrationen löser inte det saknade status-fältet.

## Relaterade

- [[entities/home-assistant]]

