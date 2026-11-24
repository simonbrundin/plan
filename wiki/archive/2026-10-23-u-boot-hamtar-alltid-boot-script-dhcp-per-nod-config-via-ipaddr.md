---
title: U-Boot hämtar alltid boot_script_dhcp - per-nod-config via ${ipaddr}
type: note
tags: [pxe, uboot, talos, bug, projekt:simon]
---

Diagnos av varför worker-2 inte PXE-bootade efter nyflashat kort.

ROTORSAK 1 (hårdvara): worker-2:s nätverksport (UniFi port 22) var up=false
med poe_power=0.00 W - noden fick ingen ström, alltsa ingen DHCP och ingen
TFTP. Kontrolleras via UniFi API: /proxy/network/api/s/default/stat/device
(port_table[].up, .poe_power). PoE-cykel pa porten hjalte inte - problemet
sit i strumforsorjningen/kabln, inte i PXE-kedjan.

ROTORSAK 2 (server): tre fel hittades i PXE-konfigurationen:
- talos/bsc saknades pa servern (omdopt till bsc.disabled), sa flodet var
  avstangt.
- dnsmasq gav bara worker-1 talos/bsc. worker-2/3 fick boot.ipxe, som
  U-Boot inte kan tolka.
- bsc.hush hade worker-1:s talos.config-URL hardkodad.

LARDOM: U-Boot hämtar ALLTID samma boot-script. Filnamnet kommer fran
env-variabeln boot_script_dhcp i U-Boot-imagen, INTE fran DHCP:s bootfil.
I boot/bootmeth_script.c: fname = env_get("boot_script_dhcp");
dhcp_run(addr, fname, true). I net/lwip/dhcp.c sats bootfile bara fran
DHCP-s svar om explicit_file ar falskt - dhcp_run med filnamn gor det
explicit, sa DHCP:s bootfil ignoreras. En per-nod-fil hade kravt en
per-nod-U-Boot-image och ny flashning per nod.

LARNING: U-Boot-script kan inte villkorligt vala config. itest finns
men ar NUMERISK (arithcomp); itest.s/test saknas i binaren, sa
strangjamforelse ar omojlig. Losningen blev att lata config-URL:en
byggas av nodens DHCP-adress: talos.config=http://10.10.10.2:8080/
assets/talos-config/nodes/${ipaddr}.yaml. Matchbox serverar configen
under nodens IP-adress; filerna genereras av pxe/deploy-node-configs.sh
(kopior av <nod>.yaml).

KONTROLL: ipaddr-env finnes i U-Boot (strings u-boot), ethaddr med
kolon fungerar inte i filnamn (riskabelt), IP-adress utan kolon gor det.
