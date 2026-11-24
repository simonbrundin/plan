---
title: Cachade IMAP-sessioner kräver hantering av döda anslutningar
type: note
tags: [mailmaster, bug, imap, go, projekt:simon]
created: 2026-10-29
updated: 2026-10-29
status: active
related: [notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code]
---

När man cachar IMAP-sessioner per konto (i stället för att stänga och
öppna ny varje gång) uppstår ett problem: mailserver stänger idle-anslutningar.
one.com svarar då "write tcp ...: write: broken pipe" och den döda sessionen
återanvänds om och om igen.

Lösningen som används i `src/go/mcp/session.go`:

1. `accountSession.broken` - flagga markeras när ett anrop ger ett
   nätverksfel.
2. `isConnectionError(err)` - identifierar broken pipe, connection reset,
   closed network connection, i/o timeout och EOF. Serverfel som "no such
   mailbox" räknas INTE som nätverksfel.
3. `ensureConnectionFor` - kastar bort en trasig session och återansluter.
4. `Server.listEmails` - gör EN retry: återansluter och kör listningen igen.

Utan detta fick UI:t "Failed to fetch" trots att servern varit uppe.

Relaterat: `imap.Client.Close()` i go-imap stänger *mboxen* (IMAP CLOSE), inte
TCP-anslutningen. Rätt sätt att stänga är `Terminate()`. Detta hade maskedats
tidigare eftersom sessioner aldrig stängdes.

## Relaterade

- [[notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code]]

