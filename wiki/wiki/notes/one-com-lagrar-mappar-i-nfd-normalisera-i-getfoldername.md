---
title: one.com lagrar mappar i NFD - normalisera i getFolderName
type: note
tags: [mail, imap, one.com, unicode, go, projekt:pappa]
created: 2026-11-16
updated: 2026-11-16
status: active
---

one.com (och IMAP i allmant) lagrar mappnamn i NFD-form. Att selecta 'INBOX.Kameo.Michael.Aterbet.Uppsala 4634' med NFC (A + ring) misslyckades, men samma mapp med NFD fungerade.

Fixen: getFolderName i src/go/imap/client.go returnerar nu norm.NFD.String(folder) - alltsa normaliseras centralt vid SELECT. ListFolders returnerar daremot läsbara mappar i NFC, sa UI:t visar dem som anvandaren skriver dem.

Varfor centralt och inte i anropskod: varje anropsplats som skickar ett mappnamn maste da veta om normen, och en glömd plats ger tyst fel. Tester i src/go/imap/folder_test.go: TestFolderNamesAreNFD och TestFolderNamesArePrefixedForOneCom (one.com anvander INBOX. som prefix, inte bara INBOX).

Relaterat: Go-klienten har nu en move-endpoint (POST .../emails/{uid}/move) sa ett arkiverat mejl kan flyttas tillbaka. Det behovdes for att flytta tillbaka ett mejl som arkiverades av misstag under test.
