---
title: rust: learning-repot lades in som submodule mot egen fork
type: note
tags: [rust, git, submodule, solution, projekt:simon]
---

För att kunna committa egna lösningar utan att sönderdela tredjepartsrepot skapades en fork: github.com/simonbrundin/learn-to-code-with-rust (fork av paskhaver/learn-to-code-with-rust). Submodulens URL i .gitmodules pekar nu på forken (https, så andra kan klona utan SSH-nyckel), 'upstream' ligger kvar som remote för framtida pulls. Användarens lösningar ligger på branch 'learning/solutions' (commit da4b177) i forken; gitlinken i parent-repot pekar dit. Viktigt: gitlinken får INTE peka på en commit som bara finns lokalt - annars kan en klon inte hämta submodulen.
