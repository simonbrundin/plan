---
title: infrastructure: AGENTS.md uppdaterad + tre fällor för agenter
type: note
summary: Genomgång av AGENTS.md i ~/repos/infrastructure (2026-10-23) hittade tre konkreta fällor som kommer att kosta tid om de inte står i AGENT...
tags: [docs, agents, gitops, sops, projekt:simon]
created: 2026-10-23
updated: 2026-10-23
status: active
related: [concepts/gitops, concepts/sops]
---

Genomgång av AGENTS.md i ~/repos/infrastructure (2026-10-23) hittade tre konkreta fällor som kommer att kosta tid om de inte står i AGENTS.md:

1. `kustomize build --validate` finns INTE i kustomize v5 - kommandot failar. Rätt sätt att validera är `kustomize build <dir> > /dev/null`.
2. `*.enc.yaml` ligger i .gitignore, men de 22 redan trackade secretsen är trackade. En NY krypterad secret måste därför `git add -f`: annars tyst hamnar den aldrig i klustret.
3. `talos/generated/` är helt gitignored och `pxe/matchbox/assets/talos-config/` innehåller kluster-PKI (CA-nycklar, kubelet-tokens) - båda är byggartefakter, inte source of truth.

Dessutom: repo använder FluxOperator (inte vanilla Flux) och apps-katalogerna är pekare till externa repos. Komponenterna har tre olika source/release-varianter (source.yaml, controller-source.yaml, ocirepository.yaml) - normalisera dem inte utan beordring.

AGENTS.md och README.md är uppdaterade: katalogstrukturen speglar nu verkligheten (authentik borta, auth/ + openclaw/ + agentgateway/ + freellmapi/ tillagda, openspec borttaget), PXE-delsystemet har en egen sektion, och README:s OIDC-avsnitt pekar nu på `talos/patches/controlplane.yaml` (KubeAuthenticationConfig, Talos v1.14) och `scripts/rollback-oidc.sh` istället för två filer som inte existerade.

## Relaterade

- [[concepts/gitops]]
- [[concepts/sops]]
