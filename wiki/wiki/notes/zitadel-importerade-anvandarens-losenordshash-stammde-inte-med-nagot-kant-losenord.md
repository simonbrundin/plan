---
title: "ZITADEL: den importerade användarens lösenordshash stämde inte med något känt lösenord"
type: note
tags: [zitadel, bug, lesson, projekt:simon]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/zitadel]
---

Efter att ZITADEL-installationen återställts från en export av ZITADEL Cloud
(klient-ID:n återställda genom eventpatchar) gick det inte att logga in som
användaren `simonbrundin@gmail.com`. Login-sidan laddade korrekt och
lösenordssteget nådde fram till ZITADEL, men varje försök föll med:

```
Password is invalid (COMMAND-3M0fs)
  err.parent="passwap: password does not match hash: "
```

Användarens state var `1` = `UserStateActive` (inte `6` = INITIAL), så det var
inte init-code-flödet som blockerade. Problemet var själva hashen: den
importerade `user.human.password.changed`-eventet innehöll en bcrypt-hash
(cost 12) som inte matchade något lösenord som fanns tillgängligt - varken
1Password-objektet "Zitadel" (Cloud-kontot) eller "Zitadel - Self Hosted".

## Varför det hände

En hash går inte att "återskapa" - den berättar bara vilket lösenord som
ska ge den. Exporten med `with_passwords: true` ger hashen, men om
lösenordet inte finns någonstans i klartext finns det bara en väg fram:
sätta ett nytt lösenord.

## Lösningen

1. Generera ett nytt lösenord och bcrypt-hash (cost 12, prefix `2a`).
2. Lägg in ett nytt `user.human.password.changed`-event i
   `eventstore.events2` med payload `{"encodedHash": "<hash>",
   "changeRequired": false}`. Övriga fält kopieras från det befintliga
   eventet: `instance_id`, `aggregate_type=user`, `creator`, `owner`,
   `revision=1`, `in_tx_order=0` och ett `position` strax större än
   `max(position)`.
3. Projektionerna plockar upp eventet inom någon sekund.

Resultat: inloggningen fungerar, `is_email_verified` blev true och
Plan-appens hela OIDC-flöde (inloggning, callback, session, `plan_session`
och `plan_user`-cookies, 200 från `/api/proxy/goals`) kunde verifieras
end-to-end med Playwright.

## Lärdomar

- Verifiera en hash mot kandidatlösenord (bcrypt) innan man drar slutsatser
  om vilket lösenord som gäller - ett 1Password-objekt kan vara inaktuellt.
- Efter en import av en användare utan fungerande inloggning: sätt lösenordet
  aktivt och dokumentera var det ligger (här: SOPS-krypterat i
  `zitadel-user-credentials` i infrastructure-repot, plus 1Password).
- `eventstore.events2` är källan för lösenordshashar; `auth.users3` bara
  anger om ett lösenord är satt.

## Relaterade

- [[concepts/zitadel]]
- [[notes/zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in]]
- [[notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om]]
