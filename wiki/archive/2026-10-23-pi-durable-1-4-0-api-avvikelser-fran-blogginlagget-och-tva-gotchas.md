---
title: Pi Durable 1.4.0: API-avvikelser från blogginlägget och två gotchas
type: note
tags: [pi, durable, api, projekt:infrastructure]
---

Blogginlägget https://earendil.com/posts/pi-durable/ beskriver API som det såg ut vid release; paketet @earendil-works/pi-durable@1.4.0 skiljer sig på några punkter: (1) ett verktyg returnerar { output: [...] } där blogginlägget visar { content: [...] } - med content-fältet får man TypeError 'content.filter is not a function' i boundContent; (2) defineDoc kräver checkpointWhen(value, ops, info) => boolean; (3) tx.createTask() i ett harness.commit (sessionnivå) kräver options.conversationId - annars 'Tx.createTask() requires options.conversationId'; använd conversation.commit() som binder konversationen; (4) transkriptposter skrivs med tx.appendEntry(conversationId, { kind, data }). Gotcha som kostade tid: en task-fas som gör fil-I/O med fasens context.abortSignal hänger när invokationen avbryts (Node:s fs.writeFile med ett abortat signal hänger i stället för att kasta) - därför ska man vänta på tasken (harness.waitForTask) innan harness.close(), annars avbrytas fasen mitt i skrivningen.också en viktig skillnad: harness.close() avbryter pågående fas-invokationer.
