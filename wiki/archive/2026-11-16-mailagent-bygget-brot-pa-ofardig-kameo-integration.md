---
title: Mailagent-bygget bröt på ofärdig Kameo-integration
type: note
tags: [bug, mailagent, go, build, projekt:pappa]
---

Dockerbygget för mailagent bröt eftersom en pågående Kameo-integration hade lämnat src/go/kameo/client.go med oanvänd os/exec-import och en anonym result-struct som inte kunde returneras som *KameoLookupResult. Samtidigt var src/go/actions/executor.go delvis omskriven: context/os-importer saknades, gamla hjälpfunktioner (getActionArg, sendEmail m.fl.) hade fallit bort, och server.go refererade s.executor efter att fältet tagits bort. Fixen återställde de saknade executor-hjälparna, använde context.Background för Kameo-anropet, städade imports/returtypen och tog bort den felaktiga server-referensen. Verifierat: go build ./..., go test ./... och docker build --platform linux/amd64.
