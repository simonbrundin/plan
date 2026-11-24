---
title: Mailagenten flyttad från mailmaster-repot in i pappa
type: note
tags: [mail, flytt, kluster, nuxt, go, projekt:pappa]
---

Mailmaster-repot (~/repos/mailmaster) är nu arkiv. Allt väsentligt ligger i
~/repos/pappa:

- Go-klienten -> pappa/src/go (modul "mailagent", oförändrad). Image byggs ur
  pappa/src/go/Dockerfile.
- Webbappen -> pappa/src/nuxt/app/pages/mail (+ components/mail, composables
  useMail*, utils/mail). Sidorna ligger under /mail.
- Klienten proxas via pappa/src/nuxt/server/api/mail/[...path].ts -> alltså
  aldrig direkt från webbläsaren. Det löser CORS och gör SSR i klustret
  möjlig (NUXT_MAIL_API_BASE=http://mailagent:8401/api/v1).
- Deployment: pappa/environments/base/mailagent-{deployment,service}.yaml,
  byggd av Tilt (docker_build("mailagent", "../../src/go")).
- Hemligheter: pappa/environments/dev/mail.env (gitignorad) -> Secret
  pappa-mail-secrets via pappa/scripts/mail-secrets.sh (npm run mail:secrets).

Varför proxy istället för att webbappen ska anropa mailagenten direkt: pappas
Nuxt-app körs i klustret, så localhost:8401 finns inte där. Proxyn gör att
webbläsaren och SSR går samma väg och att klientens adress bara behöver
sättas en gång (på servern).

Gäller framöver:
- Ändra inte i mailmaster-repot - det är arkiv. Historiken följde inte med
  (valet var att skita i githistoriken).
- Kameo-uppgifterna och mailhemligheterna är nu i pappa/environments/dev/mail.env
  och flyttades dit från mailmaster/environments/dev/.env + kameo.env.
- Regelendpointet (PUT/GET /api/v1/rules) är en stub - sparar inte. Är kvar
  från mailmaster, inte en regression.
- Testerna för maildelen kör med bun (bun test server/tests) medan resten av
  repot är npm-baserat - blandade runners, okej så länge det bara är maildelen.

Verifierat i dev-klustret (namespace repo-template): mailagent-poden kör
(Ready 1/1), /api/mail proxas via nuxt-poden tillbaka till mailagenten,
SSR av /mail renderar, och Kameo-uppsökningen svarar 200 med source=env.
Fällan som kostade tid: image med bara CMD + deploymentens `args` ger
"executable file not found" - Kubernetes ersätter CMD med args om ingen
ENTRYPOINT finns. Nu har imagen ENTRYPOINT ["/app/mailagent"].
