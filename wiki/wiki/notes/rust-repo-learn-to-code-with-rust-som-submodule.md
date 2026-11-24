---
title: rust-repo: learn-to-code-with-rust som submodule
type: note
tags: [rust, git, submodule, projekt:simon]
created: 2026-11-03
updated: 2026-11-17
status: active
related: [notes/rust-utvecklingsmiljo-pa-workstationen]
---

I /home/simon/repos/rust (simonbrundin/rust) lades den klonade repot 'learn-to-code-with-rust' in som en riktig git-submodule (.gitmodules -> https://github.com/paskhaver/learn-to-code-with-rust.git) i stället för vanliga filer. Orsak: katalogen är ett klonat tredjepartsrepo med egen .git; att checka in filerna rakt av skulle ha sönderlagt historiken. Root .gitignore har 'target/' så att cargo-artefakter under src/rust/target/ inte följer med. Submodulen har lokala ocommittade ändringar i fyra learning-filer (data-types, variables-and-mutability) - de kan inte pushas till paskhaver/repot och ligger därför bara lokalt.

## Relaterade

- [[notes/rust-utvecklingsmiljo-pa-workstationen]]

