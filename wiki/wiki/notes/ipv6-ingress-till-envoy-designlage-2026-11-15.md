---
title: "IPv6-ingress till Envoy: designläge 2026-11-15 (klustret är nere)"
type: note
tags: [ipv6, envoy, gateway-api, metallb, dns, cloudflare, starlink, projekt:infrastructure]
created: 2026-11-15
updated: 2026-11-17
status: active
related: [notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink]
---

## Vad användaren vill

`*.simonbrundin.com` → publik IPv6 → vidarebefordras till Envoy Gateway, analogt med
UniFi-regeln "HTTPS to Envoy Gateway" (IPv4-DNAT WAN:443 → 10.10.10.101:443).

## Viktiga fynd (2026-11-15)

1. **Produktionsklustret är BART.** Klustret på 10.10.10.11-13 (via 10.10.10.10) har bara
   kube-system - **ingen Flux, ingen Envoy, ingen cert-manager, inga CRD:er**, inga workers.
   MetalLB-IP:t 10.10.10.101 svarar inte (tcp/80,443 ger inget svar). Allt ingress-lager
   är alltså ned just nu. Se [[notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device]].
2. **IPv4-vägen är redan död.** Cloudflare: `*.simonbrundin.com` CNAME →
   `krokviken.duckdns.org` (DNS-only), DuckDNS A → `100.67.39.22` = UDM:ns CGNAT-WAN-IP.
   Ingen inkommande IPv4 under CGNAT → domänerna kan inte nås utifrån idag.
3. **UniFi kan inte skapa brandväggsregler via API** på denna version:
   `rest/firewallrule` svarar `api.err.InvalidPayload`/`InvalidValue` oavsett payload →
   zonbrandväggen får sättas i UI:t.
4. **Cloudflare-API:t fungerar** (token i `components/cert-manager/cloudflare-api-token-secret.enc.yaml`,
   age-nyckel i 1Password `op://Private/SOPS/password`). Zone-ID:
   `8a047802cd4e53ef77ef78b08f15e477`.

## Tekniska byggblock (verifierade via docs/källor)

- UniFi `single_network`: **bara ett LAN** kan ha Starlink-/64:en. För Envoy (i
  Datacenter-VLANet 10.10.10.0/24) måste `single_network_lan` peka på Datacenter.
- Envoy Gateway: `EnvoyProxy` kan sätta `ipFamily: DualStack` +
  `envoyService.annotations.metallb.io/loadBalancerIPs: <ipv4>,<ipv6>` → Envoy-tjänsten får
  en **stabil publik IPv6-VIP** från MetalLB (MetalLB stöder IPv6 i L2-läge via NDP).
  Gateway: `spec.addresses` + `infrastructure.parametersRef` → EnvoyProxy.
- **Pod-IPv6 krävs** för MetalLB-IPv6 (kube-proxy DNATar VIP→pod). Kräver dual-stack CNI:
  Talos `cluster.network.podSubnets/serviceSubnets` med IPv6 + flannel `EnableIPv6`
  + omstart av noder. Tungt och klumpigt i Talos.
- Alternativ utan CNI-ändring: **Envoy DaemonSet/Deployment med `hostNetwork: true`**
  (`envoyDeployment.patch` StrategicMerge + `dnsPolicy: ClusterFirstWithHostNet` +
  `useListenerPortAsContainerPort: true` för att binda 80/443). Då binder Envoy nodens
  IPv6 direkt. Kräver en **statisk publik IPv6** (t.ex. `2a0d:3341:a000:6841::101`) på
  noden via Talos `machine.network.interfaces[].addresses` → stabilt AAAA-mål.
- Envoy Gateway kan också köras som DaemonSet (`envoyDaemonSet` i EnvoyProxy) → alla
  3 controlplanes kan bära samma statiska IPv6.

## Historik: hur ingressen fungerade FÖRE Starlink (Arkaden)

