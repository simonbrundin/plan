---
title: U-Boot körs EN gång: boot_targets styr steg 2, inte steg 1
type: note
tags: [pxe, uboot, boot, fakta, projekt:simon]
---

Vanlig feltolkning: 'kortet -> U-Boot -> U-Boot igen -> PXE'. Saadan ar det inte. Kedjan ar (1) Rockchip BootROM -> idbloader -> u-boot.itb -> U-Boot proper, och U-Boot las ALLTID fran bootmedium (microSD/eMMC) - det ar steg 1 och det paverkas inte av boot_targets. (2) U-Boot kor EN gang och valjer sedan start via 'bootflow scan -lb', som provar boot_targets i ordning och startar den FORSTA giltiga bootflowen (en som nadde 'loaded'-state).

Verifierat mot upstream v2025.04: label_to_uclass() i boot/bootdev-uclass.c mapper varje post i boot_targets till en bootdev-etikett: mmc1/mmc0 -> MMC, nvme -> NVMe, usb -> UCLASS_USB (mappas till UCLASS_MASS_STORAGE), spi -> UCLASS_SPI_FLASH, och specialfallen 'pxe' -> UCLASS_ETH med BOOTFLOW_METHF_PXE_ONLY samt 'dhcp' -> UCLASS_ETH med BOOTFLOW_METHF_DHCP_ONLY. Alltsa: pxe/dhcp ar INTE bootmedia utan metoder pa nartverksenheten - det ar POSITIONEN i listan som avgor om de provas fore eller efter diskarna. Ingen av dem har BOOTMETHF_GLOBAL-flaggan i v2025.04 (bara vbe_simple och bootmeth_efi_mgr har den), sa ordningen i boot_targets galler fullt ut.

Konsekvens for noderna: med defaulten 'mmc1 mmc0 nvme scsi usb pxe dhcp spi' startar NVMe:n (Talos) alltid forra - och USB (5:e) naas aldrig, for att NVMe (3:e) redan startat. En installation som startar en karna som sedan kraschar ar fortfarande 'giltig' och blockerar PXE; PXE naas bara av en installation som inte ens får igang en karna.

Viktigare fyllning: mmc1 (kortet) scannasocksa som bootkalla - ett kort med en Talos-installation bootas (fordomande av NVMe), medan ett kort med bara U-Boot ger ingen giltig bootflow och hoppas over.
