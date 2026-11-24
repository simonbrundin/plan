---
title: Workers bootar inte via PXE trots fungerande PXE-server - UniFi PoE-styrning via rest/device
type: note
tags: [pxe, unifi, workers, felsokning, projekt:infrastructure]
---

Status 2026-11-13: PXE-servern (10.10.10.2, UniFi-switch port 8, hostname 'pxe', MAC 2c:cf:67:43:f3:b0) var igang och serverkedaan ar helt verifierad fungerande - men INGEN worker bootar via PXE.

## Serverkedaan: verifierad OK
- dnsmasq: DHCP 0.0.0.0:67 + TFTP 0.0.0.0:69 (network_mode host) -> nabar for hela 10.10.10.0/24.
- matchbox :8080 -> /boot.ipxe + profilmatchning (worker-3 MAC 6a:4e:e8:63:ad:5c -> node-worker-3).
- /assets/talos-config/nodes/10.10.10.21-29.yaml finns och har AKTUELL cluster-CA (Talos-CA sha256 B8:D2:AF:CE:A9:9F:EB:56:87:C0:B5:71:9F:81:9A:24:20:DA:3B:2D:12:12:F2:53:96:58:44:5E:E1:58:2E:FB) + endpoint https://10.10.10.10:6443.
- TFTP av talos/bsc = giltig legacy uImage 'talos-pxe' (2906 byte). Kernel/initramfs/dtb finns i tftp/rot.
- Live dnsmasq.conf pa Pi:n == pxe/dnsmasq.conf i repot.

## UniFi PoE-styrning: sa fungerar det (viktig larm)
- FUNGERAR: PUT /proxy/network/api/s/default/rest/device/{device_id} med
  {"port_overrides":[{"port_idx":N,"poe_mode":"off"|"auto"}]} -> HTTP 200 rc=ok.
- FUNGERAR INTE med samma nyckel: stat/portup, stat/port, stat/device/{id}, stat/arp (api.err.NotFound/UnknownDevice), cmd/stpoe (HTTP 400 api.err.Invalid).
- FALLA: cmd/devmgr {"cmd":"power-cycle","mac":...,"port_idx":N} svarar rc=ok men byter INTE PoE (noden startar inte om). Anvand rest/device + port_overrides.
- Switch: US 48 PoE 500W, id 3446ecf8-11d0-4a70-b78e-ce17983b5429, MAC 18:e8:29:2e:f0:93.
- Porttabellen (stat/device) ger per port: speed, poe_good, last_connection -> anvandbar for att se vilka boardar som ar link-up.

## Resultat av power-cykel (worker-1,3,4,5,6,7 = port 21,23,24,25,26,27)
- Ingen node hamtade nagot via TFTP, matchbox fick 0 requests -> ingen PXE-boot.
- Efter varje omstart: exakt ETT DHCP-par (utan hostname = bootloadern, ~1 s later med hostname worker-N) -> noden bootar sin LOKALA installation, inte nattet.
- worker-1: ingen DHCP alls efter omstart (matchar noten 2026-10-27: boarden skickar noll paket).
- worker-4: far lease men svarar inte pa ping. worker-5/6/7: lease + ping (ping via ARP-cache kan lura - anvand DHCP/TFTP-loggar).
- Kontrollplanes (10.10.10.11-13) svarar normalt pa Talos-API; deras klientcert (aktuell CA) accepteras. Workers svarar EOF pa 50000 -> de kor gamla PKI-generationen fran lokal installation.

## Verifierade glapp (i config, inte i noden)
1. dnsmasq satter INTE DHCP option 54 (next-server) -> U-Boot som anvander DHCP-bootfilen har ingen server-IP att TFTP:a fran.
2. Endast 4C+-DTB (rk3399-rock-4c-plus.dtb) finns i tftp/rot -> 4SE-noderna (worker-4..8) får fel DTB (kernel-panic) om de bootar via samma bsc.
3. Nodens U-Boot-env maste ha boot_script_dhcp=talos/bsc OCH boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi (stock: mmc1 mmc0 nvme scsi usb pxe dhcp spi -> natet ar 6e alternativet). Verifiera med printenv pa noden (kräver fysisk atkomst).

## Vad som saknas for att fa noderna in i cluster1
Fysisk atkomst (HDMI/serie) for att satta boot_targets/boot_script_dhcp eller flasha om kortet med PXE-first-image (pxe/Bootloaders/Radxa Rock 4C+/u-boot-rockchip-pxe-first.bin, patch-sd-card.sh; 4SE: fetch-pxe-first.sh). Serverkedjan behover ingen atergardering.