- **Arkaden** gav UDM:n en **riktig publik IPv4 direkt pa WAN: 62.108.213.168** (ingen CGNAT).
  Verifierat mot RIPE/RDAP 2026-11-15: `netname: SE-ARKADEN-20231221-2`,
  range `62.108.208.0-62.108.215.255`, role `ARKADEN ROLE`.
  (Stugan har istallet **Telia** med publik IPv4 `78.69.91.181` - två olika leverantorer.)
- Kedjan: Cloudflare `*.simonbrundin.com` CNAME -> krokviken.duckdns.org (A -> 62.108.213.168,
  DDNS uppdaterad var 5:e min) -> UniFi port forward 443 -> 10.10.10.101 (MetalLB) ->
  Envoy -> HTTPRoute -> tjanst. Verifierat med 4-stegstestet i
  [[concepts/http-routing-flow]] (Steg 1-4).
- Starlink bytet (efter 2026-10-24) tog bort grunden: WAN blev 100.67.39.22/18 via
  100.64.0.1 (CGNAT 100.64.0.0/10), Starlinks publika IP 212.105.145.11 delas med andra
  kunder. DuckDNS uppdateras nu till CGNAT-adressen -> den publika vagen har varit dod
  sedan bytet, inte "idag".
- Tva separata hinder med Starlink: (a) CGNAT, (b) Starlinks egen router har ingen
  port-forward-funktion alls -> krav pa bypass-lage + IP Policy "Public IP" (Priority).
- Stugan har fortfarande Telia med publik IPv4 (78.69.91.181) och fungerande port-forward
  -> beviset pa att "publik IPv4 + port forward" var hela receptet hemma.

## Inkommande IPv6 via Starlink: fungerar i princip (CGNAT gäller bara IPv4)

- Starlink delar **IPv4** (100.64.0.0/10) mellan kunder -> inkommande IPv4 blockeras.
  **IPv6 har ingen NAT**: /64 (WAN, SLAAC) och /56 (PD) ar routbara till kunden.
- Verifierat 2026-11-15: Talos-noden fick GUA och nagde `api64.ipify.org` som svarade med
  nodens egen adress -> prefixet ar routerat till oss och internet accepterar kalladresser.
  In- och utgaende trafik anvander samma routing, sa vagen for inkommande till /64:en
  ar etablerad.
- **Ej verifierat:** initierad (osynkron) paketleverans utifran. Tre hinder:
  (a) Starlinks egen router star i vagen (ingen brandvaggconfig, okant om den slapper igenom
  oinitierad IPv6) -> **bypass-lage tar bort osakerheten**, (b) UniFis zonbrandvagg
  släpper bara established/related inat -> explicit regel kravs (gor i UI),
  (c) ingen extern IPv6-vynk finns (arbetsstationen pa Netgear/Starlink-sidan har bara
  ULA `fc00:f853:ccd:e793::1`, ingen IPv6-route; Netgear-admin svarar 401 pa 192.168.1.1).
- Plan for verifiering: sla pa IPv6 pa Netgear-sidan -> extern vynk (annan Starlink-PoP),
  ping6 UDM:s WAN-GUA, och efter UniFi-regel TCP mot Envoy.
- Starlink-kontot finns i 1Password (Private/"Starlink", user+pass) -> kan loggas in for att
  se vilken plan och om "IP Policy: Public IP"-knappen finns (endast Local/Global Priority).


## MetalLB och IPv6 (verifierat 2026-11-15 mot metallb.io/docs + källkod)

- MetaLB stoder IPv6: *"IPv6 and dual stack services are supported in L2 mode, and in BGP mode
  only via the FRR-based modes"*. Vi kor L2-lage (L2Advertisement, ingen FRR) -> **stods utan
  extra mode**. Version i repot: metallb chart 0.14.5.
- L2-annonsering for IPv6 sker via **NDP** (inte ARP): "MetalLB responds to ARP requests for
  IPv4 services, and NDP requests for IPv6" (metallb.io/concepts/layer2/).
