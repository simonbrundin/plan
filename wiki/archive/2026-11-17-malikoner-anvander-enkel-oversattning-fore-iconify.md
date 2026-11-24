---
title: Målikoner använder enkel översättning före Iconify
type: note
tags: [feature, icon, translation, goals, projekt:plan]
---

Lösningen använder npm-paketet translate i Nuxt-serverrouten /api/icons. Vid automatisk målikonsökning skickas måltexten från svenska till engelska via translate-paketets Google-engine, därefter söks Iconify på hela den översatta texten och enskilda engelska ord. IconPicker påverkas inte och söker fortfarande exakt på användarens text. Klienten skickar translate=true endast när ett nytt mål skapas.
