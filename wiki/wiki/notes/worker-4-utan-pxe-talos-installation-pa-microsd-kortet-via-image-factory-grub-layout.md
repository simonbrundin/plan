---
title: worker-4 utan PXE: Talos-installation pa microSD-kortet via Image Factory (GRUB-layout)
type: note
tags: [worker-4, rock4se, talos, pxe, image-factory, beslut, projekt:infrastructure]
created: 2026-11-22
updated: 2026-11-22
status: active
---

Beslut 2026-11-20 (svar pa anvandaren: "jag vill inte anvanda pxe, bara att den ska boota fran sitt sd-kort").

Grundproblem: nodens NVMe-installation ar trasig. Talos v1.14 installerar pa arm64
alltid med systemd-boot, och ROCK 4C+/4SE:s U-Boot saknar EFI-tjansten SetVariableRT
("efivarfs: Firmware does not support SetVariableRT"). Installen avbryter i
installBootloader() - partitionerna skapas men META forblir tom och STATE blir
oanvandbar. Samtidigt fangar ESP:n pa disken noden i U-Boot: boot_targets =
"mmc1 mmc0 nvme scsi usb pxe dhcp spi" gor att disken bootas fore pxe, och
systemd-boot startar utan LoaderEntryDefault -> ingen DHCP, ingen API, ingen svar.
Symtom: endast port 10250 (kubelet med trasig cert), 50000 "connection refused".

Losning utan PXE: Talos Image Factory-imaget for ROCK 4SE (schematic
7eb05cfd..., v1.14.1, overlay siderolabs/sbc-rockchip/rock4se) laggs pa
microSD-kortet. Viktiga egenskaper:
  - imagen ar GRUB-baserad (EFI + BIOS-partition med core.img + BOOT i XFS),
    inte systemd-boot -> inga EFI-variabler behovs, alltsa inget
    SetVariableRT-problem.
  - forsta partitionen borjar pa sektor 22528, sa Rockchips loader-omrade
    (idbloader @ sektor 64, u-boot.itb @ 0x7f8000) ligger fore partition 1
    och overlever dd. Verifieras fore/efter med RK33-signatur + FIT-magic.
  - STATE och EPHEMERAL skapas av Talos volymhanterare vid forsta boot
    (systemdisken = disken med META).
Flode: BootROM -> kortets U-Boot -> ESP: BOOTAA64.EFI (GRUB) -> Talos ->
maintenance mode -> apply-config via API -> noden skriver configen till STATE
och går med i cluster1. Ingen PXE, ingen matchbox, ingen dnsmasq.

Nytt verktyg: pxe/Bootloaders/make-talos-sd-card.sh (laddar ner imagen,
verifierar loader-omradet, skriver och verifierar).

Blockerare: PXE-Pi:n (10.10.10.2) har hangt - ingen SSH/ping, alla portar
timeout. Den ar labbets DHCP-server (dhcp-range 10.10.10.50-200, lease 12h),
sa nya enheter far ingen lease tills Pi:n startas om. Pi:n har dessutom
 kortlasare och USB-NVMe-adapter => den naturliga skrivaren.
