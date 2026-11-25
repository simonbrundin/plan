---
title: PR-preview-floden var trasig i fem steg - nu fungerar den
type: note
tags: [ci, bug, pr-preview, github-actions, projekt:plan]
---

2026-11-25: .github/workflows/pr-preview.yaml (jobbet 'Skapa PR Preview' + 'Stada upp PR Preview') var trasig i fem steg. Ingen anade det - varje PR fick en röd check 'Skapa PR Preview' och ingen kunde se varfor.

1. Taggen saknade $: '{{ github.sha }}' istallet for '${{ github.sha }}' -> docker/build-push-action fick ogiltig referens ('invalid reference format').
2. Ingen inloggning i ghcr.io: buildx hantade en anonym token och pushen failade med 403 Forbidden. Fix: docker/login-action med github.actor + GITHUB_TOKEN (samma monster som commit.agent.yaml).
3. pyyaml installerades aldrig: skripten pr-preview-resources.py och pr-preview-cleanup.py dog med ModuleNotFoundError.
4. build-push-action v5 har INGEN 'image'-output (bara imageid, digest, metadata), sa steps.publish.outputs.image var tomt och preview-manifestet fick image: ''. Fix: satt ihop tagg + digest.
5. Cleanup-jobbet checkade ut main med sparse-checkout pa enbart environments/pr, sa skriptet fanns inte i arbetsytan. Och main ar skyddad ('Changes must be made through a pull request'), sa pushen till main misslyckades med GH013. Fix: sparse-checkout Plus .github/scripts, och stadjobbet oppnar nu en PR via peter-evans/create-pull-request istallet for att pusha.

Loop: PR:ar med label skip-review skapar ingen preview-miljo (annars skapar stadjobbet en ny PR for varje PR). Verifierat: PR #123 (botens stadjobb-PR) hoppade over preview-jobbet och stadjobbet skapade ingen ny PR.

Kvar (designlucka, inte fixat): preview-overlays appliceras inte av Flux - Kustomization 'plan' i cluster1 pekar pa environments/prod, sa ingen preview-miljo deployas. Image och manifest byggs, men ingen Flux Kustomization tittar pa environments/pr/.
