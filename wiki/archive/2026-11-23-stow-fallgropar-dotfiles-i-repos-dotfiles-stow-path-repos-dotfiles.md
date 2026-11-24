---
title: Stow-fallgropar: dotfiles i ~/repos/dotfiles (stow_path = repos/dotfiles)
type: note
tags: [dotfiles, stow, bug, session, projekt:dotfiles]
---

Stow-paketen ligger i ~/repos/dotfiles och stowas till $HOME (~/.stowrc: --target=$HOME, --dir=$HOME/repos/dotfiles). Då är stow_path = 'repos/dotfiles' RELATIVT till hemmappen, vilket har två viktiga konsekvenser:

1) Absoluta symlänkar ägs ALDRIG av stow ('existing target is not owned by stow'). En symlänk som pekar med absolut sökväg (/home/simon/repos/...) måste skrivas om till relativ form.

2) En fil i paketroten (t.ex. pi/.pattern-retry.json -> ~/.pattern-retry.json) MÅSTE ha symlänken 'repos/dotfiles/pi/.pattern-retry.json' (UTAN '../'). Med '../' kan stow inte äga den: stow gör abs2rel(stow_dir, target) = 'repos/dotfiles' och kanonpathar bara bort '..' när prefixet matchar stow_path. '../repos/...' matchar inte 'repos/dotfiles/' -> konflikt varje gång. Filer en halv nivå ner (t.ex. ~/.config/hypr, ~/.config/secrets/load.sh) fungerar med '../' eftersom då kanonpathar bort '../..' och prefixet matchar.

Felsökning: 'stow -vvvvv --simulate <paket>' visar raden 'is symlink destination X owned by stow?' - där syns exakt varför ägandet fallerar.

Fix i denna session: (a) pi/.pattern-retry.json togs bort ur repot - pi-pattern-retry-tillägget var redan avinstallerat (commit f8d492e 'ta bort rate-limit/pattern-retry'), filen var död vikt och låste stow pi; (b) ~/.bashrc och ~/.config/secrets/load.sh var INTE stow-länkade (bashrc: en riktig fil som hade glidit från repot och innehöll en hårdkodad MINIMAX_API_KEY; load.sh: absolut symlänk). .bashrc ersattes nu av symlänken 'repos/dotfiles/bash/.bashrc' och innehållet mergades till repot - hemligheten flyttades till ~/.config/secrets/minimax_api_key (mode 600, utanför git) och load.sh läser den cache-filen (cache-först-mönster som NEON_API_KEY).

Verifiering: 'stow --simulate <alla paket från README>' -> inga konflikter; bash -ic 'echo ${#MINIMAX_API_KEY}' -> 125 tecken, NEON -> 69.
