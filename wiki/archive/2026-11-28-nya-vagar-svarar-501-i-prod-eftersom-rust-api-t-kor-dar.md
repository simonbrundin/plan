---
title: Nya vagar svarar 501 i prod eftersom Rust-API:t kor dar
type: note
tags: [drift, rust, verifiering, fasmaskin, projekt:plan]
---

VAD: fasmaskinen och verifieringen finns i Go-API:t (PR 179), men de nya vagarna svarar 501 i Rust-API:t - och det ar Rust-imagen som kor i produktion. Granssnittet fangar 501 och sager att API:t inte stodjer funktionen annu, sa att knapparna inte gor tyst ingenting.

VARFOR: det har funnits två API-implementationer i samma repo, och dev och prod har pekat pa olika. Utan hantering av 501 hade "jag ar klar", "Godkann" och "Avsla" bara varit knappar utan funktion i prod - och det syns inte i dev, dar Go kör.

SÅ SER DET UT: Rust-route-tabellen listar de nya vagarna och svarar 501 enligt tabellens konvention - Go ar referensimplementationen tills portningen gor. Samma gap galler agenten: agenten i prod ar Rust-imagen, sa agentens fasanrop (POST /goals/:id/phase) gar inte den nya vagen i prod alls. Go-agenten i dev gor det.

GALLER FRAMOT: antingen portas fasmaskinen och verifieringen till Rust-API:t, eller sa gaar produktionen tillbaka till Go-imagen. Tills dess maste fasmaskinen och verifieringen ses som dev-funktioner. Ett annat tecken pa samma driftproblem: environments/prod/patches/api-deployment.yaml kan bytas ut av en merge och avinstallera den andra implementationen.
