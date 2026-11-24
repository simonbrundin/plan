---
title: Beroendeväljare med tangenten w
type: note
tags: [feature, keyboard, dependencies, projekt:plan]
---

Ny tangent 'w' på målsidan och mållistan: det markerade målet (förälder, huvudmål eller undermål) kan göras beroende av ett valfritt annat mål i hela grafen. Why: beroenden kunde bara sättas från beroendesektionen på målsidan och då bara för huvudmålet - inte för barnmål och inte från mållistan. Implementation: ny komponent src/nuxt/app/components/GoalDependencyPicker.vue (sökning bland alla mål, upp/ner, Enter väljer, Esc avbryter); listan filtrerar bort målet självt, befintliga beroenden och mål som skulle skapa en cykel (klienten), och API:t (POST /api/v1/goals/dependencies) verifierar ändå cykler. Hjälpmodalen (HelpModal.vue) uppdaterad. OpenSpec-change: add-goal-dependency-picker.
