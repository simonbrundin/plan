---
title: MATCHi Keycloak: session_code måste fångas med generell teckenklass
type: note
tags: [bug, matchi, keycloak, lesson, projekt:tmp]
---

Symtom: Go-klientens inloggning mot MATCHi (auth.matchi.com) misslyckades intermittent - ca 40% av försöken - med felet 'inloggning misslyckades, hamnade på auth.matchi.com'. Samma flöde fungerade i Python.

Rotorsak: regexen för session_code var ([0-9a-zA-Z\-]+). Keycloaks sessions-id kan innehålla _ och ., så värdet kapades mitt i och Keycloak avvisade det. Det fungerade bara när _ råkade ligga sent i värdet.

Fix: fånga värden med ([^&"'\s]+) - samma princip som i den fungerande Python-klienten.

Viktig regel: Keycloak-parametrar (session_code, execution, tab_id) får aldrig plockas med en restriktiv teckenklass. Test: go test ./internal/matchi/ -run TestLoginAction.

Påminnelse: när ett liknande Python-flöde fungerar och Go-flödet inte gör det, jämför regexerna och redirect-hanteringen först.
