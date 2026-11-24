---
title: sales@soulmate.se: appen kan inte logga in (OAuth-clienten saknas i soulmate-tenanten)
type: note
tags: [mailagent, pappa, microsoft365, oauth2, bug, session, projekt:simon]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Undersökning 2026-11-16 (pappa/mailagent): kan applikationen logga in mot sales@soulmate.se?

Svar: NEJ, inte just nu. Flera oberoende hinder,Verifierade:

1. sales@soulmate.se är ett Microsoft 365-konto i tenanten "Soulmate Sundsvall AB"
   (50e951bf-5241-4248-befe-767d53b02c33) - userrealm svarar Managed. Adressen
   dyker upp som avsändare ("From: Soulmate Sales <sales@soulmate.se>") av mail
   som landar i michael.brundin@soulmate.se och sedan levereras vidare till
   one.com INBOX.Simon (michael@brundins.se).

2. Basic auth är AVSTÄNGT mot outlook.office365.com ("a2 NO Basic authentication
   is disabled."). Det gäller alla soulmate.se-adresser, inkl. sales@.
   => Losenordsinloggning mot Outlook är en död väg, oavsett lösenord.

3. one.com avvisar sales@soulmate.se som användarnamn (AUTHENTICATIONFAILED) -
   one.com kräver primärt användarnamn (michael@brundins.se). Därför når man
   sales@-brevet bara via Michael's brevlåda, inte direkt.

4. Appen "MailMaster" (client id f377fda6-1879-413f-a89a-8a33e2f5f1c6) finns
   INTE i soulmate-tenanten längre:
   - devicecode mot tenanten -> AADSTS700016 "Application with identifier ... was
     not found in the directory 50e951bf-..."
   - devicecode mot /common och /organizations -> AADSTS50059 "No
     tenant-identifying information found" => appen är single-tenant-registrerad
     i en annan tenant och kan aldrig användas mot soulmate.
   Alltså: device-code-flödet i appen kan inte ens starta mot soulmate.
   Publika client ids fungerar däremot: Thunderbird 9e5f94bc-... och Outlook
   d3590ed6-... returnerar user_code mot soulmate-tenanten.

5. Lösenordet i 1Password (Private/Microsoftonline, michael.brundin@soulmate.se)
   ger AADSTS50126 vid ROPC/token -> gäller inte som Entra-lösenord (och basic
   auth är ändå avstängt).

6. Koden stöder redan delade postlådor: config.go har fältet shared_mailbox och
   imap/client.go använder SharedMailbox som XOAUTH2-identitet medan token
   tillhör Username. Så korrekt mål för sales@ är INTE att logga in som
   sales@soulmate.se, utan att logga in som michael.brundin@soulmate.se (som
   måste ha Full Access på sales@-postlådan) och sätta
   shared_mailbox: sales@soulmate.se.

Vägar framåt:
  a) Registrera en egen multi-tenant-app (eller byt client_id till en publik
     client) och få admin consent för IMAP.AccessAsUser.All + SMTP.Send i
     soulmate-tenanten.
  b) Be admin i soulmate-tenanten ge michael.brundin@soulmate.se Full Access
     på sales@soulmate.se.
  c) Sätt shared_mailbox: sales@soulmate.se i config.yaml för Soulmate-kontot.

Miljöfall: IPv6 mot Microsoft fungerar inte på maskinen (ingen global IPv6-route,
"Network is unreachable"). Anslutningar måste gå över IPv4 - testskripten fick
timeouts tills de pinade IPv4.
