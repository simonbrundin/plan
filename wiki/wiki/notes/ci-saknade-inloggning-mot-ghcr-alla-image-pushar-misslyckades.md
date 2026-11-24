---
title: CI saknade inloggning mot GHCR - alla image-pushar misslyckades
type: note
tags: [bug, ci, ghcr, projekt:stk]
created: 2026-11-06
updated: 2026-11-06
status: active
---

Sedan repot skapades (3f33ab1) misslyckades jobbet 'Bygg och pusha image' med 'failed to fetch anonymous token ... 403 Forbidden'. Orsak: .github/workflows/ci.yaml saknade steget docker/login-action, sa pushen till ghcr.io skickades utan GITHUB_TOKEN. Ingen image (nuxt-app, go-api, rust-api) har publicerats sedan dess, och jobbet 'Uppdatera prod-images' har aldrig korrekt satta taggarna. Fix: commit c89a698 lagt till login-action med registry ghcr.io, username github.actor och secrets.GITHUB_TOKEN. CI blev gron direkt efter fixen. Lardom: kontrollera att push-steg har autentisering - ett 'anonymous token'-fel pa ghcr.io betyder saknad inloggning, inte saknad package-behorighet.
