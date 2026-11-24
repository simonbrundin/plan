---
title: simon kvm: CLI-kommando för att byta HDMI-KVM-port från PXE-värd
type: note
tags: [kvm, feature, cli, projekt:simon]
---

2026-11-20: Nytt CLI-kommando 'simon kvm' i simon-cli (scripts/kvm.sh + tests/kvm-test.sh).

Varför: KVM:en (16-portars HDMI/RS232, se notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm) styrs via kvmctl.py på PXE-Pi:n. Utan CLI:t var portbytet manuellt via curl mot kvm-api.py på :8099, och felaktig portmappning gav missvisande diagnoser (notes/kvm-portmappning-worker-n-port-n-den-deployade-configen-pa-pi-n-var-fel).

Viktiga designval:
- Portmappningen läses från /home/simon/repos/infrastructure/kvm/kvm-switch-config.example.json (env KVM_PORTS_CONFIG), inte från en hardkodad lista. Därför hamnade felmappningen (controlplane-1..3 -> 1,2,3) inte i klienten.
- kvmctl.py skickas via stdin till 'ssh -T pxe python3 - switch <port>', så kommandot fungerar även om /usr/local/bin/kvmctl.py inte hunnit installeras på Pi:n.
- Web UI (10.10.10.81) öppnas i ett befintligt Chromium-fönster (fokus via hyprctl focuswindow), aldrig nytt fönster/ny flik: kontrollen av en redan öppen flik sker via CDP-endpointen 127.0.0.1:9222, med Hyprland-fönstertiteln ('KVM') som fallback när DevTools-porten inte är igång.
- fzf-listan visar alla 16 portar med --height=100% så hela listan syns samtidigt.

Verifiering: 14 tester i tests/kvm-test.sh (fzf-input, ssh-anrop, stdin-kvmctl, CDP/Hyprland-återbruk) gröna.
