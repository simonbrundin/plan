---
title: Transienta 502 fra Go API:t nar poddar startar om i dev-klustret
type: note
tags: [bug, frontend, go, kubernetes, projekt:stk]
created: 2026-11-06
updated: 2026-11-06
status: active
---

I stk:s dev-klustret (namespace repo-template) proxar nginx i nuxt-imageen /api/ -> go-api-service:8080/. Nar Tilt/kubectl byter om go-api-podden svarar nginx 502 och frontendens fetch() kastar - anvandaren ser bara ett generiskt 'Kunde inte hamta'. Fix i stk: composablen fetchAppointments() forsoker om upp till 3 ganger vid nativverksfel eller 5xx (1s, 2s backoff) och sidan visar det verkliga meddelandet (HTTP-status + URL) plus en 'Forsok igen'-knapp. Viktig regel for alla frontend-sidor mot /api/: visa status + url i felmeddelandet, aldrig en generisk text - annars kan man inte felsoka en 502 fran nginx.
