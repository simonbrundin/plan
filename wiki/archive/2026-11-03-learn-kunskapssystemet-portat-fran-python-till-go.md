---
title: Learn: kunskapssystemet portat från Python till Go
type: note
tags: [learn, go, feature, session, port, projekt:learn]
---

Användaren ville inte ha Python i repot 'learn'. Hela kunskapssystemet (e-pub-parser, konceptextraktion, kunskapsgraf, quizmotor med SM-2, förklaringar, brainstorm, CLI) är nu skrivet i Go under src/go/internal/{epub,text,knowledge,learning,ideas,store,learn} med CLI i src/go/cmd/learn. Python-delen (src/python) är borttagen.

Varför Go: repot är redan Go/Go-API/CLI-orienterat (AGENTS.md), så en Python-komponent bröt mot projektets konventioner och krävde en egen toolchain.

Viktigaste Go-besluten: SQLite via modernc.org/sqlite (ren Go, CGO_ENABLED=0 fungerar i Dockerfile:n); HTML-parsning via golang.org/x/net/html (redan i go.sum som indirekt beroende); e-pub-parsningen görs mot zip/OPF/NCX direkt. go.mod: github.com/simon/repo-template – interna paket importeras med full modulprefix, inte 'learn/internal/...'.

Fällor som upptäcktes under porten: (1) Graph.AddEdge skrev till en nil-karta i adjacency -> panic, (2) Go's flag.FlagSet slutar tolka flaggor efter första positionella argumentet -> infördes parseFlags som separerar flaggor och positionella argument, (3) rubriktext i XHTML får inte hamna i brödtexten (samma fälla som i Python-versionen).

Testerna ligger i respektive paket (_test.go) plus cmd/learn/main_test.go som bygger binären och kör kommandona end-to-end. 8 paket gröna.
