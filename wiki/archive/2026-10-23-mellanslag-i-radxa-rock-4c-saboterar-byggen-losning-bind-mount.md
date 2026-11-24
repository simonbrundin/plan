---
title: Mellanslag i 'Radxa Rock 4C+' saboterar byggen - losning: bind-mount
type: note
tags: [ubygg, uboot, verktyg, felsokning, projekt:infrastructure]
---

Katalogen 'Radxa Rock 4C+' i repot innehaller mellanslag. Det bryter bade GNU Make och U-Boots binman:

  make: *** No rule to make target '/home/simon/repos/infrastructure/pxe/Bootloaders/Radxa'.
  binman: error: unrecognized arguments: Rock 4C+/trusted-firmware-a/build/rk3399/release/bl31/bl31.elf

Symlink racker INTE - make slutar upp den fysiska slojvagen. Losningen var bind-mount till mellanslagsfria slojvagar:

  sudo mkdir -p /mnt/rock4c-plus-build/{tfa,uboot}
  sudo mount --bind '<repo>/Radxa Rock 4C+/trusted-firmware-a' /mnt/rock4c-plus-build/tfa
  sudo mount --bind '<repo>/Radxa Rock 4C+/u-boot' /mnt/rock4c-plus-build/uboot

Artefakterna hamnar i repot (mounten skriver till samma inoder). BL31:s slojvag maste vara mellanslagsfri for att binman ska kunna tolka argumenten.

pxe/Bootloaders/build-mainline-uboot.sh gör nu detta automatiskt: den valjer repotratden om den finns (annars klona till ~/tools/u-boot-build), mountar bada traden, bygger TF-A (make CROSS_COMPILE=aarch64-linux-gnu- PLAT=rk3399 bl31), bygger U-Boot och verifierar.

Fallgropar i skriptet: sudo skriver sin prompt till stdout, sa i  maste sudo anropas med >&2 - annars hamnar 'Please touch the FIDO authenticator.' i slojvagen.

Kalltraden for trusted-firmware-a/ och u-boot/ ligger nu i repot och ar git-ignorerade (78 MB + 1 GB). Nya artefakt-hashar: u-boot-rockchip-pxe-first.bin e8505e54..., idbloader.img 493a26e8..., u-boot.itb 6732fa11....
