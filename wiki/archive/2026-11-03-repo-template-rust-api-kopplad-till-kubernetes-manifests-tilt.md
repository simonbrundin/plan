---
title: repo-template: Rust-API kopplad till Kubernetes (manifests + Tilt)
type: note
tags: [rust, kubernetes, tilt, kustomize, repo-template, feature, projekt:repo-template]
---

2026-11-03: Rust-API:t i src/rust kopplades in i Kubernetes, i samma mönster som Go-API:t.

Några filer:
- environments/base/rust-api-deployment.yaml (image rust-api:latest, imagePullPolicy IfNotPresent, httpGet-probe mot /health istället för tcpSocket - Rust-API:t har /health)
- environments/base/rust-api-service.yaml (rust-api-service, ClusterIP 8080)
- environments/base/kustomization.yaml: base + de två nya resurserna
- environments/dev/Tiltfile: docker_build("rust-api", "../../src/rust") + egen portallokering + k8s_resource med länkar till /health och /docs

Viktigt: base-overlayn refereras av dev, pr och prod, så deploymenten finns i alla miljöer utan extra arbete. Tilt bygger imagen och patchar image-referensen automatiskt (samma mechanism som go-api) - därför ska image-fältet i manifestet heta 'rust-api:latest' (utan registry-prefix) för att Tilt ska hitta matchningen.

Verifierat i den riktiga kind-klustret (context kind-dev): tilt ci byggde och deployade rust-api, poden gick Running 1/1, och /health, /docs/openapi.json samt /docs svarade korrekt via service. kustomize build går igenom för base, dev, pr och prod.

Kvarvarande lucka (repo-wide, inte Rust-specifikt): prod/pr har ingen image-override - alla images pekar pa :latest, och CI-workflows refererar fortfarande till en gammal frontend/-katalog som inte finns. Det gäller go-api lika. Det bor fixas i en egen uppgift.
