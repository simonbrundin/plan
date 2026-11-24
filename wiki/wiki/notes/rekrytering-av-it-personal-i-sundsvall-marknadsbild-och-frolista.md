---
title: Rekrytering av IT-personal i Sundsvall: marknadsbild och frölista
type: note
tags: [rekrytering, sundsvall, it, sourcing, gdpr, projekt:pappa]
created: 2026-10-27
updated: 2026-11-03
status: active
related: [notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret, notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen, notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken, notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer]
---

Projekt 'pappa' (tomt repo) - automatisering av IT-rekrytering for ett 50-personersforetag i Sundsvall.

MARKNADSBILD (uppdaterad 2026-10-27 med SCB Yrkesregister, se noten "IT-arbetsplatser i Sundsvall: 3 469 enligt SCB Yrkesregistret"): Sundsvall har 3 469 anstallda i IT-yrken, Vasternorrlands lan 4 674, riket 248 417. Till detta kommer storsta enskilda poolerna: Aderian GDM 105 (Sundsvall), CoreIT 103 (Ornskoldsvik), Knowit Norrland 80, Sogeti Sundsvall 70, Visionite Sundsvall 50+, XLENT Sundsvall 48. Den tidigare handgjorda uppskattningen "~1 000-1 500 IT-personer i regionen" var FEL - den missade intern IT i stat, industri och finans.

INSIKT: marknaden ar for liten for kallkontakt/LinkedIn-scraping (LinkedIn ToS + GDPR risk). Nattverket i Sundsvall ar tatt - alla kanner alla. Ratt approach: bemanningsforetag (har redan CV-pooler, konsensbaserade), Bron Innovation + Good Tech Conference + Dragon Tech Talk/Azure UG for synlighet, SCB:s avgiftsfria fore tagsregister-API for bolagslista, och en normaliserad masterdatabas i Teams-bevakningslistan med dubblettkoll pa linkedin_urn. Kap 10-15 riktade kontakter/vecka for att inte brenna marknaden.

GDPR: legitimt intresse (Art 6.1.f) krver Art 14-information, dokumenterad intresseavvaggning, retention (6 man) och avsaknad av kontaktlista. Samtyckesbaserade CV-kallor (bemanningsforetag, AF ' Sok kandidater', egna ansokningar) undviker hela Art 14-problemet.

ARTEFAKT: /home/simon/repos/pappa/data/sundsvall-it-people.yaml - 88 namngivna personer med titel, arbetsgivare, normaliserade kompetenstaggar, publicerade kontaktuppgifter, kalla-URL, ort- och datakonfidens. Plus 36 arbetsgivare utan namngiven ledning att bearbeta manuellt.

METOD: endast offentliga kallor (orgens egna webbsidor, konferensprogram, pressmeddelanden, register). INGEN scraping av LinkedIn eller Facebook - Facebook har ingen API-vag for persondata overhuvudtaget (friends-endpointen tagits bort 2019).

## Relaterade

- [[notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer]]
- [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]]
- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]
- [[notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret]]
