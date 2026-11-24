---
title: go-imap: UidStore istället för Store med UID-mängd
type: note
tags: [bug, go, imap, projekt:pappa]
---

2026-11-19: Bugfix i src/go/imap/client.go - MarkAsRead och MarkAsUnread använde UidSearch (som returnerar UIDs) och skickade sedan dessa som SEKVENSNUMMER till Store-kommandot. Det gjorde att flaggan sattes på fel mejl, eller att servern svarade "Invalid messageset" om sekvensnumret inte fanns.

Varför det hände: go-imap har två varianter av Store - Store(seqset) tolkar mängden som sekvensnummer och UidStore(seqset) som UIDs. Sökningen via UidSearch gav UIDs, men kommandot som skickades var Store. Samma fel fanns i MarkAsUnread.

Fix: söker inte längre - UidStore anropas direkt med UID:t i mängden. Det gör att rätt mejl alltid flaggas.

Följd: åtgärden "markera som läst" i mailagenten fungerar nu, vilket var en förutsättning för att Acceptera-knappen i granskningsvyn skulle fungera överhuvudtaget.

Generell regel: i go-imap v1 ska UID-baserade kommandon (UidStore, UidFetch, UidMove, UidCopy) alltid användas när mängden kommer från ett UID, aldrig Store/Fetch/Move.
