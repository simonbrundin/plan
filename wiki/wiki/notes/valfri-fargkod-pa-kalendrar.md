---
title: Valfri färgkod på kalendrar
type: note
tags: [kalender, färg, projekt:plan]
created: 2026-11-18
updated: 2026-11-18
status: active
related: [notes/flera-kalendrar-med-egna-farger-i-plan]
---

Användaren vill kunna sätta en helt valfri färg på en kalender - en
specifik hexkod, inte bara ur en palett. Nu finns ett fält där koden skrivs,
både i skapa-formuläret och i varje kalenderrad, bredvid en förhandsvisning
av färgen. Paletten finns kvar som genväg.

## Regler

- Sex siffror (#6366f1) eller tre (#abc). Tre siffror expanderas till sex,
  versaler sänks - det görs på servern i validateColor, så att alltid sex
  siffror lagras.
- En ogiltig kod avvisas: servern svarar "Invalid color" och klienten
  skriver tillbaka kalenderns nuvarande färg i fältet med ett felmeddelande.
- Klientens normalisering ligger i normalizeHexColor i
  src/nuxt/app/utils/calendarColors.ts och har samma regler som servern.

## Varför

Paletten är snabb men låser in färgvalet. En egen färg behövs när man
matchar en kalender mot ett varumärke, ett projekt eller en färg i en
skärmdump. Att fältet finns i varje rad gör det också möjligt att justera
en befintlig kalenders färg utan att öppna systemets färgväljare.

## Gotchas

- **Färgväljaren i raden och hexfältet måste hållas i synk.** Båda skriver
  samma kalender. Hexfältet har en egen draft (draftColors) precis som
  namnfältet har, så att fältet inte hoppar tillbaka medan man skriver.
- **Serverns gamla validateColor krävde exakt sju tecken** och avvisade
  därför "#abc". Den är nu delad mellan kalendrar och händelser, så en
  ändring i den påverkar båda.

## Relaterade

- [[notes/flera-kalendrar-med-egna-farger-i-plan]]

