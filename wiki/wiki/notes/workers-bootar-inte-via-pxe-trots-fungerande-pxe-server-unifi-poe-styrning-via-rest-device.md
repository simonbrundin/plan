---
title: Workers bootar inte via PXE trots fungerande PXE-server - UniFi PoE-styrning via rest/device
type: note
tags: [pxe, unifi, workers, felsokning, projekt:infrastructure]
created: 2026-11-13
updated: 2026-11-20
status: active
related: [entities/pxe-server, notes/worker-1-startar-om-med-strom-men-skickar-noll-paket-fysisk-omstart-bevisar-samma-fel]
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

## worker-1 (port 21): bevisat att boarden inte KOR alls (2026-11-15)
Power-cykel 13:39:58Z (PoE off 8 s -> auto) med 5 s-poll av switch-porten:
- t=0-10 s: speed=100 (gammalt cached-lage), t=15-35 s: **speed=0, poe_good=False** (stromen brots),
  t=25-35 s: poe_good=True men speed=0, **t=40 s: speed=100 igen** -> PHY-lanken kommer upp ~15 s efter strom.
- Under hela fomen (13:39:58-13:42:00): **noll** DHCP-rader for MAC 0a:3f:39:25:2b:c8,
  **noll** TFTP-forrfragningar, **noll** matchbox-requests. Boarden forsoker inte ens DHCP.
- Switch-port 21 (livslangsraknare): **rx_packets=0** (aldrig ett enda mottaget paket),
  mac_table_count=0, **poe_class=Class 0 / poe_power=1.84 W** (jmf port 23: Class 4 / 4.29 W),
  hastighet 100 Mbit (jmf 1000 Mbit pa fungerande boardar).
- SLUTSATS: worker-1 har strom (PHY/Pmic lever) men SoC:n startar inte - den kan inte
  forsoka PXE for att den inte kover. Detta ar HARDVARA, inte PXE-kedjan.
  Port 22 (worker-2) och 28 (worker-8): speed=0, 0 W, rx_packets=0 -> ingen strom alls.

## Porttabell (UniFi stat/device, 2026-11-15)
| port | nod | speed | PoE W | klass | rx_pkt | tx_pkt | macs | conn |
|---|---|---|---|---|---|---|---|---|
| 21 | worker-1 | 100 | 1.84 | Class 0 | **0** | 2666014 | 0 | False |
| 22 | worker-2 | 0 | 0.00 | Unknown | 0 | 118 | 0 | False |
| 23 | worker-3 | 1000 | 4.29 | Class 4 | 26743636 | 19553796 | 1 | True |
| 24 | worker-4 | 0 | 5.43 | Class 4 | 14318553 | 11662563 | 0 | False |
| 25 | worker-5 | 1000 | 4.44 | Class 4 | 27424786 | 20042191 | 1 | True |
| 26 | worker-6 | 1000 | 3.75 | Class 4 | 30367984 | 21802382 | 1 | True |
| 27 | worker-7 | 1000 | 3.68 | Class 4 | 30395968 | 21801285 | 1 | True |
| 28 | worker-8 | 0 | 0.00 | Unknown | 0 | 148 | 0 | False |
| 29 | worker-9 | 0 | 0.00 | Unknown | 0 | 0 | 0 | False |
(tx_pkt = switchens utgaende paket mot porten; rx_pkt = portens inkommande paket)

## Samtidig test av TVÅ workers (2026-11-15 13:48:47Z)
worker-3 (ROCK 4C+, port 23) + worker-4 (ROCK 4SE, port 24) startades om
SAMTIDIGT med EN kombinerad PUT {"port_overrides":[{23,off},{24,off}]} -> 10 s -> auto.
- Strömmen bröts verkligen pa bada: port 23 W 4,29->0,44 (speed 0), port 24 W 4,39->0,00.
- Lanken kom tillbaka ~60 s efter PoE on (W 5,01 / 4,71).
- DHCP: worker-3 kl. 13:49:54/55, worker-4 kl. 13:49:57/59 (bootloader + OS med hostname).
- **Noll** TFTP-forrfragningar, **noll** matchbox-requests -> ingen av dem forsoker PXE.
- MAL: ping svarade "OK" 35-50 s efter PoE-on trots att noden var dods - ARP-cache
  lurar. Anvand dnsmasq-loggen/switchens W for att veta nar noden faktiskt lever.

