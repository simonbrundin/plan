---
title: simon pxe flash - flashar PXE-först-U-Boot till SD-kort i PXE-servern
type: note
summary: Nytt kommando i simon-cli: `simon pxe flash` (scripts/pxe.sh, main_pxe_flash).
tags: [pxe, cli, uboot, feature, projekt:simon]
created: 2026-10-23
updated: 2026-10-23
status: active
related: [entities/pxe-server, notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot, notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist, notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]
---

Nytt kommando i simon-cli: `simon pxe flash` (scripts/pxe.sh, main_pxe_flash).

Varför: rescue-kort till worker-noderna skapades manuellt med
pxe/flash-uboot.sh + lokal scp. Kommandot gör hela flödet (välj kort ->
välj datortyp -> verifiera bild -> dd seek=64 -> verifiera sha256) i ett
kommando, riktat mot PXE-servern där kortet sätts i.

Flöde:
1. Listar blockenheter pa servern via lsblk -P, markerar systemdisk och
   "ingen media" (kortlasare utan kort).
2. Sparar systemdisken: root_src fran findmnt -> PKNAME.
3. Vyer: Radxa Rock 4C+ (mainline U-Boot 2026.10 med pxe-first-boot.patch),
   Radxa Rock 4SE (upstream-build), Raspberry Pi 4 (skriver inget - satt
   UEFI-bootordning med efibootmgr pa noden istallet).
4. Verifierar bilden med Bootloaders/verify-uboot-image.sh innan flash
   (RK33-signatur, board-DTB, boot_targets, boot_script_dhcp).
5. scpar bilden och skriver med `dd bs=512 seek=64 conv=fsync` - byte 32768
   ar Rockchip-bootROM:ens offset for idbloader.
6. Laser tillbaka och jAmfor sha256 med kallfilen.

Lardom:
- `ssh` utan -n slakar stdin, sa bekräftelseprompten ("ja") forsvann. Alla
  ssh-anrop som inte skickar ett heredoc-script behover -n. Heredoc-anrop
  far INTE ha -n (da skickas scriptet inte fram).
- scp har inget -n-flagga.
- PXE-serverns fjarrskal ar Nushell: komplexa kommandon maste wrappas i
  `sh -c '...'` eller skickas som `ssh pxe 'sh -s' <<EOF`.

## Relaterade

- [[entities/pxe-server]]
- [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]]
- [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]]
- [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]]
