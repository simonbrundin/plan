---
title: Learn: designsystem A3 Folio i Nuxt
type: note
tags: [nuxt, design, rörelse, ui, projekt:simon]
---

Learn-SPA:t har fått en validerad designriktning 'A3 Folio' (väljd av användaren 2026-11-03 efter tre förslag: A1 varm studio, A2 nattbibliotek, A3 folio).

Varför: den ursprungliga SPA:n var funktionell men platt - blå standardpalette, ingen typografisk hierarki, ingen rörelse. Användaren ville ha ett gränssnitt som känns som ett rum, inte en instrumentpanel.

Vad: djup olivgrön mark (#121513), terrakotta (#c97f5a) som enda accent, Georgia som display-typ, mono-etiketter, hårlinjer, nästan hårda hörn (4px). Rörelsen är CSS-native utan nya beroenden: inträde 300ms cubic-bezier(0.2,0,0,1) med 40ms stagger, hover 120ms, grafens kanter ritar sig 520ms, starkaste noden andas med 3s puls + halo. prefers-reduced-motion stänger av allt.

Viktig lärdom: förhandsvisningen måste loopa rörelsen. Första förslaget spelade animationerna en gång vid inläsning, och användaren svarade 'jag ser ingen rörelse' - trots att animationerna faktiskt körde (mätt i Chromium). En rörelse som bara sker en gång vid sidladdning är osynlig för den som landar på sidan efter att den spelat klart. Halon var dessutom så subtil (skala 1->1.35 pa 13px) att den knappt märktes. Fixen var en demo-sektion som loopar, tydligare halo och en scroll-triggerad uppspelning.

Verktyg: genjutsu _jutsu/tells (katalogen over model-reflexer) hjalp till att resonera kring vilka val som var explicita i tesen och vilka som var modellens default. Audit-skriptet (design-audit/scripts/audit.py) flaggar 'hover-no-transition' som falsk positiv - transition ligger pa basregeln, inte i hover-regeln.
