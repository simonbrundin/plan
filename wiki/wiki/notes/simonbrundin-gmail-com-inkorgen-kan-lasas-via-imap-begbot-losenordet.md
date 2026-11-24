---
title: simonbrundin@gmail.com-inkorgen kan läsas via IMAP (begbot-lösenordet)
type: note
tags: [mail, gmail, imap, session, projekt:tmp]
created: 2026-11-13
updated: 2026-11-13
status: active
---

För att gå igenom inkorgen för simonbrundin@gmail.com används IMAP (imap.gmail.com:993) med app-lösenordet som ligger i ~/repos/begbot/.env (SMTP_PASSWORD) - det fungerar för både SMTP och IMAP. Lösenordet i ~/repos/neuro/.env (EMAIL_PASSWORD) är däremot ogiltigt (AUTHENTICATIONFAILED). Inloggningen sker via op-cli/1Password med posten 'Google' (simonbrundin@gmail.com) om man behöver huvudlösenordet.

Metod: hämta headers med UID SEARCH SINCE + UID FETCH (BODY.PEEK[HEADER.FIELDS ...]), filtrera bort nyhetsbrev/marknadsföring lokalt (Gmail X-GM-RAW-sökning fungerar inte via imaplib utan citattecken). Relevanta mejl flaggades med etiketten 'Att-handla' + \Flagged via STORE +X-GM-LABELS. Etiketten skapas via CREATE 'Att-handla' (ASCII-namn fungerar; Gmail mappar det till etiketten).

Viktigaste obesvarade punkter i genomgången (13 nov 2026): one.com-påminnelse om obetald hosting för brundins.se (1 491,25 kr, suspendering/inkasso risk), Aiai tidrapport oktober ej godkänd, Kivra-fakturor (REKO Sundsvall, Sundsvall Vatten, Sundsvall Energi) + brev från AMF och Transportstyrelsen, Svensk Padel medlemsregistrering senast 30 nov, UniFi WAN1 (Starlink) fladdrar upprepade gånger, WordPress Vulnerability Monitor om advisoryspain.com och tretjak.se, NeuroQuant-kontot låst + monitorn kunde inte hämta verifieringskoden.