- KRAV: "there must be at least one address pool having both addresses of version v4 and v6"
  -> ip-pool.yaml maste innehalla BADE v4- och v6-intervall i samma IPAddressPool.
- For dual-stack services kan `spec.loadBalancerIP` inte anvandas (rak for flera IP:n) ->
  annotationen `metallb.io/loadBalancerIPs: "10.10.10.101,2a0d:3341:a000:6841::101"`.
- **MetalLB har två roller:** (1) tilldela VIP till Envoy-Servicen, (2) annonsera VIP:n i LAN
  (NDP). **Leveransen till podden gor kube-proxy (DNAT VIP -> podens IP)** -> for en IPv6-VIP
  maste poden ha en **IPv6-adress** => dual-stack CNI. MetalLB ar alltsa INTE blockeraren;
  pod-IPv6 (flannel EnableIPv6/Cilium + podSubnets/serviceSubnets + reboot) ar det.
- Känd risk (metallb#2685): NDP-respondern använder link-local som källadress och kan faila
  med "ipv6-icmp bind: cannot assign requested address" pa interface där den globala
  IPv6-adressen inte ar den forsta -> testa nar noden har bade SLAAC-adress och statisk ::101.
- Envoy Gateway: `EnvoyProxy` med `spec.ipFamily: DualStack` +
  `envoyService.annotations.metallb.io/loadBalancerIPs`; Gateways `infrastructure.parametersRef`
  pekar pa EnvoyProxy.

## Viktigt: IPv4-ingress har ALDRIG fungerat under Starlink

- Starlinks router har ingen port-forward-funktion och CGNAT (100.64.0.0/10) blockerar
  inkommande IPv4. UDM:s WAN = 100.67.39.22/18 via gateway 100.64.0.1 => ingen publik IPv4.
  DuckDNS A-record pekar pa CGNAT-adressen, sa "jag har IPv4 igang"-kanslan bygger pa
  den forra ISP:n. Public IPv4 ING krav bypass-lage pa Starlink + IP Policy "Public IP"
  (endast Local/Global Priority).
- Bravare IPv4-vag: (1) Starlink Priority + bypass + Public IP -> **noll klusterandrorr**,
  allt idag fungerar igen; bonus: Starlinks reservationssystem behaller da och da IPv6-/64-prefixet.
  (2) Cloudflare Tunnel -> repot har redan `components/cloudflared/` (token + cert),
  aggregatorn har kommentaren "Ev losning ifall jag inte far till en fast publik ipadress".
  (3) IPv6 enligt nedan.
- Anmarkning: Starlinks router blockar port-forwarding -> brandvaggning maste ske i UniFi
  (vilket redan ar fallet: "HTTP/HTTPS to Envoy Gateway").

## Rekommenderad väg (att besluta om)

- **B (lätt):** /64 i Datacenter-VLANet + statisk IPv6 `::101` på controlplane-noderna +
  Envoy som DaemonSet med hostNetwork + Cloudflare wildcard AAAA → `::101`.
  Ingen CNI-migrering. Nackdel: Envoy utan k8s-Service i vägen, TLS-terminering oförändrad.
- **A (robust):** dual-stack CNI + MetalLB-IPv6-pool + EnvoyProxy `ipFamily: DualStack`.
  Nackdel: podSubnets/serviceSubnets + flannel IPv6 + reboot av controlplanes.

## Övrigt

- AAAA + CNAME på samma namn är **inte tillåtet** i Cloudflare → wildcard-CNAMET mot
  duckdns måste tas bort och ersättas med wildcard A + wildcard AAAA.
- Starlink kan byta /64-prefix → både statisk nodadress och AAAA-posten måste uppdateras.
- Inkommande IPv6 måste släppas in i UniFis zonbrandvägg (WAN → Datacenter, TCP 80/443
  + ICMPv6) - det finns ingen NAT att dölja sig bakom, alla klienter blir globala.

## Relaterade

- [[notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink]]

