---
title: Talos DNS-fallback mot publika resolvrar
type: note
tags: [talos, dns, felsokning, projekt:infrastructure]
---

Flera Talos-noder loggade 'error serving dns request' med read udp <nod>:<port>->10.10.10.1:53: i/o timeout och DiscoveryService rapporterade tillfälligt server misbehaving. ResolverStatus visade endast [10.10.10.1]. Direkt från system-containrar fungerade däremot 10.10.10.1, 1.1.1.1 och 8.8.8.8.

Åtgärd 2026-11-20: ändrade source of truth i infrastructure/pxe/generate-node-configs.py så ResolverConfig använder 10.10.10.1 först med 1.1.1.1 och 8.8.8.8 som fallback. Ändrade även talos/patches/all.yaml. Genererad worker-3-konfiguration validerades med talosctl validate --mode metal. Patchen applicerades utan reboot på controlplane-1/2/3 och worker-3; ResolverStatus bekräftar alla tre resolvrarna. Kubernetes-noderna förblev Ready.

Övriga worker-noder som inte hade fungerande Talos-API kunde inte patchas live, men framtida regenerering/applicering får den nya fallback-konfigurationen.
