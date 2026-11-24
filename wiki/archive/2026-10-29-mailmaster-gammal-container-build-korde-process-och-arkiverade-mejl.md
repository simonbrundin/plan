---
title: MailMaster: gammal container-build körde /process och arkiverade mejl
type: note
tags: [bug, mailmaster, dev-miljo, projekt:mailmaster]
---

Symtom: efter att granskningsflödet (analysera + OK/Neka) implementerats syntes fortfarande den gamla UI:n, och mejl försvann ur listan direkt när 'Processa mejl' klickades - den gamla /process-endpointen utförde arkivering automatiskt.

Varför: docker-composen bygger Go-backen ur en image (Dockerfile-go kopierar källan vid build) och Nuxt-bilden har app/ monterad som volym. Utan ombyggnad/restart av containern kördes fortfarande den gamla koden, och webbläsaren cachade den gamla bundle-filen.

Lärdom: efter ändringar i src/go måste containern byggas om (docker compose build/up), och efter ändringar i src/nuxt/app måste Nuxt-devservern startas om så att den nya bundle-filen laddas. Testa alltid den nya endpointen direkt med curl innan man drar slutsatser om att frontend-fixen inte fungerar.
