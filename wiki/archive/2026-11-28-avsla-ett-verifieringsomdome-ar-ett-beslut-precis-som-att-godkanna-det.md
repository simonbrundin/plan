---
title: Avsla ett verifieringsomdome ar ett beslut precis som att godkanna det
type: note
tags: [verifiering, human-gate, agent, gransnitt, projekt:plan]
---

VAD: ett verifieringsomdome kan nu avslas av en manska, inte bara godkannas. Ny endpoint POST /api/v1/verification-results/:id/reject satter omdomet till avslagen och loggar beslutet.

VARFOR: godkannande var den enda handlingen. Det gick att godkanna ett omdome men inte att saga att det inte haller - en manska som tyckte omdomet var fel fick loppa over det eller gora om verifieringen. Ett avslag ar ett beslut, precis som ett godkannande, och det ska gora att synas.

SÅ SER DET UT: avslaget ogiltigforklarar de godkannade omdomena som fanns innan (raderna tas bort, det avslagna omdomet behalls for sparning). Annars skulle malet raknas som verifierat trots att manskan just sagt att omdomet inte haller - HasPassedVerification soka bland annat passed-omdomen med manska. Uppdatumet sker i en transaktion, sa att malet aldrig kan hamna i ett halvt lamge.

MALLET: agent-scope blockerar bade approve och reject for agentens nyckel - det ar human gate i API:t, inte i en prompt. Agenten far lamma bevis och kora verifieringen, men aldrig besluta om omdomet.

GALLER FRAMOT: granssnittet har knapparna Godkann och Avsla i bade mal-spanelen och agent-docken (fliken Korning). Malet visar "avslutades som" - deklarerat eller verifierat - pa malet och i listan av avslutade mal, sa att skillnaden mellan "jag tror det ar klart" och "det har ar bevisat" gar att se direkt.
