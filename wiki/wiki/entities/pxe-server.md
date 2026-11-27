---
title: PXE-server
type: entity
summary: PXE-servern i infrastructure-repot - dnsmasq, matchbox och mainline-U-Boot byggda för nätverksboot av Talos-noderna.
tags: [pxe, talos, uboot, matchbox, dnsmasq, boot, infrastructure]
sources: [infrastructure/pxe/README.md]
created: 2026-10-24
updated: 2026-11-27
related: [notes/dns-och-dhcp-har-skilda-agar-dnsmasq-pa-pxe-ager-dhcp-unifi-pa-10-10-10-1-ager-dns-och-worker-9-s-lease-hostname-ar-klientens-eget, notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc]
---

# PXE-server

## Definition

PXE-servern ligger i `infrastructure/pxe/` och bootar Talos Linux till
noderna i klustret över nätverket. Den består av två tjänster:

| Komponent | Port | Syfte |
|-----------|------|--------|
| **dnsmasq** | 67/UDP DHCP, 69/UDP TFTP | DHCP-leaser + TFTP |
| **matchbox** | 8080/TCP | iPXE-skript + statiska filer, matchar MAC → bootprofil |

En tredje del är **bootloadern**: mainline U-Boot byggs för ROCK 4C+ och
4SE samt Raspberry Pi 4, patchas för att prioritera PXE, och skrivs till
nodens egen flash.

## Bootkedjan

```
Nod (PXE i BIOS) → dnsmasq DHCP → TFTP bootfil
  → U-Boot (mainline, patchad: PXE först)
  → matchbox: iPXE-skript per nod (MAC → grupp → profil)
  → Hämtar nodens Talos-config via matchbox
  → Bootar Talos
```

Viktigt: en PXE-boot **skriver aldrig till disk av sig själv**. Profiler
pekar på *maintenance*-configs; det är U-Boot-patchen som styr om
installationen ska ske.

## Nodtyper

| Nod | Hårdvara | Boot-flöde |
|-----|----------|-----------|
| controlplane-1..3 | Raspberry Pi 4 | SPI-flash → U-Boot → PXE |
| worker-1..3 | Radxa ROCK 4C+ | eMMC → U-Boot → PXE |
| worker-4..8 | Radxa ROCK 4SE | eMMC → U-Boot → PXE |
| worker-9 | Mac mini | egen U-Boot, inte PXE |

## Filer

| Fil | Roll |
|-----|------|
| `pxe/dnsmasq.conf` | DHCP + TFTP |
| `pxe/matchbox/groups`, `pxe/matchbox/profiles` | MAC → grupp → profil |
| `pxe/generate-matchbox-profiles.py` | Genererar grupper/profiler från nodtabellen i scriptet |
| `pxe/generate-node-configs.py` | Genererar Talos-config mot klustrets **aktuella** PKI |
| `pxe/Bootloaders/` | Bygger och patchar U-Boot/TF-A |
| `pxe/test-dhcp-boot.sh` | Verifierar DHCP/TFTP |

`pxe/generate-node-configs.py` läser `talos/nodes.yaml` och
`talos/patches/workers/<nod>/`. Resultatet innehåller CA-nycklar och
kubelet-tokens och är **inte versionerat**.

## Fällor som upptäckts under arbetet

- **U-Boot-patchen måste sätta PXE först.** `boot_targets` styr steg 2,
  inte steg 1 - fel ordning gör att U-Boot ignorerar PXE.
- **RK3399 bootrom söker SPI/eMMC/SD/USB i fast ordning.** ROCK-plattorna
  har ingen SPI-flash i fabriken - U-Boot ligger på microsd/eMMC.
- **DHCP-boot kräver `mkimage -uimage`, inte ren hush-text.** Matchbox
  serverar ett iPXE-skript, inte ett sh-skript.
- **Genererade nodconfigs blir stale efter Talos-uppgradering.** Kör
  `generate-node-configs.py` igen varje gång PKI:n eller versionen ändras.
- **Matchbox 404ar på gamla nodconfigs** om filerna inte regenererats.
- **PXE-boten gör en nod till en PXE-nod.** Matchbox-profilen pekar på
  maintenance-mode; utan rätt patch fortsätter den att hämta configen.

## Verifikationsstatus 2026-11-13

Serverkedaan ar verifierad fungerande, men noderna tar inte den:
- PXE-varden = Radxa-board 10.10.10.2 (MAC 2c:cf:67:43:f3:b0, UniFi-port 8),
  inte en Raspberry Pi. dnsmasq (DHCP+TFTP, network_mode host) + matchbox (:8080).
- Deployade nodconfigs har AKTUELL cluster-CA och endpoint https://10.10.10.10:6443.
- Efter PoE-cykel av worker-1/3/4/5/6/7: 1 DHCP-par per nod (bootloader + OS med
  hostname), **noll** TFTP-forrfragningar och **noll** matchbox-requests.
  Noderna bootar sin lokala installation, inte natet.
- Korrigering till tidigare "portstat kan inte lasas": PoE per port styrs via
  `PUT /proxy/network/api/s/default/rest/device/{device_id}` med
  `{"port_overrides":[{"port_idx":N,"poe_mode":"off"|"auto"}]}`. `cmd/devmgr
  power-cycle` svarar rc=ok men gor ingenting - falsk framgang.
- Kvarvarande glapp: dnsmasq satter inte DHCP option 54 (next-server), och
  tftp/rot har endast 4C+-DTB (4SE-noderna far fel DTB).

Se [[notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device]].

## Relaterade

- [[notes/dns-och-dhcp-har-skilda-agar-dnsmasq-pa-pxe-ager-dhcp-unifi-pa-10-10-10-1-ager-dns-och-worker-9-s-lease-hostname-ar-klientens-eget]]
- [[notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc]]
- [[entities/talos-linux]]
- [[entities/kubernetes]]
- [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]]
- [[notes/rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc]]
- [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]]
- [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
- [[notes/u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr]]
- [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]]
- [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]]
- [[notes/pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen]]
- [[notes/pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text]]
- [[notes/pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd]]
- [[notes/pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg]]
- [[notes/pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404]]
- [[notes/pxe-aterstallning-nodspecifika-matchbox-profiler-och-maintenance-mode-flode]]
- [[notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen]]
- [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]]
- [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]]
- [[notes/korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image]]
- [[notes/controlplane-recovery-via-pxe-ar-inte-konfigurerad]]
- [[notes/dnsmasq-dhcp-boot-tar-bara-en-tagg-anvand-tag-if-for-villkor]]
- [[notes/tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes]]
- [[notes/rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas]]
- [[notes/mellanslag-i-radxa-rock-4c-saboterar-byggen-losning-bind-mount]]
- [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]]
- [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]]
- [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]]
- [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]]
- [[notes/worker-1-svarar-inte-i-natet-felsokningsstatus]]
- [[notes/worker-2-utan-strom-port-22-up-false-poe-power-0]]
- [[notes/simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern]]
- [[notes/simon-pxe-flash-clean-code-omgang-2-fokuserade-funktioner]]
- [[notes/bash-fallgropar-i-pxe-flash-tab-som-ifs-whitespace-och-read-med-for-fa-variabler]]
