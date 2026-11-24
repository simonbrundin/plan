---
title: Recovery av en ROCK-worker kräver fysisk rättning: ESP-filen på eMMC fångar booten
type: note
tags: [pxe, talos, rock4c, recovery, pxe-first, projekt:infrastructure]
---

En worker som en gång installerats till eMMC kan inte nås via API igen utan
fysisk åtgärd. Skriptet pxe/recover-worker.sh (permanent i infra-repot) hanterar
allt utom själva den fysiska rättningen.

Varför API:t försvinner

Installationen skriver EFI/BOOT/BOOTAA64.efi och EFI/Linux/Talos-v1.14.1.efi
till ESP-paketen innan EFI-variablerna skrivs. När ESP finns på eMMC vinner
U-Boots bootflow mmc0 framför pxE (boot_targets=mmc1 mmc0 nvme scsi usb pxe
dhcp spi). Noden startar systemd-boot utan LoaderEntryDefault och hänger utan
DHCP och utan API - bara port 50000/443 svarar med TLS-fel.

Förutsättningen är därför att rescue-kortet
(pxe/Bootloaders/Radxa Rock 4C+/u-boot-rockchip-pxe-first.bin) är flashat,
vilket sätter boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi. Utan det
kommer ingen recovery automatik fram.

Vad recovery-skriptet gör

recover-worker.sh slår upp noden i talos/nodes.yaml (namn -> ip, poe),
power-cyklar via UniFi om API:t inte svarar, verifierar att rätt nod svarar
(hostname måste stämma, samma skydd som vid hostname: auto-incidenten), tömmer
systemdisken med talosctl reset --wipe-mode=system-disk --reboot
--graceful=false, applicerar worker-N-install.yaml, följer
unattendedinstallstatus, applicerar tillbaka worker-N.yaml och verifierar
volymer, maskinstatus och medlemskap i cluster1.

Nyckelpunkt: efivarfs-felet på ROCK är väntat och icke-fatalt. Partitionerna
(EFI/META/STATE/EPHEMERAL) skapas innan installBootloader() fallerar, så
skriptet fortsätter även när unattendedinstallstatus blir failed.

Inga hemligheter i skriptet: UniFi-nyckeln läses från UNIFI_API_KEY, annars
via op read.

Förbättring av wiki-noten om efivarfs

Den tidigare noten sade att install-config aldrig får appliceras på ROCK.
Det var fel: partitionerna MÅSTE skapas för att noden ska kunna gå med i
cluster1 (STATE -> EPHEMERAL -> image-cache -> kubelet). Rätt ordning är
wipe -> install-config (reboot: false) -> maintenance-config -> behåll
PXE-first.
