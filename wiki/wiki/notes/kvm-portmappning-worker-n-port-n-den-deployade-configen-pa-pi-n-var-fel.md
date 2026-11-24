---
title: KVM-portmappning: worker-N = port N - den deployade configen pa Pi:n var fel
type: note
tags: [kvm, bug, projekt:infrastructure]
created: 2026-11-20
updated: 2026-11-20
status: active
---

2026-11-20: Den KVM-portmappning som var deployad pa PXE-Pi:n var FEL och ledde
till en missvisande diagnos ("worker-1 visar svart skarm" - men det var port 4
= worker-2 som visades).

Ratt mapping enligt anvandaren (och enligt kvm/kvm-switch-config.example.json
i repot):

    pxe        -> port 16
    worker-1   -> port 1
    worker-2   -> port 2
    worker-3   -> port 3
    ...
    worker-9   -> port 9

Controlplane-1..3 ar INTE inkopplade i KVM:n (portarna ar fria).

Det som var fel pa Pi:n (/etc/kvm-switch/config.json): controlplane-1..3 ->
1,2,3 och worker-1..9 -> 4..12. Backup: config.json.bak-2026-11-20.

Kommandon:
    # byt till en port
    curl -X POST http://10.10.10.2:8099/port/3 -d '{}'
    # se mappningen
    curl http://10.10.10.2:8099/ports
