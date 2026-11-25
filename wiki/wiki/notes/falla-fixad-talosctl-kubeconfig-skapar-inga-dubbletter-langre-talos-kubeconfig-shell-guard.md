---
title: Fälla fixad: talosctl kubeconfig skapar inga dubbletter längre (talos-kubeconfig + shell-guard)
type: note
tags: [kubeconfig, cluster1, talosctl, fälla, projekt:infrastructure]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna, notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat]
---

2026-11-25, efter kubeconfig-städningen (se [[notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna]]): fällan att `talosctl kubeconfig` mergar in i ~/.kube/config och skapar cluster1-N-dubbletter är nu blockerad på tre nivåer.

1) SKRIPT ~/.local/bin/talos-kubeconfig
   - hämtar med `talosctl kubeconfig <tmp> --merge=false`
   - väljer automatiskt en controlplane som svarar (testar controlplane-IP:n ur talos/nodes.yaml, en i taget - talosctl kubeconfig kräver exakt EN nod)
   - installerar alltid under fasta namnen cluster1 / admin@cluster1 (kubectl config set-cluster/set-credentials/set-context)
   - tar backup av ~/.kube/config (config.pre-talos-kubeconfig.<ts>) och städar bort kvarvarande cluster1-N / admin@cluster1-N
   - kör: talos-kubeconfig   (extra flaggor som --nodes <ip> skickas vidare till talosctl)

2) SHELL-GUARD
   - dotfiles/bash/.bashrc och ~/.zshrc definierar en funktion `talosctl` som fångar `talosctl kubeconfig` utan --merge=false, skriver en varning till stderr och tvingar --merge=false, samt pekar på talos-kubeconfig.
   - Funktionen anropar `command talosctl` så recursion undviks.
   - OBS: guarden ligger i dotfiles/bash/.bashrc - ändras den måste guarden kopieras till ~/.zshrc manuellt.

3) MINNESNOTIS + AGENTS.md
   - memory: fact kubeconfig_admin + lesson "använd aldrig talosctl kubeconfig rakt in i ~/.kube/config".
   - plan/AGENTS.md och infrastructure/pxe/Bootloaders/make-talos-sd-card.sh pekar på admin@cluster1.

Varför tre nivåer: skriptet ensamt räcker inte om någon kör talosctl direkt (t.ex. i ett CI-kommando eller från en agent). Guarden fångar den människan. Och minnesnotisen gör att en agent inte återupptäckter fällan via search.

Verifierat 2026-11-25: talos-kubeconfig kördes mot det live-klustret och admin@cluster1 fungerar (9 noder syns via kubectl --context admin@cluster1 get nodes).

## Relaterade

- [[notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat]]
- [[notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna]]
