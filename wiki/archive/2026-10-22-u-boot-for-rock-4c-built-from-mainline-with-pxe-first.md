---
title: U-Boot for ROCK 4C+ built from mainline with PXE first
type: note
tags: [pxe, uboot, rock4cplus, radxa, mainline, projekt:simon]
---

The repo's U-Boot image for ROCK 4C+ actually contained a ROCK Pi 4A DTB (model = 'Radxa ROCK Pi 4A') despite the filename. Wrong DTB means the board never boots - no console, no network - which is indistinguishable from a dead board.

Resolution: build from mainline per https://docs.u-boot.org/en/latest/board/rockchip/rockchip.html:
  export CROSS_COMPILE=aarch64-linux-gnu-
  export BL31=../trusted-firmware-a/build/rk3399/release/bl31/bl31.elf
  make rock-4c-plus-rk3399_defconfig && make
  sudo dd if=u-boot-rockchip.bin of=/dev/sdX seek=64 conv=fsync sync

Key learnings:
- mainline has its own RK3399 DRAM init (drivers/ram/rockchip/sdram_rk3399.c), so no vendor DDR blob is needed. The DDR blob inside the official inindev idbloader does NOT match any rkbin rk3399_ddr_*.bin - it is mainline's own TPL.
- The board's DTB comes from mainline's rk3399-rock-4c-plus.dts - authoritative for the board.
- boot_targets and boot_script_dhcp are set cleanly in include/configs/rockchip-common.h and rk3399_common.h, no need to patch env strings in a binary (the env blob has a fixed size, so an earlier hack replaced kernel_comp_size=0x8000000).
- include/configs/rockchip-common.h also defines the canonical partition map: loader1 @32K, loader2 @8M, trust @12M, boot @16M, rootfs rest.
- TF-A needs arm-none-eabi-gcc for the m0 co-processor, aarch64-linux-gnu-gcc for the rest.

Artifacts live in pxe/Bootloaders/Radxa Rock 4C+/ and are built by pxe/Bootloaders/build-mainline-uboot.sh with pxe/Bootloaders/patches/pxe-first-boot.patch. pxe/Bootloaders/verify-uboot-image.sh checks RK33/FIT/DTB-board/env in 5 seconds and would have caught the 4A mix-up.
