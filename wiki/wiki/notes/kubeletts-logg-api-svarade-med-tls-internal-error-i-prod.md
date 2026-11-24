---
title: Kubeletts logg-API svarade med tls: internal error i prod
type: note
tags: [diagnostik, talos, kubernetes, projekt:infrastructure]
created: 2026-11-24
updated: 2026-11-24
status: active
---

I prod-klustret gick det inte att läsa poddarnas loggar med 'kubectl logs' på noderna worker-4, worker-5, worker-6 och worker-7. Felet var 'remote error: tls: internal error' från apiserverns proxy mot kubeletten. På worker-3 fungerade det, så problemet var nodspecifikt, inte klustergemensamt.

Lärdom: när loggarna inte går att läsa, flytta podden till en nod där logg-API:t svarar innan man felsöker i workloaden. Jag cordonerade worker-5, worker-6 och worker-7 och lät source-controller landa på worker-3, och då kom loggarna.

Diagnostiken blev mycket snabbare när jag dessutom kunde läsa klusterloggar via poddar: jag körde kommandon i en podd och skrev resultatet till /dev/termination-log, eftersom poddar på trasiga noder inte kan läsas med kubectl logs.
