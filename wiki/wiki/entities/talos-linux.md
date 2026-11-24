---
title: Talos Linux
type: entity
summary: Minimalt, API-drivet Linux-OS för Kubernetes; styrs deklarativt via maskinpatches.
tags: [bare-metal, baremetal, infrastructure, kubernetes, linux, pxe, talos]
sources: [infrastructure/talos/nodes.yaml, infrastructure/talos/patches/all.yaml, infrastructure/talos/patches/controlplane.yaml]
created: 2026-06-02
updated: 2026-10-24
status: active
related: [entities/kubernetes, entities/pxe-server, notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning]
---

## Sammanfattning

Talos är ett minimalt, API-drivet Linux-OS designat för Kubernetes-kluster. Kontrollpanel via `talosctl`, konfiguration via maskiner-patches. Kör uppgraderingar via `simon talos upgrade`. Stödjer både bare-metal och virtuella distributioner.

Talos är ett minimalistiskt Linux-OS för Kubernetes — operativsystemet är read-only och konfigurationen hanteras deklarativt via API. Alla noder styrs från en kontrollpanel (talosconfig) utan SSH.

## Detaljer

### Konfigurationsstruktur

```
talos/
├── nodes.yaml                # Nodtabell: namn, ip, roll, typ, ark, schematic-id
├── patches/
│   ├── all.yaml              # Gäller alla noder
│   ├── controlplane.yaml     # Endast control planes
│   ├── worker.yaml           # Endast workers
│   ├── controlplanes/        # Per-nod control plane-patches
│   └── workers/<NODNAMN>/
│       ├── patch.yaml        # Maskinspecifik config
│       └── disks/            # UserVolumeConfig för diskar
└── generated/                # Genererade maskinkonfiger (gitignoreras)
```

`generated/` är **gitignorerat** - innehållet har CA-nycklar och
kubelet-tokens. Det regenereras med `pxe/generate-node-configs.py` mot
klustrets aktuella PKI.

### Talos v1.14 - dokumentflytt

I Talos v1.14 har flera gamla fält flyttat till egna dokument. De
gamla `--oidc-*`-flaggorna i `KubeAPIServerConfig.extraArgs` avvisas
med "use KubeAuthenticationConfig". Se `talos/patches/all.yaml` och
`talos/patches/controlplane.yaml` för hur detta är hanterat, och
[[notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument]].

### Maskinpatches

Varje nodtyp kan ha specifika patches:

| Fil | Beskrivning |
|-----|-------------|
| `patch.yaml` | machine.type, install.diskSelector |
| `disks/*.yaml` | UserVolumeConfig för datadiskar |

### UserVolumeConfig

Talos hanterar diskar via `UserVolumeConfig` (v1alpha1):

```yaml
apiVersion: v1alpha1
kind: UserVolumeConfig
name: ssd-1
volumeType: disk                    # OBLIGATORISKT
provisioning:
  diskSelector:
    match: "'/dev/disk/by-id/usb-...' in disk.symlinks"
```

**Viktigt:**
- `volumeType: disk` krävs för vanliga diskar
- Stöd saknas för `minSize`, `maxSize`, `grow` i v1.12.5
- Disken monteras på `/var/mnt/<name>` efter boot

### Talos-kommandon

| Kommando | Beskrivning |
|----------|-------------|
| `simon talos upgrade` | Uppgradera alla noder |
| `simon talos update config <nod>` | Applicera patch-konfiguration |
| `talosctl -n <IP> get volumestatus` | Visa diskstatus med UUID |
| `talosctl -n <IP> mounts` | Visa alla mountade filsystem |
| `talosctl -n <IP> reset` | Wipea diskar vid felsökning |
| `talosctl -n <IP> get mc` | Visa machine config |
| `talosctl -n <IP> list /dev/disk/by-id` | Lista tillgängliga diskar |

### Nodtyper

| Typ | Ansvar |
|-----|--------|
| **controlplane** | Kubernetes control plane, etcd |
| **worker** | Applikationsworkloads |

### Multi-document patches

**Viktigt:** `JSON6902 patches` stöder INTE multi-document machine configuration (MachineConfig + UserVolumeConfig i samma fil).

**Korrekt metod:**
```bash
# 1. Generera basconfig
talosctl gen config ... -o generated/

# 2. Applicera node-specifik patch EFTER generering
talosctl machineconfig patch generated/worker.yaml \
  --patch "@patches/nodes/worker-1.yaml" -o generated/worker.yaml
```

