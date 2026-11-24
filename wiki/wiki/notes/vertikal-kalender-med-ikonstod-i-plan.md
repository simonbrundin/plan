---
title: Vertikal kalender med ikonstöd i Plan
type: note
tags: [kalender, feature, ikon, quick-add, projekt:plan]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/flera-kalendrar-med-egna-farger-i-plan, notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades, notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning, notes/kalenderns-zoom-styrs-av-vertikal-scrollning, notes/nya-mal-far-ikon-fran-titeln, notes/plan-api-anvander-scalar-for-openapi, notes/repo-template-ny-rust-api-src-rust-med-axum-utoipa-scalar, notes/vertikal-kalender-med-kontinuerlig-zoom]
---

## Vad som byggdes

Vertikal kalendersida i Plan där tiden löper uppifrån och ned. Zoom sker
genom vertikal skrollning/dragning - nedåt komprimerar tiden (mer
framtid syns), uppåt expanderar den. Zoomnivåerna går från timmar till
praktiskt taget obegränsat (JS Date:s tekniska maxgräns).

## Viktiga beslut och varför

- **Ett gemensamt 'icon'-fält** i stället för separata 'emoji'- och
  'icon'-kolumner. En kalenderhändelse har antingen en emoji ("🏃") eller
  ett Iconify-namn ("heroicons:star"). De skiljs åt vid rendering med
  kolonlogiken: Iconify-namn har alltid kolon, emoji har aldrig det.
- **Tre-state PATCH-semantik för ikon**: nil = behåll nuvarande värde,
  tom sträng = töm ikonen, värde = sätt ikonen. Utan detta gick det
  inte att radera en ikon, eftersom PATCH normalt inte kan skilja
  'sätt inte' från 'sätt till null'.
- **Overrides och split-serier ärver förälderns ikon** via resolveIcon() i
  repository-lagret, men kan tömma eller ändra den explicit.
- **Emoji extraheras före all annan tolkning** i Quick Add, så att emoji
  inte stör datum-, tid- eller upprepningstolkning, och tas bort ur
  titeln. Endast tecken med emoji-presentation matchas - annars skulle
  '©' och '▶' felaktigt blivit ikoner.
- **Månadsgränser har egna ikoner** (❄ ♥ ✿ ☀ ...) och etiketten
  tonas ned så att linjer inte syns genom texten. Ordinarie
  datumetiketter som överlappar månadsgränsen döljs via
  MONTH_LABEL_COLLISION mätt i viewport-procent, vilket fungerar
  oberoende av zoomnivå.
- **Mobil first**: under 640 px döljs titel, beskrivning, skala-legend
  och tangentbordsinfo, och kalendern får 100dvh-baserad höjd.

## Buggar som hittades under verifieringen

- **Quick Add:s förhandsvisning visade aldrig emoji-ikonen**. Template
  refererade till 'icon' utan att den var definierad i script setup.
  Nu finns en computed 'icon' som läser parsed.value.icon.
- **Escape stängde inte Quick Add** när fokus låg på en knapp i overlayen
  (t.ex. lägesknappen 'Mål') - bara tangenttryck i textfältet hanterades.
  Nu hanteras Escape globalt i layouten.
- **'n' kunde öppna Quick Add när overlayen redan var öppen**. Nu
  kontrolleras isQuickAddOpen innan öppning.

## Verktyg som användes

- Go 1.24 + Gin + pgx för API:t, RFC 5545 RRULE via teambition/rrule-go.
- Schema hanteras deklarativt via Atlas (environments/base/db/schema.yaml).
- Dev kör mot Neon branch 'dev' via environments/dev/db/neon.env.
- Cucumber (feature/step-definitions) för parser- och ikonlogiken.
- Playwright + systemets chromium för DOM-granskning av den byggda appen.

## Gotchas att komma ihåg

- swag-generering: 'swag init -g cmd/server/main.go --parseInternal -o docs'
  (utan --parseDependency annars byter definitionerna namn till
  plan-api_internal_models.* och bryter Scalar UI-referenser).
- go.mod behöver 'go mod tidy' efter att swaggo/files och gin-swagger
  tagits bort, annars ligger gamla beroenden kvar i go.sum.
- Iconify renderas client-side via CSS-mask-image, inte som inline-SVG.
  Det är därför innerHTML ser tomt ut i DOM-inspektion - kontrollera
  computed style 'maskImage' istället.

## Relaterade

- [[notes/flera-kalendrar-med-egna-farger-i-plan]]
- [[notes/repo-template-ny-rust-api-src-rust-med-axum-utoipa-scalar]]
- [[notes/plan-api-anvander-scalar-for-openapi]]
- [[notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning]]
- [[notes/nya-mal-far-ikon-fran-titeln]]
- [[notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades]]
- [[notes/kalenderns-zoom-styrs-av-vertikal-scrollning]]
- [[notes/vertikal-kalender-med-kontinuerlig-zoom]]
