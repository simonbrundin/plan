---
title: IMAP-inloggning för michael.brundin@soulmate.se kräver admin consent
type: note
tags: [mailmaster, imap, oauth2, microsoft365, bug, projekt:mailmaster]
---

Undersökning 2026-10-29: Kan michael.brundin@soulmate.se loggas in via IMAP?

Fynd:
- soulmate.se MX = soulmate-se.mail.protection.outlook.com (Microsoft 365 / Exchange Online, tenant 50e951bf-5241-4248-befe-767d53b02c33). basic auth är AVSTÄNGT på outlook.office365.com och imap-mail.outlook.com ("Basic authentication is disabled").
- IMAP med OAuth2/XOAUTH2 fungerar tekniskt, men appen "MailMaster" (client id f377dda6-1879-413f-a89a-8a33e2f5f1c6) kräver admin consent i soulmate-tenanten. Device-code-flödet (scope https://outlook.office.com/IMAP.AccessAsUser.All + SMTP.Send) ger "Need admin approval" eftersom Microsoft 2025 lade IMAP.AccessAsUser.All (Office 365 Exchange Online) i listan över riskiga delegated-permissions där user consent är avstängt.
- one.com-kontot michael@brundins.se (imap.one.com:993, lösenord MAIL_PASSWORD_MICHAEL) fungerar och är det konto som används i src/go/config/config.yaml. Aliaset michael.brundin@soulmate.se levererar dit (To-header visar båda adresserna; 13 mail i INBOX nämner aliaset), så det mesta soulmate-posten är nåbar via one.com-kontot.
- imap.one.com avvisar LOGIN med användarnamnet michael.brundin@soulmate.se (AUTHENTICATIONFAILED) – one.com kräver primärt användarnamn.

Varför: Detta avgör om repot ska konfigureras mot Microsoft 365 (OAuth2, kräver admin consent) eller mot one.com (lösenord, fungerar nu).

Gäller framöver: Använd michael@brundins.se + imap.one.com för IMAP tills en admin i soulmate-tenanten beviljat admin consent för MailMaster-appen. Admin-consent-url: https://login.microsoftonline.com/50e951bf-5241-4248-befe-767d53b02c33/v2.0/adminconsent?client_id=f377dda6-1879-413f-a89a-8a33e2f5f1c6&redirect_uri=http://localhost:8400/callback&scope=openid+profile+offline_access+https://outlook.office.com/IMAP.AccessAsUser.All+https://outlook.office.com/SMTP.Send
