---
title: pr-preview.yaml bygger ogiltig image-tagg - alla PR-preview-floden faller
type: note
tags: [ci, bug, pr-preview, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: superseded
---

2026-11-25: .github/workflows/pr-preview.yaml rad 49 har taggen

  ${{ env.REGISTRY }}/plan/pr-preview:${{ github.event.pull_request.number }}-{{ github.sha }}

Saknar '$' fore '{{ github.sha }}', sa docker/build-push-action far en ogiltig tag och jobbet 'Skapa PR Preview' faller med 'invalid reference format'. Det hander for varje PR - det ar inte kopplat till nagon diff.

Varfor det inte fixades i PR #119: fixen ar en teckenfix, men nar den blir grann taggar byggningen lyckas och workflowen skapar environments/pr/<pr-nr>-manifest och committar dem till PR-grenen, vilket deployar en preview-miljo (namespace plan-pr-<nr>) i produktionsklustret. Det ar en sidoeffekt som man inte vill gora omedelbart i ett docs-PR.

Fixen nar den gors: ${{ github.sha }}. Losningen pa sikt: lat preview-floden bara gora det den ska gora, och var samma ar bakom att 'main' inte skyddas av en kravd check for det jobbet.

---

**Superseded 2026-11-25.** Taggen ar fixad (\`${{ github.sha }}\`), men det visade sig vara fyra fel till i samma workflow - inloggning i ghcr.io, saknad pyyaml, tomt \`image\`-output och ett stadjobb som inte kunde skicka nagonting till den skyddade main-grenen. Alla fem ar fixade och mergeade; se [[notes/pr-preview-floden-var-trasig-i-fem-steg-nu-fungerar-den]].
