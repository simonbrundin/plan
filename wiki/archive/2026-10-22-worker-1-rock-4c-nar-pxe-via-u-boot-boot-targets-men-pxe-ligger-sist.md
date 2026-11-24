---
title: worker-1 (Rock 4C+) når PXE via U-Boot boot_targets, men PXE ligger sist
type: note
tags: [pxe, talos, uboot, worker, projekt:simon]
---

Radxa ROCK 4C+ (worker-1..3) bootar via U-Boot i SPI-flash, inte UEFI. boot_targets ur pxe/Bootloaders/rock4c-plus.bin är 'mmc1 mmc0 nvme scsi usb pxe dhcp spi' och bootcmd='bootflow scan -lb'. PXE är alltså alternativ 6 av 8 och nås bara om inget bootbart media (SD/eMMC/NVMe) startar. En installation som startar men kraschar blockerar alltså fortfarande PXE - bara en installation som inte startar alls faller vidare. Vill man ha PXE först krävs U-Boot-konsol: 'setenv boot_targets pxe dhcp mmc1 mmc0 nvme scsi usb' + saveenv, eller 'bootflow menu' + 'bootflow select <nr>' (bootdelay=2). Detta är dokumenterat i pxe/UBOOT.md och pxe/README.md.
