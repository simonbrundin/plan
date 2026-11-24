---
title: Kanata switch-session: tmuxinator-fönster öppnades inte
type: note
tags: [bug, kanata, tmux, tmuxinator, projekt:simon]
---

Symtom: tangenterna i kanatas tmux-lager (t.ex. 'k8s', 'dot') skapade bare tmux-sessioner med 1 fönster istället för tmuxinator-layouten.

Tre rotorsaker i ~/repos/dotfiles/kanata/.config/kanata/switch-session.sh:
1. 'tmuxinator list' skriver ut i TRE kolumner, så 'grep -xF "Kubernetes"' matchade aldrig (raden var 'Homelab  Infrastructure  Kubernetes'). Projektet hittades aldrig -> bare session. Fix: 'tmuxinator list -n' (en per rad).
2. 'tmuxinator start X -d' är inte ett giltigt flagg i tmuxinator 3.x (-d skickas vidare som projekt-argument), och tmuxinator försöker.attacha -> 'open terminal failed: not a terminal' i kanatas tty-less context. Fix: --no-attach.
3. 'tmux has-session -t Home' prefix-matchar och träffar t.ex. Homelab. Fix: '=name' (exakt matchning).

Följd: de gamla bare sessionerna (Simon CLI, Dotfiles, MatchMaker) låg kvar med 1 fönster, och eftersom skriptet bara startar tmuxinator OM sessionen saknas, kördes tmuxinator aldrig -> fönstren kom aldrig. Nu jämförs fönsternamn/antal mot yaml och en naken session (<=1 fönster) återskapas från tmuxinator.

Övriga lärdomar:
- Ruby/tmuxinator failar med 'invalid byte sequence in US-ASCII' pa yaml med icke-ASCII om LANG inte satts -> sätt LANG/LC_ALL i skriptet.
- tmuxinator --append (3.x) lagger in fönstren i den AKTUELLA klientens session, inte i projektets session -> lat INTE anvanda --append.
- Scriptet loggar nu till ~/.cache/kanata-switch-session.log.