## UniFi-API: port_overrides ersatter HELA listan
Ett PUT med port_overrides for EN port raderar overrides for alla andra portar.
Forsta forsoket at starta om 23 och 24 "samtidigt" med två separata PUT misslyckades:
den andra PUT:n raderade override for port 23, sa worker-3 aldrig tappade strom
(och fick darfor ingen DHCP alls). RATT: alla portar i samma array i ett enda PUT.

## Test av SAMTLIGA workers samtidigt (2026-11-15 14:11:08Z)
En enda PUT med port_overrides for port 21-28 (alla "off") -> 12 s -> alla "auto".
Port 29 (worker-9) undanskildes: har ingen PoE (poe_enable=False) och startas manuellt via iBoot.
- Alla 8 portar: speed 0 / 0,00 W under ~35 s, sedan länk+W tillbaka ~45 s efter PoE on.
- DHCP-par (bootloader + OS med hostname) for worker-3/4/5/6/7 mellan 14:12:21 och 14:12:28.
- worker-1: **fortfarande noll DHCP-rader** även efter färsk power-cykel -> boarden dor verkligen.
- worker-2 och worker-8: inget alls (ingen strom).
- **Noll TFTP-forrfragningar och noll matchbox-requests for hela gruppen.**
- Slutstatus: worker-3..7 svarar pa ping men Talos-API:et ger "connection refused" pa 50000
  (de kor alltsa INTE en fungerande Talos - den lokala installationen är trasig/gammal).
  worker-1/2/8/9: "no route to host". Klustret har fortfarande noll worker-noder.

## Nytt test 2026-11-20 (worker-1, port 21): fortfarande noll paket
Fresh PoE-cykel pa port 21 (13:06 lokal tid): `port_overrides` off -> 25 s -> auto,
samtidigt tcpdump pa PXE-varden (eth0) och poll av switch-porten.
- **Noll** paket fran `0a:3f:39:25:2b:c8` i tcpdump: ingen DHCP, ingen ARP-svar,
  ingen TFTP. Endast broadcast ARP-fragor "who-has 10.10.10.21" fran controlplane-1/2/3.
- Switch port 21 efter cykeln: `up=True speed=100 poe=1.83W class=Class 0 rx_packets=0 macs=0`.
- Matchar 2026-10-27 och 2026-11-15: boarden får strøm och PHY-lank (100 Mbit),
  men SoC:n kommer aldrig igang till U-Boot-nätverket. Detta går INTE att løsa
  från serversidan.

## Serverkedjan verifierad igen 2026-11-20 (allt OK)
- `~/PXE/dnsmasq.conf` pa Pi:n == `pxe/dnsmasq.conf` i repot (byte-for-byte).
- TFTP svarar lokalt pa Pi:n: `tftp://127.0.0.1/talos/bsc` -> 2970 byte (legacy uImage "talos-pxe").
- matchbox: `/assets/talos-config/nodes/10.10.10.21.yaml` har md5 5aa1e827... == repots
  `worker-1.yaml` (maintenance-config, `disk.serial == "0xFFFFFFFF"`).
- TFTP-rot: `talos/rock4c-plus/kernel-arm64` (19 MB), `talos/rock4c-plus/initramfs-arm64.xz`
  (87 MB), `dtb/rockchip/rk3399-rock-4c-plus.dtb` finns pa plats.
- Deployad `talos/bsc` = WIPE-varianten for worker-1 (2026-10-27): bootargs har
  `talos.experimental.wipe=system` nar `${ipaddr} = 10.10.10.21`. **Maste bytas tillbaka
  mot `bsc.hush` nar worker-1 ar sanad**, annars wipear noden sin egen systemdisk varje boot.
- `pxe/Bootloaders/Radxa Rock 4C+/u-boot-rockchip-pxe-first.bin` verifierad med
  `verify-uboot-image.sh`: RK33 OK, FIT OK, DTB "Radxa ROCK 4C+",
  `boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi`, `boot_script_dhcp=talos/bsc`,
  U-Boot 2026.10. Rescuedagen är redo att skriva.

