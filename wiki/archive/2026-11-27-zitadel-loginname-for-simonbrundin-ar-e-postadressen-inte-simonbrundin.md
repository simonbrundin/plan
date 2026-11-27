---
title: Zitadel-loginname för simonbrundin är e-postadressen, inte 'simonbrundin'
type: note
tags: [zitadel, oidc, bug, lesson, projekt:infrastructure]
---

Symtom: loginvyn på auth.simonbrundin.com svarar 'User not found in the system' när loginname 'simonbrundin' skrivs.

Diagnos (verifierad 2026-11-27 via Zitadel SessionService):
- Användaren 346062090475757945 har userName = 'simonbrundin@gmail.com' (samma värde i e-postfältet). Det kom med i exporten från ZITADEL Cloud (eventet user.human.added).
- Org-policy 'must_be_domain' = false (projections.login_names3_policies), så det finns ingen suffixform 'simonbrundin@<orgdomän>'. Enda giltiga loginname är alltså hela e-postadressen (ev. telefonnummer om satt).
- Test via POST /zitadel.session.v2.SessionService/CreateSession: loginName 'simonbrundin' -> 404 'User could not be found (QUERY-Dfbg2)', loginName 'simonbrundin@gmail.com' -> session skapas. Loginvyn är alltså korrekt; felaktig loginname.
- disable_login_with_email = false, så e-postadressen fungerar både som username och som e-post-login.

Verifieringsteknik (användbar framöver): en JWT för login-client kan signeras med RSA-nyckeln i /login-service-key/tls.key i zitadel-login-podden (iss=login-client, aud=https://auth.simonbrundin.com) och ger API-åtkomst utan PAT.

Åtgärd om kort loginname önskas: byt användarens userName till 'simonbrundin' (via konsolen eller eventstore-eventet user.human.changed) eller slå på 'Add organization domain as suffix to loginnames' och gör simon-brundin-ab.auth.simonbrundin.com till primär orgdomän.
