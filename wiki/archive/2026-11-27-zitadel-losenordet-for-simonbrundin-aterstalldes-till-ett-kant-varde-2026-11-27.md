---
title: Zitadel-lösenordet för simonbrundin återställdes till ett känt värde 2026-11-27
type: note
tags: [zitadel, oidc, bug, lesson, projekt:infrastructure]
---

Symtom: loginname 'simonbrundin' hittades inte ('User not found in the system') och efter att username bytts till 'simonbrundin' ville det gamla lösenordet inte fungera ('password.check.failed' x10 i eventstore).

Varför det gamla lösenordet inte fungerade: hashen som importerades från ZITADEL:s export (user.human.password.changed, $2a$12$QzG6...) är ZITADEL Clouds hash för det lösenord som gällde där. Det matchar inte det lösenord användaren faktiskt skriver in - kontrollerat med bcrypt mot hashen (ZITADEL Cloud-miljöns lösenord gav ingen träff). Det kan inte avgöras vilket lösenord hashen motsvarar, eftersom hashen inte går att återställa till klartext.

Åtgärd 2026-11-27: lösenordet sattes om till ett känt slumpvärde (aitHPRqbE#Y48C9S!2AS35Z_) via Zitadels API, och värdet sparades krypterat i repot: infrastructure-flux/components/zitadel/chart/user-credentials.enc.yaml (SOPS+age, måste adderas med 'git add -f' eftersom *.enc.yaml matchas av .gitignore). Secreten appliceras i klustret som zitadel-user-credentials i namespace zitadel.

Samtidigt byttes username från 'simonbrundin@gmail.com' till 'simonbrundin' (zitadel.user.v2.UserService/UpdateHumanUser) och e-posten markerades verifierad, så både 'simonbrundin' och 'simonbrundin@gmail.com' fungerar som loginname.

Verifiering: riktig inloggning i headless Chromium mot auth.simonbrundin.com/ui/v2/login - loginname 'simonbrundin' och 'simonbrundin@gmail.com' leder till '/ui/v2/login/signedin' ('Welcome ZITADEL Admin! You are signed in.'), medan det gamla lösenordet stannar på lösenordssidan. password.check.succeeded-eventen ökade med 2 per lyckad testinloggning.

Viktigt: SMTP är inte konfigurerat (projections.smtp_configs6_smtp är tom), så 'Glömt lösenord' i loginvyn går inte - ingen e-post med återställningskod skickas. Lösenordsbyte måste göras via API:t/konsolen tills SMTP sätts upp.

Kommandot som gav åtkomst till API:t (ingen PAT behövdes): en JWT för login-client signeras med RSA-nyckeln i /login-service-key/tls.key i zitadel-login-podden (iss=login-client, aud=https://auth.simonbrundin.com) och skickas som 'Authorization: Bearer'. Användarnamnet i Kubernetes är 'zitadel-' + sub (oförändrat av username-bytet).
