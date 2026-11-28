---
title: PR 179 och definitionsassistenten krockar i sex filer
type: note
tags: [git, merge, konflikt, wiki, projekt:plan]
---

VAD: nar PR 179 (fasmaskin och verifiering) mergades till main, och det parallella arbetet med definitionsassistenten sedan applicerades ovanpa, uppstod konflikter i sex filer:

- environments/base/db/schema.yaml
- src/api/internal/database/db.go
- src/api/internal/models/goal.go
- src/api/internal/repository/goal_repository.go
- src/nuxt/app/pages/goal/[id].vue
- src/rust/crates/plan-api/src/routes/table.rs

VARFOR: bada arbetena tog den vanliga formen - samma modeller, samma schema, samma route-tabell. Det är inte ett fel i nagon implementation, det ar att tvA strangar i samma repo gick parallellt. Det hande for att arbetet pa fasmaskinen gjordes i en ren worktree fran origin/main medan definitionsassistenten arbetade i huvudrepot.

SÅ SER DET UT: konflikterna ar semantiska men sma - i models/goal.go lagger fasmaskinen in CompletionKind medan assistenten lagger in Criteria, och i route-tabellen lagger fasmaskinen in fas- och verifieringsvagar medan assistenten lagger in definition-vagar. Losningen ar att behalla bada fälten och bada vagarna.

GALLER FRAMOT: innan nasta stora arbetet startar - fraga om nagon annan session ar igang i samma repo, eller anvand en egen worktree fran origin/main och en egen branch, som det gjordes for fasmaskinen. När en merge sker ska arbetsstromarna slas ihop medvetet, inte av en slump.
