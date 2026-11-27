---
title: kubectl hanger nar current-context ar oidc-cluster1 och OIDC-token gatt ut
type: note
tags: [kubeconfig, oidc, felsokning, kubectl, projekt:infrastructure]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [entities/talos-linux]
---

kubectl mot cluster1 hangde meddelandet "Command timed out" trots att API-servern svarade pa /healthz med ok. Orsaken var inte klustret utan kubeconfigen: current-context var oidc-cluster1, vilket använder exec-pluginen `kubectl oidc-login get-token --oidc-issuer-url=https://auth.simonbrundin.com`. Nar den cachade OIDC-token går ut kan pluginen inte förnya den utan interaktiv inloggning, och kubectl väntar då i stället för att ge ett fel.

Workaround: använd kontexten admin@cluster1 (klientcertifikat, ingen exec-plugin), eller peka om current-context till den. Att felsöka klusterproblem med en hängande kubectl är en tidstjuven - kontrollera /healthz med curl och testa kubectl mot en kontext utan exec-plugin innan man drar slutsatser om API-servern.

Notera att detta inte handlar om op eller 1Password-login; OIDC-inloggningen mot auth.simonbrundin.com är en separat sak.

## Relaterade

- [[entities/talos-linux]]

