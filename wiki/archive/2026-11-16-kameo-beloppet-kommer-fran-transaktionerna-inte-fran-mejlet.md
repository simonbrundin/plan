---
title: Kameo-beloppet kommer fran transaktionerna, inte fran mejlet
type: note
tags: [kameo, rante, pdf, projekt:pappa]
---

Rantebetalningsmejlet fran Kameo innehaller bara amne ('Rantebetalning - lon 5133'), lantagare och datum - INTE beloppet. Beloppet maste hamtas ur Kameos transaktionslista via GET /v1/account/transactions-by-loan?home_currency=SEK, och matchas genom att jemfora transaktionsdatumet mot mejldatumet (exakt datum forst, narbelagen inom tre dagar som fallback, alltid markerat i granskningsvyn).

Varfor detta matt: transaktionslistan ar paginerad. Ett anrop med per_page=50&page=1 missade lan 4634 och 4246 - de ligger pa senare sidor. Funktionen loopar nu upp till 10 sidor och avbryter pa tom sida eller fardre an 50 poster. Utan fixen fanns ingen transaktion och beloppet blev 0.

Klassificeringen (kortsiktig 8313 / langsiktig 8251) kommer fran lopets slutforsdatum, inte fran mejlet. Filnamnet blir 'Kameo - 5122 - Ranteutbetalning - 12,2 kr - 8313.pdf' med beloppet formaterat med komma.

PDF:en byggs server-side med pdfkit och sparas i kameo_interest_payment (bytea) - permanent lagring utan nytt Kubernetes-volume. Nyckeln (mail_account, mail_folder, mail_uid) gor att samma mejl inte kan bokforas tvat.
