---
title: Globala agentinstruktioner deployas via dotfiles
type: note
summary: Globala agentinstruktioner bor i dotfiles-repot och deployas med stow till agenternas konfigurationskataloger.
tags: [konventioner, agent, wiki, projekt:typetrainer]
created: 2026-10-21
updated: 2026-10-21
status: active
---

Den globala AGENTS.md ligger i dotfiles-repot under pi/.pi/agent/AGENTS.md. Stow deployar pi-paketet till ~/.pi/agent/, vilket gör filen till pi:s agent-dir-kontextfil och den laddas i varje session. ~/AGENTS.md är en symlänk till samma fil, så andra agentverktyg (Codex, OpenCode) som läser AGENTS.md från cwd och föräldrar också får reglerna. Stow-körs av bootstrap.sh: for dir in */; do stow "$dir" --target="$HOME"; done.

## Relaterade