### iSCSI Tools

Verifierade hashes för olika HW-plattformar:

| Platform | SHA256 |
|----------|--------|
| Radxa Rock 4C+ | `b2bda8664a9e63afcddc80b9d351ec25b6c465a5c143383e8aa579278c439b25` |
| Radxa Rock 4SE | `7eb05cfd849450051d4560f3fca869f8b9ad6ceec2543bb22041261845ed0684` |
| Raspberry Pi 4 | `ee21ef4a5ef808a9b7484cc0dda0f25075021691c8c09a276591eedb638ea1f9` |
| Mac Mini | `613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245` |

### PXE Boot

Stöd för nätverksboot via Matchbox. Används för:
- Rock Pi (Radxa)
- Raspberry Pi 4
- Andra nätverksbootbara enheter

## Flöde: Talos → Kubernetes

```
1. PXE boot via Matchbox
       ↓
2. Talos installerar på systemdisk
       ↓
3. Machine config appliceras (patches)
       ↓
4. Talos registrerar sig i klustret
       ↓
5. Kubernetes componenter startar
       ↓
6. FluxOperator syncar infrastruktur
```

## Relaterade

- [[notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning]]
- [[concepts/flux-operator]]
- [[concepts/gitops]]
- [[concepts/sops]]
- [[entities/flux-gitops]]
- [[entities/kubernetes]]
- [[entities/pxe-server]]
- [[summaries/longhorn-disk-guide]]
## Klusterstruktur
### Noder
| Namn | IP | Roll | Typ | Arkitektur |
|------|-----|------|-----|------------|
| controlplane-1 | 10.10.10.11 | controlplane | RPi4 | arm64 |
| controlplane-2 | 10.10.10.12 | controlplane | RPi4 | arm64 |
| controlplane-3 | 10.10.10.13 | controlplane | RPi4 | arm64 |
| worker-1 | 10.10.10.21 | worker | Rock4C+ | arm64 |
| worker-2 | 10.10.10.22 | worker | Rock4C+ | arm64 |
| worker-3 | 10.10.10.23 | worker | Rock4C+ | arm64 |
| worker-4 | 10.10.10.24 | worker | Rock4SE | arm64 |
| worker-5 | 10.10.10.25 | worker | Rock4SE | arm64 |
| worker-6 | 10.10.10.26 | worker | Rock4SE | arm64 |
| worker-7 | 10.10.10.27 | worker | Rock4SE | arm64 |
| worker-8 | 10.10.10.28 | worker | Rock4SE | arm64 |
| worker-9 | 10.10.10.29 | worker | MacMini | amd64 |
### VIP för controlplane
- **10.10.10.10** — Virtuell IP för API-servern
## Arkitektur
### Katalogstruktur
```
infrastructure/talos/
├── nodes.yaml              # Nod-definitioner (name, IP, role, typ)
├── node-upgrades.md        # Dokumentation om uppgraderingar
├── talosconfig             # Klient-konfiguration
├── patches/                # Machine config patches
│   ├── all.yaml           # Gäller alla noder
│   ├── controlplane.yaml  # Endast controlplanes
│   └── worker.yaml        # Endast workers
└── generated/              # Genererade config-filer
    ├── controlplane.yaml
    └── worker.yaml
```
### Machine config features
#### Controlplane features
- **KubePrism** (port 7445) — Intern kube-apiserver-proxy
- **OIDC** — Zitadel som auth provider
- **RBAC** — Rollbaserad åtkomstkontroll
- **hostDNS** — Forwardar kubeDNS till host
#### Worker features
- **iSCSI tools** — För Longhorn storage
- **Disk quota support** — För NFS/Longhorn
- **PoE-hantering** — Via PoE-port-nummer
### Extra manifests
```yaml
cluster:
  extraManifests:
    - https://raw.githubusercontent.com/alex1989hu/kubelet-serving-cert-approver/main/deploy/standalone-install.yaml
```
Installerar kubelet-serving-cert-approver för att automatiskt godkänna CSR:er för server-certifikat.
## Kommandon
```bash
# Uppgradera alla noder
simon talos upgrade
# Anslut till specifik nod
talosctl console <nodnamn> --talosconfig talosconfig
# Visa nod-status
talosctl get members --talosconfig talosconfig
```
