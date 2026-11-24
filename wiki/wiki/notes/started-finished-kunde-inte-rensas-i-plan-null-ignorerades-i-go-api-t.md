---
title: started/finished kunde inte rensas i plan (null ignorerades i Go API:t)
type: note
tags: [bug, plan, api, projekt:simon]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Bug: frontend skickade {started: null} vid avmarkering, men Go-modellen UpdateGoalRequest hade Started *string med omitempty - JSON null gjorde pekaren nil, fältet hoppades över i repository och tidsstämpeln stod kvar. Samma fel för finished. En andra bugg i repository: 'started = NULL' lade ingen parameter i args, så tidigare guard 'len(args)==0' returnerade målet oförändrat. Fix: ny modell NullableTime{Set bool; Value *string} med egen UnmarshalJSON (skiljer absent / null / värde), repository använder hasUpdate-flaggan istället för len(args)==0, ogiltig tidsstämpel ger nu 400 via ErrInvalidTimestamp. Verifierat mot Neon dev-DB och via gin ShouldBindJSON. Verktyg: sourcing av environments/dev/db/neon.env fungerar inte (& i URL:en bryter raden) - använd grep|cut istället.
