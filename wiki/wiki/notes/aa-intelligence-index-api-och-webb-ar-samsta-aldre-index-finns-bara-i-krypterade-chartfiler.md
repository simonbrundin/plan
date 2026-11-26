---
title: AA Intelligence Index: API och webb ar samsta - aldre index finns bara i krypterade chartfiler
type: note
tags: [dashboard, aa, intelligence-index, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Gransking av 'modeller saknar intelligence-poang trots att AA visar dem': (1) AA free-API har index for 155 av 766 modeller, och dashboardens DB har samtliga 155 - syncen tappar inget. (2) AA:s EGEN modellsida visar 'Intelligence: N/A / Unknown' for modeller utan index (verifierat for claude-3-5-haiku: 'Unknown out of 4 units for Intelligence'); payloaden har intelligenceIndex:null aven for gemma-3-27b. (3) Mer index-data (aldre index-versioner for gamla modeller) ligger i AA:s /data/<hash>.txt-chartfiler - entropy 8.0, krypterade (AES/XOR/ChaCha-antal gissningar med manifest.key som 32-byte-nyckel gav inget lasbart) - inte lasbara utan AA:s klientkod. (4) API:t ronatar index till en decimal (webb 30.465 -> api 30.5) - uppsatt mojlig diskrepans. (5) Bare 18 av de 611 utan index ar effort-varianter dar bas-slugen har index. Slutsats: dashboardens 'arliga tomma' ar korrekt; AA:s webb visar N/A pa samma stallen i model summary. Om en specifik modell verkligen visar poang pa AA - undersok exakt den (t ex via jamforelselistans non-reasoning-varianter, som stammer overens med API utom rundning).