## KVM: konsol for worker-1 (2026-11-20)
`kvm-api` (systemd-tjanst `kvm-api` pa PXE-Pi:n, port 8099) har nod-karta:
RATT KART: worker-N = port N (worker-1 = port 1 ... worker-9 = port 9, pxe = port 16);
controlplane-1..3 ar inte inkopplade i KVM:n. Den deployade configen pa Pi:n var
tidigare FEL (worker-1 = port 4) - se noten om KVM-portmappningen 2026-11-20.
`curl -X POST http://10.10.10.2:8099/port/3 -d '{}'` byter till port 3,
så monitorn visar worker-1:s HDMI-konsol (U-Boot/Talos). Detta ar den enda vägen
att se vad U-Boot gör pa boarden. Ingen HDMI-capture finns pa Pi:n (/dev/video* är
bara Pi:ns egna CSI/ISP-noder), så bilden måste läsas av en människa.

## FALLGROP (bekräftad live 2026-11-20)
Ett PUT med ENDA port 21 i `port_overrides` raderade overrides för port 22-28 -
då tappade de portarna `native_networkconf_id` (datacenter-VLAN:et ec2a295b...)
och hamnade i Default-nätet. Återställning: ETT PUT med port 21-28 i samma
array, varje rad med `poe_mode`, `forward:"all"` och `native_networkconf_id`.
Kontrollera efteråt med `stat/device` att port 21-28 fortfarande har
`native_networkconf_id=ec2a295b-...`.

## Nasta hinder fore worker-1 (2026-11-20)
worker-3/4/5/6/7 gor ett DHCP-par (bootloader + OS) och sedan tyst - ingen TFTP,
ingen matchbox-request, Talos-API "connection refused" pa .23/.25. De bootar sin
LOKALA installation, inte natet, for att deras U-Boot-env har nätet sist i
`boot_targets`. Aven en frisk board far alltsa INTE PXE-boota for att env:en
satte boot_targets. Den fysiska insatsen maste gora bada delar:
1. fa boarden att overhuvudtaget fa en U-Boot-prompt (KVM-konsol / serie),
2. satt `boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi` + `boot_script_dhcp=talos/bsc`,
   eller flasha rescue-kortet med `u-boot-rockchip-pxe-first.bin`.

## 2026-11-20: 75 s-lang PoE-cykel + blank HDMI = boarden dor
- PoE av 75 s (full avlastning) -> auto: fortfarande **0 paket** i tcpdump pa PXE-varden
  (pcap 24 byte = tomt), port 21: speed=100, poe=1.67 W, poe_class=Class 0, rx_packets=0.
- UniFi `last_connection` for `0a:3f:39:25:2b:c8`: connected=false, **last_seen 2026-10-23**
  (det var sista gangen boarden syntes pa natet - dagen rescuekortet skrevs).
- KVM port 4: **ingen bild pa monitorn** - MEN det var port 4 = worker-2 enligt
  den RATTA mappningen, sa slutsatsen om worker-1 gick inte att dra. Port 1
  (worker-1) var aldrig testad da. KONTOROLL: byt KVM till port 1
  (controlplane-1) - om det också blir svart ar det monitorn/kedjan, inte boarden.
  (Talos skriver inte till HDMI; mainline-U-Boot gor det, men bara i sekunder.)
- TOLKNING: PHY:n negotiates 100 Mbit pa egen hand (MAC:en ar inte initierad) och PoE
  visar Class 0 / ~1.7 W (fungerande boardar: Class 4 / 4.3 W). SoC:n startar alltsa
  inte -> ingen HDMI-utsignal, ingen nätverksstack. Detta ar en trasig/dod board.
- Nasta steg (fysiskt): (1) verifiera skarm+kedja via KVM port 1, (2) lysa pa LED och
  kontrollera PoE-kabel/barreljack, (3) kallstart via RESET, (4) testa rescue-microSD,
  (5) multimeter pa SoC-rails (5 V / 3.3 V / 1.1 V) och pa PoE-ingangen.


## Relaterade

- [[notes/worker-1-startar-om-med-strom-men-skickar-noll-paket-fysisk-omstart-bevisar-samma-fel]]
- [[entities/pxe-server]]
