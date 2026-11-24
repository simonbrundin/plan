---
title: workmux merge misslyckas i huvud-worktree - kora kommandot i feature-worktreet
type: note
tags: [workmux, git, worktree, bug, projekt:agents]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Felet 'Cannot merge branch main because it is checked out in the main worktree' uppstar när workmux merge körs i repoets huvud-worktree. Workmux skapar då en temporär worktree för main-branchen, vilket git förbjuder när main redan är utcheckad i huvud-worktreet.

Verifierat 2026-11-20 med workmux 0.1.283 i ren testrepo:
- 'workmux merge' i huvud-worktree (main utcheckad) -> felet ovan.
- 'workmux merge' inuti feature-worktreet -> fungerar, mergar och städar.
- 'workmux merge <branch>' i huvud-worktree -> fungerar OM branchen har en workmux-skapad worktree. Utan worktree: 'Worktree X not found'.

Konsekvens för infrastructure: repoet har bara main-worktreet och två omergade lokala branchar (oidc, feat/kvm-switch-control). workmux merge oidc skulle faila med 'Worktree oidc not found' - man måste först 'workmux add oidc' skapa worktreet, eller merga manuellt med git.
