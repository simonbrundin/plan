# Wiki Log

> **Kronologisk historik** över alla operationer. Append-only.

Format: `## [YYYY-MM-DD] type | Title`

---

## [2026-06-03] update | HTTP Routing Flow
- [update] [[concepts/http-routing-flow]] — utökade felsökningssektionen med alla 4 steg-tester och förklaringar av adresserna
  - Lade till: 10.10.10.101, 62.108.213.168, krokviken.duckdns.org, grafana.simonbrundin.com
  - Förklarade vad varje adress representerar och vad testet verifierar
  - Lade till felsökningsmatris för att identifiera vilket steg som misslyckas
- [update] index.md med nytt koncept
- [update] [[log]] (denna fil)

## [2026-06-03] init | Setup
- [init] Wiki initialiserad baserat på Karpathys LLM Wiki-pattern
- [init] Skapade AGENTS.md schema
- [init] Skapade index.md och log.md

## [2026-06-03] create | Flux Components (eftermiddag)
- [create] [[entities/kubernetes]] — entity-sida med komplett klusterarkitektur, alla Flux/ArgoCD-komponenter
- [create] [[concepts/cert-manager]] — TLS-certifikat med Let's Encrypt och Cloudflare DNS
- [create] [[concepts/longhorn]] — distribuerad storage med CSI, replicas, och backup
- [create] [[concepts/metallb]] — L2 load-balancing för LoadBalancer-tjänster
- [create] [[concepts/kube-prometheus]] — observabilitetsstack med Prometheus, Grafana, Alertmanager
- [create] [[concepts/envoy-gateway]] — Gateway API-implementation med HTTP-routing
- [create] [[concepts/cloudnative-pg]] — PostgreSQL-operator med HA och backups
- [create] [[concepts/kubelet-csr-approver]] — automatisk CSR-godkännande för noder
- [create] [[concepts/cloudflare-tunnel]] — säker tunnel utan öppna portar
- [create] [[concepts/authentik]] — identity provider med OIDC/SAML SSO
- [create] [[summaries/flux-components]] — sammanfattning av alla Flux-hanterade HelmReleases
- [update] Uppdaterade [[entities/talos-linux]] med mer detaljer (PXE boot, multi-document patches, iSCSI tools)
- [update] index.md med 4 nya entities, 12 koncept, 4 sammanfattningar

## [2026-06-02] init | Setup
- [init] Wiki initialiserad baserat på Karpathys LLM Wiki-pattern
- [init] Skapade AGENTS.md schema
- [init] Skapade index.md och log.md

## [2026-06-02] create | Flux Components (eftermiddag)
- [create] [[entities/kubernetes]] — entity-sida med komplett klusterarkitektur, alla Flux/ArgoCD-komponenter
- [create] [[concepts/cert-manager]] — TLS-certifikat med Let's Encrypt och Cloudflare DNS
- [create] [[concepts/longhorn]] — distribuerad storage med CSI, replicas, och backup
- [create] [[concepts/metallb]] — L2 load-balancing för LoadBalancer-tjänster
- [create] [[concepts/kube-prometheus]] — observabilitetsstack med Prometheus, Grafana, Alertmanager
- [create] [[concepts/envoy-gateway]] — Gateway API-implementation med HTTP-routing
- [create] [[concepts/cloudnative-pg]] — PostgreSQL-operator med HA och backups
- [create] [[concepts/kubelet-csr-approver]] — automatisk CSR-godkännande för noder
- [create] [[concepts/cloudflare-tunnel]] — säker tunnel utan öppna portar
- [create] [[concepts/authentik]] — identity provider med OIDC/SAML SSO
- [create] [[summaries/flux-components]] — sammanfattning av alla Flux-hanterade HelmReleases
- [update] Uppdaterade [[entities/talos-linux]] med mer detaljer (PXE boot, multi-document patches, iSCSI tools)
- [update] index.md med 4 nya entities, 12 koncept, 4 sammanfattningar

## [2026-06-02] ingest | Flux-dokumentation
- [ingest] Syntetiserade Flux/GitOps-information från infrastruktur-repo
- [create] Skapade [[entities/flux-gitops]] — entity-sida med katalogstruktur och components
- [create] Skapade [[concepts/flux-operator]] — concept-sida med FluxInstance CRD och skillnader mot vanilla Flux
- [create] Skapade [[concepts/gitops]] — koncept-sida med GitOps-principer
- [create] Skapade [[concepts/sops]] — koncept-sida med SOPS/age-hemligheter
- [create] Skapade [[summaries/infrastructure-flux]] — sammanfattning av infrastructure-flux/README.md
- [update] Uppdaterade index.md med nya sidor
- [update] index.md med nya koncept efter Kyverno CEL-uppdatering

## [2026-06-02] ingest | Kyverno CEL
- [ingest] Flyttade dokumentation om Kyverno CEL expressions till concepts/
- [update] [[concepts/kyverno-policies]] uppdaterad med CEL-beskrivning

## [2026-10-21] init | Wiki (globalisering)
- [init] Wikin flyttad till `~/repos/agents/wiki` och symlänkad som `~/wiki`
- [init] `wikictl` byggd om för den kuraterade strukturen: `entities/`, `concepts/`, `summaries/`, `notes/`
- [create] Nio seed-anteckningar om typetrainers arkitektur, session, algoritm, kända buggar och konventioner
- [create] Anteckning om hur agentinstruktionerna deployas via dotfiles och stow
- [init] Global `AGENTS.md` med de två wikinreglerna (via `dotfiles/pi/.pi/agent/AGENTS.md`)
- [init] `wikictl init` körd i infrastructure, plan, dashboard, foodflow, matchmaker, mailmaster och simon-cli
- [update] index.md och graph.md byggdes om av `wikictl index`
- [lint] En trasig länk hittad: [[concepts/oidc]] saknar en sida

## [2026-10-21] create | Ingest
- [create] [[notes/test-av-global-flode]] — ny sida från inboxen
- [archive] 2026-10-21-test-av-global-flode.md
- [merge?] concepts/Cloudflare-Tunnel + concepts/Http-Routing-Flow (delade taggar: cloudflare, dns)
- [merge?] concepts/FluxOperator + concepts/GitOps (delade taggar: flux, gitops, kubernetes)
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/FluxOperator + concepts/gitops (delade taggar: flux, gitops, kubernetes)
- [merge?] concepts/FluxOperator + entities/Flux GitOps (delade taggar: flux, gitops, kubernetes)
- [merge?] concepts/FluxOperator + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] concepts/FluxOperator + summaries/Flux-Components (delade taggar: flux, gitops, kubernetes)
- [merge?] concepts/FluxOperator + summaries/infrastructure-flux (delade taggar: flux, gitops, kubernetes)
- [merge?] concepts/GitOps + concepts/flux-operator (delade taggar: kubernetes, flux)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/GitOps + entities/Flux GitOps (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/GitOps + entities/flux (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/GitOps + summaries/Flux-Components (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/GitOps + summaries/infrastructure-flux (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/Kyverno Policies + concepts/kyverno-cel (delade taggar: kyverno, policy, kubernetes, security, cel)
- [merge?] concepts/Kyverno Policies + kyverno-cel (delade taggar: kyverno, policy, kubernetes, security, cel)
- [merge?] concepts/Longhorn + summaries/longhorn-disk-guide (delade taggar: longhorn, storage)
- [merge?] concepts/flux-operator + concepts/gitops (delade taggar: flux, kubernetes)
- [merge?] concepts/flux-operator + entities/Flux GitOps (delade taggar: flux, kubernetes)
- [merge?] concepts/flux-operator + entities/flux (delade taggar: flux, kubernetes)
- [merge?] concepts/flux-operator + summaries/Flux-Components (delade taggar: flux, kubernetes)
- [merge?] concepts/flux-operator + summaries/infrastructure-flux (delade taggar: flux, kubernetes)
- [merge?] concepts/gitops + entities/Flux GitOps (delade taggar: gitops, flux, kubernetes)
- [merge?] concepts/gitops + entities/flux (delade taggar: gitops, flux, kubernetes)
- [merge?] concepts/gitops + summaries/Flux-Components (delade taggar: gitops, flux, kubernetes)
- [merge?] concepts/gitops + summaries/infrastructure-flux (delade taggar: gitops, flux, kubernetes)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Flux GitOps + summaries/Flux-Components (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Flux GitOps + summaries/infrastructure-flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)
- [merge?] entities/Talos Linux + summaries/longhorn-disk-guide (delade taggar: talos, kubernetes)
- [merge?] entities/flux + entities/talos (delade taggar: kubernetes, infrastructure)
- [merge?] entities/flux + summaries/Flux-Components (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/flux + summaries/infrastructure-flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/talos + summaries/longhorn-disk-guide (delade taggar: talos, kubernetes)
- [merge?] notes/globala-agentinstruktioner-deployas-via-dotfiles + notes/kunskaps-wiki (delade taggar: agent, wiki)
- [merge?] summaries/Flux-Components + summaries/infrastructure-flux (delade taggar: flux, gitops, kubernetes)

## [2026-10-21] update | Ingest
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/pi-desktop-notis-nar-agent-ar-klar]] — ny sida från inboxen
- [archive] 2026-10-22-pi-desktop-notis-nar-agent-ar-klar.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/controlplane-recovery-via-pxe-ar-inte-konfigurerad]] — ny sida från inboxen
- [create] [[notes/kanata-switch-session-tmuxinator-fonster-oppnades-inte]] — ny sida från inboxen
- [create] [[notes/pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen]] — ny sida från inboxen
- [create] [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]] — ny sida från inboxen
- [archive] 2026-10-22-controlplane-recovery-via-pxe-ar-inte-konfigurerad.md
- [archive] 2026-10-22-kanata-switch-session-tmuxinator-fonster-oppnades-inte.md
- [archive] 2026-10-22-pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen.md
- [archive] 2026-10-22-worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/dnsmasq-dhcp-boot-tar-bara-en-tagg-anvand-tag-if-for-villkor]] — ny sida från inboxen
- [archive] 2026-10-22-dnsmasq-dhcp-boot-tar-bara-en-tagg-anvand-tag-if-for-villkor.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] update | Ingest
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]] — ny sida från inboxen
- [archive] 2026-10-22-pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc]] — ny sida från inboxen
- [archive] 2026-10-22-rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]] — ny sida från inboxen
- [create] [[notes/korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image]] — ny sida från inboxen
- [archive] 2026-10-22-gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode.md
- [archive] 2026-10-22-korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]] — ny sida från inboxen
- [archive] 2026-10-22-u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]] — ny sida från inboxen
- [archive] 2026-10-22-pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]] — ny sida från inboxen
- [archive] 2026-10-22-rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404]] — ny sida från inboxen
- [archive] 2026-10-22-pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]] — ny sida från inboxen
- [archive] 2026-10-22-u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/pxe-aterstallning-nodspecifika-matchbox-profiler-och-maintenance-mode-flode]] — ny sida från inboxen
- [archive] 2026-10-22-pxe-aterstallning-nodspecifika-matchbox-profiler-och-maintenance-mode-flode.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-22] create | Ingest
- [create] [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]] — ny sida från inboxen
- [create] [[notes/worker-1-svarar-inte-i-natet-felsokningsstatus]] — ny sida från inboxen
- [archive] 2026-10-22-u-boot-for-rock-4c-built-from-mainline-with-pxe-first.md
- [archive] 2026-10-22-worker-1-svarar-inte-i-natet-felsokningsstatus.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]] — ny sida från inboxen
- [archive] 2026-10-23-worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes]] — ny sida från inboxen
- [archive] 2026-10-23-tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg]] — ny sida från inboxen
- [archive] 2026-10-23-pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]] — ny sida från inboxen
- [archive] 2026-10-23-worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/arbetsstationen-ar-omarchy-arch-apt-get-finns-inte-anvand-pacman-yay]] — ny sida från inboxen
- [archive] 2026-10-23-arbetsstationen-ar-omarchy-arch-apt-get-finns-inte-anvand-pacman-yay.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/mellanslag-i-radxa-rock-4c-saboterar-byggen-losning-bind-mount]] — ny sida från inboxen
- [archive] 2026-10-23-mellanslag-i-radxa-rock-4c-saboterar-byggen-losning-bind-mount.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]] — ny sida från inboxen
- [archive] 2026-10-23-worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas]] — ny sida från inboxen
- [archive] 2026-10-23-rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/pi-tog-bort-pi-rate-limit-och-pi-pattern-retry-fixade-omarchy-temasy]] — ny sida från inboxen
- [create] [[notes/pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text]] — ny sida från inboxen
- [create] [[notes/pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd]] — ny sida från inboxen
- [create] [[notes/simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern]] — ny sida från inboxen
- [archive] 2026-10-23-pi-tog-bort-pi-rate-limit-och-pi-pattern-retry-fixade-omarchy-temasy.md
- [archive] 2026-10-23-pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text.md
- [archive] 2026-10-23-pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd.md
- [archive] 2026-10-23-simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/bash-fallgropar-i-pxe-flash-tab-som-ifs-whitespace-och-read-med-for-fa-variabler]] — ny sida från inboxen
- [archive] 2026-10-23-bash-fallgropar-i-pxe-flash-tab-som-ifs-whitespace-och-read-med-for-fa-variabler.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/simon-pxe-flash-clean-code-omgang-2-fokuserade-funktioner]] — ny sida från inboxen
- [archive] 2026-10-23-simon-pxe-flash-clean-code-omgang-2-fokuserade-funktioner.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen]] — ny sida från inboxen
- [create] [[notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument]] — ny sida från inboxen
- [archive] 2026-10-23-pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen.md
- [archive] 2026-10-23-talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr]] — ny sida från inboxen
- [create] [[notes/worker-2-utan-strom-port-22-up-false-poe-power-0]] — ny sida från inboxen
- [archive] 2026-10-23-u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr.md
- [archive] 2026-10-23-worker-2-utan-strom-port-22-up-false-poe-power-0.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/infrastructure-agents-md-uppdaterad-tre-fallor-for-agenter]] — ny sida från inboxen
- [create] [[notes/pi-durable-demo-paketet-earendil-works-pi-durable-fungerar]] — ny sida från inboxen
- [archive] 2026-10-23-infrastructure-agents-md-uppdaterad-tre-fallor-for-agenter.md
- [archive] 2026-10-23-pi-durable-demo-paketet-earendil-works-pi-durable-fungerar.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/pi-durable-1-4-0-api-avvikelser-fran-blogginlagget-och-tva-gotchas]] — ny sida från inboxen
- [archive] 2026-10-23-pi-durable-1-4-0-api-avvikelser-fran-blogginlagget-och-tva-gotchas.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/pi-durable-fungerar-med-minimax-coding-plan-anthropic-messages-api]] — ny sida från inboxen
- [archive] 2026-10-23-pi-durable-fungerar-med-minimax-coding-plan-anthropic-messages-api.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-23] create | Ingest
- [create] [[notes/beslut-pi-durable-satts-pa-hyllan-for-tillfallet]] — ny sida från inboxen
- [archive] 2026-10-23-beslut-pi-durable-satts-pa-hyllan-for-tillfallet.md
- [merge?] concepts/FluxOperator + concepts/flux-operator (delade taggar: flux, kubernetes)
- [merge?] concepts/GitOps + concepts/gitops (delade taggar: gitops, kubernetes, flux)
- [merge?] concepts/kyverno-cel + kyverno-cel (delade taggar: kyverno, policy, cel, kubernetes, security)
- [merge?] entities/Flux GitOps + entities/flux (delade taggar: flux, gitops, kubernetes)
- [merge?] entities/Talos Linux + entities/talos (delade taggar: talos, kubernetes)

## [2026-10-24] create | Ingest
- [create] [[notes/md-viewer-tauri-installerades-men-avinstallerades-igen]] — ny sida från inboxen
- [create] [[notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc]] — ny sida från inboxen
- [archive] 2026-10-24-md-viewer-tauri-installerades-men-avinstallerades-igen.md
- [archive] 2026-10-24-pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc.md

## [2026-10-24] migrate | Karpathys LLM-wiki-modell
- [migrate] Sidnamn normaliserade till kebab-case; dubbletter slåt samma (GitOps, FluxOperator, Flux GitOps, Talos Linux)
- [migrate] Alla sidor fick `type`, `summary`, `created`, `updated` och `title` i frontmatter
- [migrate] Korsreferenser byggdes mellan relaterade sidor; inga ödelappar kvar
- [create] [[entities/pxe-server]] — hubbsida för PXE-kedjan
- [create] [[concepts/oidc]] — OIDC-protokollet mot Kubernetes API-servern
- [create] [[concepts/zitadel]] — klustrets identity provider
- [create] [[summaries/llm-wiki-karpathy]] — syntes av Karpathys gist
- [source] [[raw/2026-10-24-llm-wiki-karpathys-monster-for-llm-underhallna-kunskapsbaser]] — råkälla
- [update] [[concepts/authentik]] — markerad `deprecated`, ersatt av Zitadel
- [update] [[entities/kubernetes]] — ArgoCD/Authentik/Vault borttaget, Zitadel och Flux-appar tillagda
- [update] [[concepts/gitops]] — ArgoCD borttaget, allt hanteras av Flux
- [migrate] log.md normaliserades till det grep-barra formatet `## [YYYY-MM-DD] op | Titel`
- [update] `wiki/AGENTS.md` skrevs om till Karpathys tre lager: raw, wiki, schema

## [2026-10-24] ingest | create notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning
- [create] [[notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning]] — ny sida från inboxen
- [archive] 2026-10-24-usb-disken-acer-wd-320gb-status-och-longhorn-bedomning.md


## [2026-10-24] link | notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning -> concepts/longhorn


## [2026-10-24] link | notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning -> entities/talos-linux


## [2026-10-24] link | concepts/longhorn -> notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning


## [2026-10-24] link | entities/talos-linux -> notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning


## [2026-10-24] ingest | create notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin
- [create] [[notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin]] — ny sida från inboxen
- [archive] 2026-10-24-kureringspasset-hittade-foraldrad-kunskap-i-wikin.md


## [2026-10-24] link | summaries/llm-wiki-karpathy -> notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin


## [2026-10-24] link | notes/kunskaps-wiki -> notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin


## [2026-10-24] ingest | create notes/rock-4c-4se-talos-kan-aldrig-installeras-till-disk-u-boot-saknar-efi-setvariablert
- [create] [[notes/rock-4c-4se-talos-kan-aldrig-installeras-till-disk-u-boot-saknar-efi-setvariablert]] — ny sida från inboxen
- [archive] 2026-10-24-rock-4c-4se-talos-kan-aldrig-installeras-till-disk-u-boot-saknar-efi-setvariablert.md


## [2026-10-24] ingest | create notes/rock-4c-4se-utan-batteribackup-rtc-star-pa-2017-tls-mot-factory-talos-dev-misslyckas
- [create] [[notes/rock-4c-4se-utan-batteribackup-rtc-star-pa-2017-tls-mot-factory-talos-dev-misslyckas]] — ny sida från inboxen
- [create] [[notes/talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod]] — ny sida från inboxen
- [create] [[notes/talos-v1-14-utan-unattendedinstallconfig-oinstallerad-metal-nod-bootar-om-i-en-loop]] — ny sida från inboxen
- [archive] 2026-10-24-rock-4c-4se-utan-batteribackup-rtc-star-pa-2017-tls-mot-factory-talos-dev-misslyckas.md
- [archive] 2026-10-24-talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod.md
- [archive] 2026-10-24-talos-v1-14-utan-unattendedinstallconfig-oinstallerad-metal-nod-bootar-om-i-en-loop.md
## [2026-10-24] ingest | create notes/talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod
## [2026-10-24] ingest | create notes/talos-v1-14-utan-unattendedinstallconfig-oinstallerad-metal-nod-bootar-om-i-en-loop


## [2026-10-24] ingest | create notes/talos-hostname-auto-i-en-generisk-config-byter-nodens-namn-och-kastar-ut-den-ur-klustret
- [create] [[notes/talos-hostname-auto-i-en-generisk-config-byter-nodens-namn-och-kastar-ut-den-ur-klustret]] — ny sida från inboxen
- [archive] 2026-10-24-talos-hostname-auto-i-en-generisk-config-byter-nodens-namn-och-kastar-ut-den-ur-klustret.md


## [2026-10-25] ingest | create notes/popcorn-time-installerad-som-appimage-i-local
- [create] [[notes/popcorn-time-installerad-som-appimage-i-local]] — ny sida från inboxen
- [create] [[notes/popcorn-time-remote-movies-api-failed-to-respond-orsak-och-fix]] — ny sida från inboxen
- [create] [[notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection]] — ny sida från inboxen
- [create] [[notes/qbittorrent-ar-standardhanterare-for-torrent-och-magnet-lankar]] — ny sida från inboxen
- [create] [[notes/qbittorrent-installerad-pa-workstation-omarchy-arch]] — ny sida från inboxen
- [create] [[notes/recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten]] — ny sida från inboxen
- [archive] 2026-10-25-popcorn-time-installerad-som-appimage-i-local.md
- [archive] 2026-10-25-popcorn-time-remote-movies-api-failed-to-respond-orsak-och-fix.md
- [archive] 2026-10-25-popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection.md
- [archive] 2026-10-25-qbittorrent-ar-standardhanterare-for-torrent-och-magnet-lankar.md
- [archive] 2026-10-25-qbittorrent-installerad-pa-workstation-omarchy-arch.md
- [archive] 2026-10-25-recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten.md
## [2026-10-25] ingest | create notes/popcorn-time-remote-movies-api-failed-to-respond-orsak-och-fix
## [2026-10-25] ingest | create notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection
## [2026-10-25] ingest | create notes/qbittorrent-ar-standardhanterare-for-torrent-och-magnet-lankar
## [2026-10-25] ingest | create notes/qbittorrent-installerad-pa-workstation-omarchy-arch
## [2026-10-25] ingest | create notes/recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten


## [2026-10-25] ingest | create notes/talos-etcd-en-gammal-ephemeral-kopia-pa-controlplanet-ger-en-trasig-raft-log
- [create] [[notes/talos-etcd-en-gammal-ephemeral-kopia-pa-controlplanet-ger-en-trasig-raft-log]] — ny sida från inboxen
- [archive] 2026-10-25-talos-etcd-en-gammal-ephemeral-kopia-pa-controlplanet-ger-en-trasig-raft-log.md


## [2026-10-25] ingest | create notes/talos-noden-kan-hamna-i-maintenance-mode-om-state-volymen-hittas-efter-config-fasen
- [create] [[notes/talos-noden-kan-hamna-i-maintenance-mode-om-state-volymen-hittas-efter-config-fasen]] — ny sida från inboxen
- [archive] 2026-10-25-talos-noden-kan-hamna-i-maintenance-mode-om-state-volymen-hittas-efter-config-fasen.md


## [2026-10-26] ingest | create notes/rescue-wipe-pa-rock-maste-tomma-hela-systemdisken-wipe-system-esp-filen-ar-det-som-fangar-booten
- [create] [[notes/rescue-wipe-pa-rock-maste-tomma-hela-systemdisken-wipe-system-esp-filen-ar-det-som-fangar-booten]] — ny sida från inboxen
- [create] [[notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen]] — ny sida från inboxen
- [archive] 2026-10-26-rescue-wipe-pa-rock-maste-tomma-hela-systemdisken-wipe-system-esp-filen-ar-det-som-fangar-booten.md
- [archive] 2026-10-26-twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen.md
## [2026-10-26] ingest | create notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen


## [2026-10-27] ingest | create notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken
- [create] [[notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken]] — ny sida från inboxen
- [archive] 2026-10-27-worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken.md


## [2026-10-27] ingest | create notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret
- [create] [[notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret]] — ny sida från inboxen
- [create] [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]] — ny sida från inboxen
- [archive] 2026-10-27-it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret.md
- [archive] 2026-10-27-rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista.md
## [2026-10-27] ingest | create notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista


## [2026-10-27] link | notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista -> notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret


## [2026-10-27] link | notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret -> notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista


## [2026-10-27] ingest | create notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen
- [create] [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]] — ny sida från inboxen
- [archive] 2026-10-27-masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen.md


## [2026-10-27] link | notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen -> notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista


## [2026-10-27] link | notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista -> notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen


## [2026-10-28] ingest | create notes/mailmaster-api-v1-gav-failed-to-fetch-go-servern-panickade-och-saknade-cors
- [create] [[notes/mailmaster-api-v1-gav-failed-to-fetch-go-servern-panickade-och-saknade-cors]] — ny sida från inboxen
- [archive] 2026-10-28-mailmaster-api-v1-gav-failed-to-fetch-go-servern-panickade-och-saknade-cors.md


## [2026-10-29] ingest | create notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan
- [create] [[notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan]] — ny sida från inboxen
- [archive] 2026-10-29-mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan.md


## [2026-10-29] ingest | create notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent
- [create] [[notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent]] — ny sida från inboxen
- [create] [[notes/kameo-uppgifter-synkas-till-fil-sa-docker-slipper-op-cli]] — ny sida från inboxen
- [create] [[notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo]] — ny sida från inboxen
- [create] [[notes/mailmaster-gammal-container-build-korde-process-och-arkiverade-mejl]] — ny sida från inboxen
- [create] [[notes/mailmaster-granskning-av-ai-atgarder-fore-utforande]] — ny sida från inboxen
- [create] [[notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code]] — ny sida från inboxen
- [archive] 2026-10-29-imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent.md
- [archive] 2026-10-29-kameo-uppgifter-synkas-till-fil-sa-docker-slipper-op-cli.md
- [archive] 2026-10-29-kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo.md
- [archive] 2026-10-29-mailmaster-gammal-container-build-korde-process-och-arkiverade-mejl.md
- [archive] 2026-10-29-mailmaster-granskning-av-ai-atgarder-fore-utforande.md
- [archive] 2026-10-29-mailmaster-kontodropdown-outlook-inloggning-via-device-code.md
## [2026-10-29] ingest | create notes/kameo-uppgifter-synkas-till-fil-sa-docker-slipper-op-cli
## [2026-10-29] ingest | create notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo
## [2026-10-29] ingest | create notes/mailmaster-gammal-container-build-korde-process-och-arkiverade-mejl
## [2026-10-29] ingest | create notes/mailmaster-granskning-av-ai-atgarder-fore-utforande
## [2026-10-29] ingest | create notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code


## [2026-10-29] link | notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code -> notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent


## [2026-10-29] link | notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent -> notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code


## [2026-10-29] ingest | create notes/cachade-imap-sessioner-kraver-hantering-av-doda-anslutningar
- [create] [[notes/cachade-imap-sessioner-kraver-hantering-av-doda-anslutningar]] — ny sida från inboxen
- [archive] 2026-10-29-cachade-imap-sessioner-kraver-hantering-av-doda-anslutningar.md


## [2026-10-29] link | notes/cachade-imap-sessioner-kraver-hantering-av-doda-anslutningar -> notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code


## [2026-10-29] link | notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code -> notes/cachade-imap-sessioner-kraver-hantering-av-doda-anslutningar


## [2026-11-02] ingest | create notes/rust-utvecklingsmiljo-pa-workstationen
- [create] [[notes/rust-utvecklingsmiljo-pa-workstationen]] — ny sida från inboxen
- [archive] 2026-11-02-rust-utvecklingsmiljo-pa-workstationen.md


## [2026-11-02] ingest | create notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer
- [create] [[notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer]] — ny sida från inboxen
- [archive] 2026-11-02-rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer.md


## [2026-11-02] ingest | create notes/rust-analyzer-tysta-health-status-notisen-istallet-for-att-stanga-av-standalone
- [create] [[notes/rust-analyzer-tysta-health-status-notisen-istallet-for-att-stanga-av-standalone]] — ny sida från inboxen
- [archive] 2026-11-02-rust-analyzer-tysta-health-status-notisen-istallet-for-att-stanga-av-standalone.md


## [2026-11-02] page | concepts/poolad-eller-direkt-anslutning-mot-neon


## [2026-11-02] link | concepts/poolad-eller-direkt-anslutning-mot-neon -> notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen


## [2026-11-02] ingest | create notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring
- [create] [[notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring]] — ny sida från inboxen
- [archive] 2026-11-02-pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring.md


## [2026-11-02] link | notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring -> concepts/poolad-eller-direkt-anslutning-mot-neon


## [2026-11-02] link | notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen -> notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring


## [2026-11-02] page | concepts/neon-cli-var-neon-och-neon-ts-maste-ligga


## [2026-11-02] link | notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring -> concepts/neon-cli-var-neon-och-neon-ts-maste-ligga


## [2026-11-02] ingest | create notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon
- [create] [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]] — ny sida från inboxen
- [archive] 2026-11-02-nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon.md


## [2026-11-02] link | notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon -> notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen


## [2026-11-02] link | notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen -> notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon


## [2026-11-02] ingest | create notes/sa-dampas-bada-rust-analyzer-popupen-no-project-root-found
- [create] [[notes/sa-dampas-bada-rust-analyzer-popupen-no-project-root-found]] — ny sida från inboxen
- [archive] 2026-11-02-sa-dampas-bada-rust-analyzer-popupen-no-project-root-found.md


## [2026-11-02] ingest | create notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet
- [create] [[notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet]] — ny sida från inboxen
- [archive] 2026-11-02-root-dir-funktionen-stoppar-no-project-root-found-meddelandet.md


## [2026-11-02] ingest | create notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok
- [create] [[notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok]] — ny sida från inboxen
- [create] [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]] — ny sida från inboxen
- [archive] 2026-11-02-neuroquant-kontot-last-av-for-manga-inloggningsforsok.md
- [archive] 2026-11-02-pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres.md
## [2026-11-02] ingest | create notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres


## [2026-11-02] ingest | create notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02
- [create] [[notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02]] — ny sida från inboxen
- [archive] 2026-11-02-skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02.md


## [2026-11-02] link | notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02 -> notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen


## [2026-11-02] ingest | create notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer
- [create] [[notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]] — ny sida från inboxen
- [archive] 2026-11-02-skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer.md


## [2026-11-02] ingest | create notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet
- [create] [[notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet]] — ny sida från inboxen
- [archive] 2026-11-02-skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet.md


## [2026-11-02] link | notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer -> notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet


## [2026-11-02] link | notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet -> notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer


## [2026-11-03] ingest | update notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres
- [update] [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]] — tillägg från inboxen
- [archive] 2026-11-03-pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres.md


## [2026-11-03] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring


## [2026-11-03] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon


## [2026-11-03] link | notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring -> notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres


## [2026-11-03] link | notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon -> notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres


## [2026-11-03] ingest | create notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen
- [create] [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]] — ny sida från inboxen
- [archive] 2026-11-03-skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen.md


## [2026-11-03] link | notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer -> notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen


## [2026-11-03] link | notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen -> notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer


## [2026-11-03] ingest | create notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund
- [create] [[notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund]] — ny sida från inboxen
- [archive] 2026-11-03-erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund.md


## [2026-11-03] link | notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen -> notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund


## [2026-11-03] link | notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund -> notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen


## [2026-11-03] ingest | create notes/learn-kunskapssystem-implementerat-i-src-python
- [create] [[notes/learn-kunskapssystem-implementerat-i-src-python]] — ny sida från inboxen
- [archive] 2026-11-03-learn-kunskapssystem-implementerat-i-src-python.md


## [2026-11-03] page | notes/learn-kunskapssystem-implementerat-i-src-python


## [2026-11-03] ingest | create notes/learn-kunskapssystemet-portat-fran-python-till-go
- [create] [[notes/learn-kunskapssystemet-portat-fran-python-till-go]] — ny sida från inboxen
- [archive] 2026-11-03-learn-kunskapssystemet-portat-fran-python-till-go.md


## [2026-11-03] link | notes/learn-kunskapssystem-implementerat-i-src-python -> notes/learn-kunskapssystemet-portat-fran-python-till-go


## [2026-11-03] link | notes/learn-kunskapssystemet-portat-fran-python-till-go -> notes/learn-kunskapssystem-implementerat-i-src-python


## [2026-11-03] ingest | create notes/ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor
- [create] [[notes/ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor]] — ny sida från inboxen
- [create] [[notes/skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open]] — ny sida från inboxen
- [archive] 2026-11-03-ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor.md
- [archive] 2026-11-03-skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open.md
## [2026-11-03] ingest | create notes/skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open


## [2026-11-03] link | notes/skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open -> notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen


## [2026-11-03] link | notes/ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor -> notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen


## [2026-11-03] ingest | create notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering
- [create] [[notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering]] — ny sida från inboxen
- [archive] 2026-11-03-neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering.md


## [2026-11-03] link | notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering -> notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok


## [2026-11-03] link | notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok -> notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering


## [2026-11-03] ingest | create notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken
- [create] [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]] — ny sida från inboxen
- [archive] 2026-11-03-pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken.md


## [2026-11-03] link | notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund -> notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken


## [2026-11-03] link | notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen -> notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken


## [2026-11-03] link | notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista -> notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken


## [2026-11-03] link | notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken -> notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen


## [2026-11-03] ingest | create notes/skogsduvans-ha-utsatts-for-automatiserade-skannningar-fran-internet
- [create] [[notes/skogsduvans-ha-utsatts-for-automatiserade-skannningar-fran-internet]] — ny sida från inboxen
- [archive] 2026-11-03-skogsduvans-ha-utsatts-for-automatiserade-skannningar-fran-internet.md


## [2026-11-03] link | notes/skogsduvans-ha-utsatts-for-automatiserade-skannningar-fran-internet -> notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer


## [2026-11-03] ingest | create notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering
- [create] [[notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]] — ny sida från inboxen
- [archive] 2026-11-03-neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering.md


## [2026-11-03] link | notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering -> notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering


## [2026-11-03] link | notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering -> notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok


## [2026-11-03] link | notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering -> notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering


## [2026-11-03] link | notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok -> notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering


## [2026-11-03] ingest | create notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo
- [create] [[notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo]] — ny sida från inboxen
- [archive] 2026-11-03-cookies-json-var-versionsstyrd-i-offentligt-neuro-repo.md


## [2026-11-03] link | notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo -> notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering


## [2026-11-03] link | notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering -> notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo


## [2026-11-03] ingest | create notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online
- [create] [[notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online]] — ny sida från inboxen
- [archive] 2026-11-03-skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online.md


## [2026-11-03] link | notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online -> notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen


## [2026-11-03] link | notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen -> notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online


## [2026-11-03] ingest | create notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas
- [create] [[notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas]] — ny sida från inboxen
- [archive] 2026-11-03-learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas.md


## [2026-11-03] link | notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas -> notes/learn-kunskapssystemet-portat-fran-python-till-go


## [2026-11-03] ingest | create notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen
- [create] [[notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen]] — ny sida från inboxen
- [create] [[notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer]] — ny sida från inboxen
- [archive] 2026-11-03-pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen.md
- [archive] 2026-11-03-utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer.md
## [2026-11-03] ingest | create notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer


## [2026-11-03] link | notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen -> notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen


## [2026-11-03] link | notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer -> notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista


## [2026-11-03] link | notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer -> notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken


## [2026-11-03] link | notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista -> notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer


## [2026-11-03] link | notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen -> notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen


## [2026-11-03] link | notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken -> notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer


## [2026-11-03] link | notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer -> notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken


## [2026-11-03] link | notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen -> notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken


## [2026-11-03] ingest | create notes/learn-designsystem-a3-folio-i-nuxt
- [create] [[notes/learn-designsystem-a3-folio-i-nuxt]] — ny sida från inboxen
- [archive] 2026-11-03-learn-designsystem-a3-folio-i-nuxt.md


## [2026-11-03] ingest | create notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning
- [create] [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]] — ny sida från inboxen
- [archive] 2026-11-03-pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning.md


## [2026-11-03] link | notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning -> notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen


## [2026-11-03] link | notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen -> notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning


## [2026-11-03] link | notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen -> notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning


## [2026-11-03] ingest | create notes/learn-calibre-synk-och-automatisk-import-via-systemd
- [create] [[notes/learn-calibre-synk-och-automatisk-import-via-systemd]] — ny sida från inboxen
- [archive] 2026-11-03-learn-calibre-synk-och-automatisk-import-via-systemd.md


## [2026-11-03] ingest | create notes/median-i-stallet-for-snitt-for-wpm
- [create] [[notes/median-i-stallet-for-snitt-for-wpm]] — ny sida från inboxen
- [create] [[notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin]] — ny sida från inboxen
- [archive] 2026-11-03-median-i-stallet-for-snitt-for-wpm.md
- [archive] 2026-11-03-stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin.md
## [2026-11-03] ingest | create notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin


## [2026-11-03] link | notes/median-i-stallet-for-snitt-for-wpm -> notes/utvecklingsflode


## [2026-11-03] link | notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin -> notes/kunskaps-wiki


## [2026-11-03] ingest | create notes/learn-korsanalys-av-varje-bok-mot-den-samlade-kunskapsbasen
- [create] [[notes/learn-korsanalys-av-varje-bok-mot-den-samlade-kunskapsbasen]] — ny sida från inboxen
- [create] [[notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking]] — ny sida från inboxen
- [create] [[notes/repo-template-ny-rust-api-src-rust-med-axum-utoipa-scalar]] — ny sida från inboxen
- [create] [[notes/repo-template-rust-api-kopplad-till-kubernetes-manifests-tilt]] — ny sida från inboxen
- [create] [[notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork]] — ny sida från inboxen
- [create] [[notes/rust-repo-learn-to-code-with-rust-som-submodule]] — ny sida från inboxen
- [archive] 2026-11-03-learn-korsanalys-av-varje-bok-mot-den-samlade-kunskapsbasen.md
- [archive] 2026-11-03-pappa-uppdragsmatchning-delpoang-forklarande-ranking.md
- [archive] 2026-11-03-repo-template-ny-rust-api-src-rust-med-axum-utoipa-scalar.md
- [archive] 2026-11-03-repo-template-rust-api-kopplad-till-kubernetes-manifests-tilt.md
- [archive] 2026-11-03-rust-learning-repot-lades-in-som-submodule-mot-egen-fork.md
- [archive] 2026-11-03-rust-repo-learn-to-code-with-rust-som-submodule.md
## [2026-11-03] ingest | create notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking
## [2026-11-03] ingest | create notes/repo-template-ny-rust-api-src-rust-med-axum-utoipa-scalar
## [2026-11-03] ingest | create notes/repo-template-rust-api-kopplad-till-kubernetes-manifests-tilt
## [2026-11-03] ingest | create notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork
## [2026-11-03] ingest | create notes/rust-repo-learn-to-code-with-rust-som-submodule


## [2026-11-03] link | notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning -> notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking


## [2026-11-03] ingest | create notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet
- [create] [[notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet]] — ny sida från inboxen
- [create] [[notes/repo-template-ci-gitops-flodet-fixat-bygger-3-images-ghcr-tagging-pr-overlays-per-pr]] — ny sida från inboxen
- [archive] 2026-11-03-pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet.md
- [archive] 2026-11-03-repo-template-ci-gitops-flodet-fixat-bygger-3-images-ghcr-tagging-pr-overlays-per-pr.md
## [2026-11-03] ingest | create notes/repo-template-ci-gitops-flodet-fixat-bygger-3-images-ghcr-tagging-pr-overlays-per-pr


## [2026-11-03] link | notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet -> entities/pxe-server


## [2026-11-03] ingest | create notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista
- [create] [[notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista]] — ny sida från inboxen
- [archive] 2026-11-03-pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista.md


## [2026-11-03] link | notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista -> notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking


## [2026-11-03] link | notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista -> notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning


## [2026-11-03] ingest | create notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo
- [create] [[notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo]] — ny sida från inboxen
- [archive] 2026-11-03-pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo.md


## [2026-11-03] link | notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo -> notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista


## [2026-11-03] link | notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo -> notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning


## [2026-11-03] ingest | create notes/pappa-hemlighet-i-build-context-env-i-docker-imagen
- [create] [[notes/pappa-hemlighet-i-build-context-env-i-docker-imagen]] — ny sida från inboxen
- [archive] 2026-11-03-pappa-hemlighet-i-build-context-env-i-docker-imagen.md


## [2026-11-03] link | notes/pappa-hemlighet-i-build-context-env-i-docker-imagen -> notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo


## [2026-11-03] ingest | create notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet
- [create] [[notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet]] — ny sida från inboxen
- [archive] 2026-11-03-pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet.md


## [2026-11-03] link | notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet -> entities/pxe-server


## [2026-11-03] link | notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet -> notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet


## [2026-11-03] source | 2026-11-03-learn-kunskapsbasen-kan-lagras-i-postgresql.md


## [2026-11-03] ingest | create notes/learn-kunskapsbasen-kan-lagras-i-postgresql
- [create] [[notes/learn-kunskapsbasen-kan-lagras-i-postgresql]] — ny sida från inboxen
- [archive] 2026-11-03-learn-kunskapsbasen-kan-lagras-i-postgresql.md


## [2026-11-03] link | notes/learn-kunskapsbasen-kan-lagras-i-postgresql -> summaries/2026-11-03-learn-kunskapsbasen-kan-lagras-i-postgresql


## [2026-11-03] link | notes/learn-korsanalys-av-varje-bok-mot-den-samlade-kunskapsbasen -> notes/learn-kunskapsbasen-kan-lagras-i-postgresql


## [2026-11-03] link | notes/learn-calibre-synk-och-automatisk-import-via-systemd -> notes/learn-kunskapsbasen-kan-lagras-i-postgresql


## [2026-11-06] ingest | create notes/query-parametrar-i-actor-smartbook-endpointen-url-avkodas-till-mellanslag
- [create] [[notes/query-parametrar-i-actor-smartbook-endpointen-url-avkodas-till-mellanslag]] — ny sida från inboxen
- [create] [[notes/stk-dev-miljon-kor-mot-neon-istallet-for-cnpg]] — ny sida från inboxen
- [create] [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]] — ny sida från inboxen
- [archive] 2026-11-06-query-parametrar-i-actor-smartbook-endpointen-url-avkodas-till-mellanslag.md
- [archive] 2026-11-06-stk-dev-miljon-kor-mot-neon-istallet-for-cnpg.md
- [archive] 2026-11-06-synk-hallbokningar-kommunens-actor-smartbook-matchi.md
## [2026-11-06] ingest | create notes/stk-dev-miljon-kor-mot-neon-istallet-for-cnpg
## [2026-11-06] ingest | create notes/synk-hallbokningar-kommunens-actor-smartbook-matchi


## [2026-11-06] ingest | create notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas
- [create] [[notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas]] — ny sida från inboxen
- [create] [[notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass]] — ny sida från inboxen
- [create] [[notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts]] — ny sida från inboxen
- [archive] 2026-11-06-matchi-bokning-slot-id-andras-nar-bokningen-skapas.md
- [archive] 2026-11-06-matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass.md
- [archive] 2026-11-06-stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts.md
## [2026-11-06] ingest | create notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass
## [2026-11-06] ingest | create notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts


## [2026-11-06] link | notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass -> notes/synk-hallbokningar-kommunens-actor-smartbook-matchi


## [2026-11-06] link | notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas -> notes/synk-hallbokningar-kommunens-actor-smartbook-matchi


## [2026-11-06] link | notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts -> notes/synk-hallbokningar-kommunens-actor-smartbook-matchi


## [2026-11-06] link | notes/synk-hallbokningar-kommunens-actor-smartbook-matchi -> notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass


## [2026-11-06] link | notes/synk-hallbokningar-kommunens-actor-smartbook-matchi -> notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas


## [2026-11-06] link | notes/synk-hallbokningar-kommunens-actor-smartbook-matchi -> notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts


## [2026-11-06] ingest | create notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force
- [create] [[notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force]] — ny sida från inboxen
- [archive] 2026-11-06-krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force.md


## [2026-11-06] link | notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force -> notes/synk-hallbokningar-kommunens-actor-smartbook-matchi


## [2026-11-06] link | notes/synk-hallbokningar-kommunens-actor-smartbook-matchi -> notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force


## [2026-11-06] ingest | create notes/nuxt-cli-via-bun-run-dev-gar-sonder-i-stk
- [create] [[notes/nuxt-cli-via-bun-run-dev-gar-sonder-i-stk]] — ny sida från inboxen
- [create] [[notes/transienta-502-fra-go-api-t-nar-poddar-startar-om-i-dev-klustret]] — ny sida från inboxen
- [archive] 2026-11-06-nuxt-cli-via-bun-run-dev-gar-sonder-i-stk.md
- [archive] 2026-11-06-transienta-502-fra-go-api-t-nar-poddar-startar-om-i-dev-klustret.md
## [2026-11-06] ingest | create notes/transienta-502-fra-go-api-t-nar-poddar-startar-om-i-dev-klustret


## [2026-11-06] ingest | create notes/gron-markering-av-tider-som-ar-bokade-i-matchi-skrivlasande-jamforelse
- [create] [[notes/gron-markering-av-tider-som-ar-bokade-i-matchi-skrivlasande-jamforelse]] — ny sida från inboxen
- [archive] 2026-11-06-gron-markering-av-tider-som-ar-bokade-i-matchi-skrivlasande-jamforelse.md


## [2026-11-06] ingest | create notes/matchi-inloggningen-ar-flaky-klienten-forsoker-om-och-dev-klustret-behover-ombyggnad
- [create] [[notes/matchi-inloggningen-ar-flaky-klienten-forsoker-om-och-dev-klustret-behover-ombyggnad]] — ny sida från inboxen
- [archive] 2026-11-06-matchi-inloggningen-ar-flaky-klienten-forsoker-om-och-dev-klustret-behover-ombyggnad.md


## [2026-11-06] ingest | create notes/ci-saknade-inloggning-mot-ghcr-alla-image-pushar-misslyckades
- [create] [[notes/ci-saknade-inloggning-mot-ghcr-alla-image-pushar-misslyckades]] — ny sida från inboxen
- [create] [[notes/lunchmenyn-delar-en-modell-mellan-byggskript-och-webblasare]] — ny sida från inboxen
- [archive] 2026-11-06-ci-saknade-inloggning-mot-ghcr-alla-image-pushar-misslyckades.md
- [archive] 2026-11-06-lunchmenyn-delar-en-modell-mellan-byggskript-och-webblasare.md
## [2026-11-06] ingest | create notes/lunchmenyn-delar-en-modell-mellan-byggskript-och-webblasare


## [2026-11-06] ingest | create notes/boka-enskild-kommunbokning-i-matchi-fran-bookings
- [create] [[notes/boka-enskild-kommunbokning-i-matchi-fran-bookings]] — ny sida från inboxen
- [archive] 2026-11-06-boka-enskild-kommunbokning-i-matchi-fran-bookings.md


## [2026-11-10] ingest | create notes/dev-inloggning-i-plan-session-token-db-rattigheter-och-oauth-url-er
- [create] [[notes/dev-inloggning-i-plan-session-token-db-rattigheter-och-oauth-url-er]] — ny sida från inboxen
- [archive] 2026-11-10-dev-inloggning-i-plan-session-token-db-rattigheter-och-oauth-url-er.md


## [2026-11-10] ingest | create notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt
- [create] [[notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt]] — ny sida från inboxen
- [archive] 2026-11-10-authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt.md


## [2026-11-10] link | notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt -> notes/dev-inloggning-i-plan-session-token-db-rattigheter-och-oauth-url-er


## [2026-11-10] link | notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt -> concepts/zitadel


## [2026-11-10] ingest | create notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg
- [create] [[notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg]] — ny sida från inboxen
- [archive] 2026-11-10-plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg.md


## [2026-11-10] link | notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg -> notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt


## [2026-11-13] ingest | create notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11
- [create] [[notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11]] — ny sida från inboxen
- [create] [[notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe]] — ny sida från inboxen
- [create] [[notes/simonbrundin-gmail-com-inkorgen-kan-lasas-via-imap-begbot-losenordet]] — ny sida från inboxen
- [create] [[notes/twingate-ressurserna-forsvann-pa-arbetsstationen-2026-11-04-hemnatet-10-1-1-1-och-datacenter-10-10-10-0-24-oppnarbara]] — ny sida från inboxen
- [create] [[notes/u7-lite-94-2a-6f-7a-4b-33-adopterar-inte-saknar-management-ip-och-nattverkssegment-pa-switchen]] — ny sida från inboxen
- [create] [[notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device]] — ny sida från inboxen
- [archive] 2026-11-13-cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11.md
- [archive] 2026-11-13-picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe.md
- [archive] 2026-11-13-simonbrundin-gmail-com-inkorgen-kan-lasas-via-imap-begbot-losenordet.md
- [archive] 2026-11-13-twingate-ressurserna-forsvann-pa-arbetsstationen-2026-11-04-hemnatet-10-1-1-1-och-datacenter-10-10-10-0-24-oppnarbara.md
- [archive] 2026-11-13-u7-lite-94-2a-6f-7a-4b-33-adopterar-inte-saknar-management-ip-och-nattverkssegment-pa-switchen.md
- [archive] 2026-11-13-workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device.md
## [2026-11-13] ingest | create notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe
## [2026-11-13] ingest | create notes/simonbrundin-gmail-com-inkorgen-kan-lasas-via-imap-begbot-losenordet
## [2026-11-13] ingest | create notes/twingate-ressurserna-forsvann-pa-arbetsstationen-2026-11-04-hemnatet-10-1-1-1-och-datacenter-10-10-10-0-24-oppnarbara
## [2026-11-13] ingest | create notes/u7-lite-94-2a-6f-7a-4b-33-adopterar-inte-saknar-management-ip-och-nattverkssegment-pa-switchen
## [2026-11-13] ingest | create notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device


## [2026-11-13] link | notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device -> entities/pxe-server


## [2026-11-13] link | notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device -> notes/worker-1-startar-om-med-strom-men-skickar-noll-paket-fysisk-omstart-bevisar-samma-fel


## [2026-11-13] link | notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11 -> notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device


## [2026-11-13] source | 2026-11-13-16x1-hdmi-kvm-rs232-lan-communication-protocol.md


## [2026-11-13] ingest | create notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm
- [create] [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]] — ny sida från inboxen
- [archive] 2026-11-13-kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm.md


## [2026-11-13] page | summaries/16x1-hdmi-kvm-rs232-lan-protokoll


## [2026-11-13] page | entities/hdmi-kvm-switch-16-port-labbet


## [2026-11-13] link | summaries/16x1-hdmi-kvm-rs232-lan-protokoll -> entities/hdmi-kvm-switch-16-port-labbet


## [2026-11-13] link | summaries/16x1-hdmi-kvm-rs232-lan-protokoll -> notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm


## [2026-11-13] link | entities/hdmi-kvm-switch-16-port-labbet -> summaries/16x1-hdmi-kvm-rs232-lan-protokoll


## [2026-11-13] link | entities/hdmi-kvm-switch-16-port-labbet -> notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm


## [2026-11-13] link | notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm -> entities/hdmi-kvm-switch-16-port-labbet


## [2026-11-13] link | notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm -> summaries/16x1-hdmi-kvm-rs232-lan-protokoll


## [2026-11-13] link | notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm -> notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe


## [2026-11-15] ingest | create notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink
- [create] [[notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink]] — ny sida från inboxen
- [archive] 2026-11-15-unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink.md


## [2026-11-16] ingest | create notes/lightcraft-photocraft-installerade-pa-omarchy-workstation-utan-sudo
- [create] [[notes/lightcraft-photocraft-installerade-pa-omarchy-workstation-utan-sudo]] — ny sida från inboxen
- [create] [[notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa]] — ny sida från inboxen
- [create] [[notes/photocraft-agentstyrning-via-control-protokoll-control-7878]] — ny sida från inboxen
- [create] [[notes/photocraft-kamerans-look-varning-vid-raw-oppning-och-hur-den-hanteras]] — ny sida från inboxen
- [archive] 2026-11-16-lightcraft-photocraft-installerade-pa-omarchy-workstation-utan-sudo.md
- [archive] 2026-11-16-mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa.md
- [archive] 2026-11-16-photocraft-agentstyrning-via-control-protokoll-control-7878.md
- [archive] 2026-11-16-photocraft-kamerans-look-varning-vid-raw-oppning-och-hur-den-hanteras.md
## [2026-11-16] ingest | create notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa
## [2026-11-16] ingest | create notes/photocraft-agentstyrning-via-control-protokoll-control-7878
## [2026-11-16] ingest | create notes/photocraft-kamerans-look-varning-vid-raw-oppning-och-hur-den-hanteras


## [2026-11-16] ingest | create notes/plan-dev-kor-mot-neon-istallet-for-lokal-postgres
- [create] [[notes/plan-dev-kor-mot-neon-istallet-for-lokal-postgres]] — ny sida från inboxen
- [archive] 2026-11-16-plan-dev-kor-mot-neon-istallet-for-lokal-postgres.md


## [2026-11-16] ingest | create notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades
- [create] [[notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades]] — ny sida från inboxen
- [create] [[notes/plan-cachad-spa-skal-gjorde-att-webblasaren-korde-gammalt-js-bundle-med-ikon-provider-none]] — ny sida från inboxen
- [create] [[notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan]] — ny sida från inboxen
- [archive] 2026-11-16-ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades.md
- [archive] 2026-11-16-plan-cachad-spa-skal-gjorde-att-webblasaren-korde-gammalt-js-bundle-med-ikon-provider-none.md
- [archive] 2026-11-16-sortering-av-undermal-med-piltangenterna-var-trasig-i-plan.md
## [2026-11-16] ingest | create notes/plan-cachad-spa-skal-gjorde-att-webblasaren-korde-gammalt-js-bundle-med-ikon-provider-none
## [2026-11-16] ingest | create notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan


## [2026-11-16] ingest | create notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter
- [create] [[notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter]] — ny sida från inboxen
- [create] [[notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema]] — ny sida från inboxen
- [archive] 2026-11-16-mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter.md
- [archive] 2026-11-16-mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema.md
## [2026-11-16] ingest | create notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema


## [2026-11-16] link | notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema -> notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa


## [2026-11-16] link | notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema -> notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter


## [2026-11-16] page | notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter


## [2026-11-16] link | notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter -> notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema


## [2026-11-16] page | notes/mailagentens-kontofel-gav-alltid-http-500-och-listan-foljde-inte-kontot


## [2026-11-16] link | notes/mailagentens-kontofel-gav-alltid-http-500-och-listan-foljde-inte-kontot -> notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter


## [2026-11-16] link | notes/mailagentens-kontofel-gav-alltid-http-500-och-listan-foljde-inte-kontot -> notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema


## [2026-11-16] link | notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter -> notes/mailagentens-kontofel-gav-alltid-http-500-och-listan-foljde-inte-kontot


## [2026-11-16] ingest | create notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet
- [create] [[notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet]] — ny sida från inboxen
- [create] [[notes/mailvyn-mail-omstegad-till-tailwind-tema]] — ny sida från inboxen
- [create] [[notes/tilt-filbevakning-kan-do-starta-om-tilt-up]] — ny sida från inboxen
- [archive] 2026-11-16-enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet.md
- [archive] 2026-11-16-mailvyn-mail-omstegad-till-tailwind-tema.md
- [archive] 2026-11-16-tilt-filbevakning-kan-do-starta-om-tilt-up.md
## [2026-11-16] ingest | create notes/mailvyn-mail-omstegad-till-tailwind-tema
## [2026-11-16] ingest | create notes/tilt-filbevakning-kan-do-starta-om-tilt-up


## [2026-11-16] ingest | create notes/nuxt-sidomeny-kandidater-som-startsida-i-pappa
- [create] [[notes/nuxt-sidomeny-kandidater-som-startsida-i-pappa]] — ny sida från inboxen
- [create] [[notes/pappa-regelendpointet-laser-och-sparar-riktig-rules-md-var-stub]] — ny sida från inboxen
- [archive] 2026-11-16-nuxt-sidomeny-kandidater-som-startsida-i-pappa.md
- [archive] 2026-11-16-pappa-regelendpointet-laser-och-sparar-riktig-rules-md-var-stub.md
## [2026-11-16] ingest | create notes/pappa-regelendpointet-laser-och-sparar-riktig-rules-md-var-stub


## [2026-11-16] ingest | create notes/kameo-sidan-flyttad-till-toppniva-och-delad-i-flikar
- [create] [[notes/kameo-sidan-flyttad-till-toppniva-och-delad-i-flikar]] — ny sida från inboxen
- [create] [[notes/sales-soulmate-se-appen-kan-inte-logga-in-oauth-clienten-saknas-i-soulmate-tenanten]] — ny sida från inboxen
- [archive] 2026-11-16-kameo-sidan-flyttad-till-toppniva-och-delad-i-flikar.md
- [archive] 2026-11-16-sales-soulmate-se-appen-kan-inte-logga-in-oauth-clienten-saknas-i-soulmate-tenanten.md
## [2026-11-16] ingest | create notes/sales-soulmate-se-appen-kan-inte-logga-in-oauth-clienten-saknas-i-soulmate-tenanten


## [2026-11-16] ingest | create notes/rekryteringspipeline-board-per-person-sourcing-som-attribut
- [create] [[notes/rekryteringspipeline-board-per-person-sourcing-som-attribut]] — ny sida från inboxen
- [archive] 2026-11-16-rekryteringspipeline-board-per-person-sourcing-som-attribut.md


## [2026-11-16] ingest | create notes/admin-inloggning-via-losenord-i-stallet-for-token-i-ui-t
- [create] [[notes/admin-inloggning-via-losenord-i-stallet-for-token-i-ui-t]] — ny sida från inboxen
- [create] [[notes/mailagent-bygget-brot-pa-ofardig-kameo-integration]] — ny sida från inboxen
- [archive] 2026-11-16-admin-inloggning-via-losenord-i-stallet-for-token-i-ui-t.md
- [archive] 2026-11-16-mailagent-bygget-brot-pa-ofardig-kameo-integration.md
## [2026-11-16] ingest | create notes/mailagent-bygget-brot-pa-ofardig-kameo-integration


## [2026-11-16] ingest | create notes/started-finished-kunde-inte-rensas-i-plan-null-ignorerades-i-go-api-t
- [create] [[notes/started-finished-kunde-inte-rensas-i-plan-null-ignorerades-i-go-api-t]] — ny sida från inboxen
- [archive] 2026-11-16-started-finished-kunde-inte-rensas-i-plan-null-ignorerades-i-go-api-t.md


## [2026-11-16] ingest | create notes/eviga-mal-i-plan-kan-inte-slutforas
- [create] [[notes/eviga-mal-i-plan-kan-inte-slutforas]] — ny sida från inboxen
- [archive] 2026-11-16-eviga-mal-i-plan-kan-inte-slutforas.md


## [2026-11-16] link | notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet -> notes/eviga-mal-i-plan-kan-inte-slutforas


## [2026-11-16] ingest | create notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet
- [create] [[notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet]] — ny sida från inboxen
- [create] [[notes/kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr]] — ny sida från inboxen
- [archive] 2026-11-16-kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet.md
- [archive] 2026-11-16-kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr.md
## [2026-11-16] ingest | create notes/kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr


## [2026-11-16] ingest | create notes/one-com-lagrar-mappar-i-nfd-normalisera-i-getfoldername
- [create] [[notes/one-com-lagrar-mappar-i-nfd-normalisera-i-getfoldername]] — ny sida från inboxen
- [archive] 2026-11-16-one-com-lagrar-mappar-i-nfd-normalisera-i-getfoldername.md


## [2026-11-16] ingest | create notes/goals-listar-bara-grundmal-scroll-vid-j-k
- [create] [[notes/goals-listar-bara-grundmal-scroll-vid-j-k]] — ny sida från inboxen
- [create] [[notes/sortering-av-grundmal-pa-goals-med-piltangenterna]] — ny sida från inboxen
- [archive] 2026-11-16-goals-listar-bara-grundmal-scroll-vid-j-k.md
- [archive] 2026-11-16-sortering-av-grundmal-pa-goals-med-piltangenterna.md
## [2026-11-16] ingest | create notes/sortering-av-grundmal-pa-goals-med-piltangenterna


## [2026-11-16] ingest | create notes/rantebetalningsmejl-hamnar-i-inbox-deleted-messages-och-arkivet-matchades-aldrig
- [create] [[notes/rantebetalningsmejl-hamnar-i-inbox-deleted-messages-och-arkivet-matchades-aldrig]] — ny sida från inboxen
- [archive] 2026-11-16-rantebetalningsmejl-hamnar-i-inbox-deleted-messages-och-arkivet-matchades-aldrig.md


## [2026-11-16] ingest | create notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t
- [create] [[notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t]] — ny sida från inboxen
- [create] [[notes/plan-appen-morkt-designsystem-med-semantiska-farg-token]] — ny sida från inboxen
- [archive] 2026-11-16-mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t.md
- [archive] 2026-11-16-plan-appen-morkt-designsystem-med-semantiska-farg-token.md
## [2026-11-16] ingest | create notes/plan-appen-morkt-designsystem-med-semantiska-farg-token


## [2026-11-16] link | notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t -> notes/eviga-mal-i-plan-kan-inte-slutforas


## [2026-11-16] link | notes/eviga-mal-i-plan-kan-inte-slutforas -> notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t


## [2026-11-16] link | notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t -> notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet


## [2026-11-17] ingest | create notes/automatisk-malikon-behovde-sokord-och-svenska-alias
- [create] [[notes/automatisk-malikon-behovde-sokord-och-svenska-alias]] — ny sida från inboxen
- [create] [[notes/kalenderns-zoom-styrs-av-vertikal-scrollning]] — ny sida från inboxen
- [create] [[notes/malikoner-anvander-enkel-oversattning-fore-iconify]] — ny sida från inboxen
- [create] [[notes/manuell-sortering-av-mal-pa-priority]] — ny sida från inboxen
- [create] [[notes/mellanslag-blockerades-i-malbeskrivningen]] — ny sida från inboxen
- [create] [[notes/nya-mal-far-ikon-fran-titeln]] — ny sida från inboxen
- [create] [[notes/plan-api-anvander-scalar-for-openapi]] — ny sida från inboxen
- [create] [[notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat]] — ny sida från inboxen
- [create] [[notes/vertikal-kalender-med-ikonstod-i-plan]] — ny sida från inboxen
- [create] [[notes/vertikal-kalender-med-kontinuerlig-zoom]] — ny sida från inboxen
- [archive] 2026-11-17-automatisk-malikon-behovde-sokord-och-svenska-alias.md
- [archive] 2026-11-17-kalenderns-zoom-styrs-av-vertikal-scrollning.md
- [archive] 2026-11-17-malikoner-anvander-enkel-oversattning-fore-iconify.md
- [archive] 2026-11-17-manuell-sortering-av-mal-pa-priority.md
- [archive] 2026-11-17-mellanslag-blockerades-i-malbeskrivningen.md
- [archive] 2026-11-17-nya-mal-far-ikon-fran-titeln.md
- [archive] 2026-11-17-plan-api-anvander-scalar-for-openapi.md
- [archive] 2026-11-17-priorityfilter-ska-behandla-saknat-started-som-ej-paborjat.md
- [archive] 2026-11-17-vertikal-kalender-med-ikonstod-i-plan.md
- [archive] 2026-11-17-vertikal-kalender-med-kontinuerlig-zoom.md
## [2026-11-17] ingest | create notes/kalenderns-zoom-styrs-av-vertikal-scrollning
## [2026-11-17] ingest | create notes/malikoner-anvander-enkel-oversattning-fore-iconify
## [2026-11-17] ingest | create notes/manuell-sortering-av-mal-pa-priority
## [2026-11-17] ingest | create notes/mellanslag-blockerades-i-malbeskrivningen
## [2026-11-17] ingest | create notes/nya-mal-far-ikon-fran-titeln
## [2026-11-17] ingest | create notes/plan-api-anvander-scalar-for-openapi
## [2026-11-17] ingest | create notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat
## [2026-11-17] ingest | create notes/vertikal-kalender-med-ikonstod-i-plan
## [2026-11-17] ingest | create notes/vertikal-kalender-med-kontinuerlig-zoom


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/vertikal-kalender-med-kontinuerlig-zoom


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/kalenderns-zoom-styrs-av-vertikal-scrollning


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/nya-mal-far-ikon-fran-titeln


## [2026-11-17] source | 2026-11-17-kalenderns-detaljnivaer-och-lane-packning.md


## [2026-11-17] ingest | create notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning
- [create] [[notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning]] — ny sida från inboxen
- [archive] 2026-11-17-kalenderns-detaljnivaer-fem-nivaer-och-lane-packning.md


## [2026-11-17] link | notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning -> notes/vertikal-kalender-med-kontinuerlig-zoom


## [2026-11-17] link | notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning -> notes/vertikal-kalender-med-ikonstod-i-plan


## [2026-11-17] ingest | create notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10
- [create] [[notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10]] — ny sida från inboxen
- [archive] 2026-11-17-relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10.md


## [2026-11-17] link | notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10 -> notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning


## [2026-11-17] link | notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10 -> notes/rekryteringspipeline-board-per-person-sourcing-som-attribut


## [2026-11-17] ingest | create notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater
- [create] [[notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater]] — ny sida från inboxen
- [archive] 2026-11-17-detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater.md


## [2026-11-17] link | notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater -> notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10


## [2026-11-17] link | notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10 -> notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater


## [2026-11-17] page | summaries/kalenderns-detaljnivaer-och-lane-packning


## [2026-11-17] ingest | create notes/neon-api-token-hamnade-i-dotfiles-git-historiken
- [create] [[notes/neon-api-token-hamnade-i-dotfiles-git-historiken]] — ny sida från inboxen
- [create] [[notes/redigerare-och-agarskap-for-relationer-i-pappa]] — ny sida från inboxen
- [archive] 2026-11-17-neon-api-token-hamnade-i-dotfiles-git-historiken.md
- [archive] 2026-11-17-redigerare-och-agarskap-for-relationer-i-pappa.md
## [2026-11-17] ingest | create notes/redigerare-och-agarskap-for-relationer-i-pappa


## [2026-11-17] link | notes/redigerare-och-agarskap-for-relationer-i-pappa -> notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater


## [2026-11-17] link | notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater -> notes/redigerare-och-agarskap-for-relationer-i-pappa


## [2026-11-17] link | notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning -> summaries/kalenderns-detaljnivaer-och-lane-packning


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/plan-api-anvander-scalar-for-openapi


## [2026-11-17] link | notes/nya-mal-far-ikon-fran-titeln -> notes/automatisk-malikon-behovde-sokord-och-svenska-alias


## [2026-11-17] link | notes/nya-mal-far-ikon-fran-titeln -> notes/malikoner-anvander-enkel-oversattning-fore-iconify


## [2026-11-17] link | notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet -> notes/goals-listar-bara-grundmal-scroll-vid-j-k


## [2026-11-17] link | notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet -> notes/sortering-av-grundmal-pa-goals-med-piltangenterna


## [2026-11-17] link | notes/eviga-mal-i-plan-kan-inte-slutforas -> notes/manuell-sortering-av-mal-pa-priority


## [2026-11-17] link | notes/eviga-mal-i-plan-kan-inte-slutforas -> notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat


## [2026-11-17] link | notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t -> notes/mellanslag-blockerades-i-malbeskrivningen


## [2026-11-17] link | notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt -> notes/started-finished-kunde-inte-rensas-i-plan-null-ignorerades-i-go-api-t


## [2026-11-17] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/tilt-filbevakning-kan-do-starta-om-tilt-up


## [2026-11-17] link | notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo -> notes/admin-inloggning-via-losenord-i-stallet-for-token-i-ui-t


## [2026-11-17] link | notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo -> notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet


## [2026-11-17] link | notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo -> notes/kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr


## [2026-11-17] link | notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo -> notes/kameo-sidan-flyttad-till-toppniva-och-delad-i-flikar


## [2026-11-17] link | notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa -> notes/mailagent-bygget-brot-pa-ofardig-kameo-integration


## [2026-11-17] link | notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa -> notes/pappa-regelendpointet-laser-och-sparar-riktig-rules-md-var-stub


## [2026-11-17] link | notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon -> notes/nuxt-sidomeny-kandidater-som-startsida-i-pappa


## [2026-11-17] link | notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon -> notes/mailvyn-mail-omstegad-till-tailwind-tema


## [2026-11-17] link | notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code -> notes/one-com-lagrar-mappar-i-nfd-normalisera-i-getfoldername


## [2026-11-17] link | notes/plan-appen-morkt-designsystem-med-semantiska-farg-token -> notes/mailvyn-mail-omstegad-till-tailwind-tema


## [2026-11-17] link | notes/rantebetalningsmejl-hamnar-i-inbox-deleted-messages-och-arkivet-matchades-aldrig -> notes/kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr


## [2026-11-17] ingest | create notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin
- [create] [[notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin]] — ny sida från inboxen
- [archive] 2026-11-17-pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin.md


## [2026-11-17] link | notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater -> notes/redigerare-och-agarskap-for-relationer-i-pappa


## [2026-11-17] link | notes/kunskaps-wiki -> notes/neon-api-token-hamnade-i-dotfiles-git-historiken


## [2026-11-17] link | notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin -> notes/redigerare-och-agarskap-for-relationer-i-pappa


## [2026-11-17] link | notes/redigerare-och-agarskap-for-relationer-i-pappa -> notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin


## [2026-11-17] unlink | notes/kunskaps-wiki -> notes/neon-api-token-hamnade-i-dotfiles-git-historiken


## [2026-11-17] link | notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo -> notes/neon-api-token-hamnade-i-dotfiles-git-historiken


## [2026-11-17] link | notes/mailvyn-mail-omstegad-till-tailwind-tema -> notes/plan-appen-morkt-designsystem-med-semantiska-farg-token


## [2026-11-17] link | notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet -> notes/rantebetalningsmejl-hamnar-i-inbox-deleted-messages-och-arkivet-matchades-aldrig


## [2026-11-17] link | notes/rust-utvecklingsmiljo-pa-workstationen -> notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet


## [2026-11-17] link | notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet -> notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer


## [2026-11-17] link | notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer -> notes/rust-analyzer-tysta-health-status-notisen-istallet-for-att-stanga-av-standalone


## [2026-11-17] link | notes/rust-analyzer-tysta-health-status-notisen-istallet-for-att-stanga-av-standalone -> notes/sa-dampas-bada-rust-analyzer-popupen-no-project-root-found


## [2026-11-17] link | notes/rust-utvecklingsmiljo-pa-workstationen -> notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork


## [2026-11-17] link | notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork -> notes/rust-repo-learn-to-code-with-rust-som-submodule


## [2026-11-17] link | notes/rust-repo-learn-to-code-with-rust-som-submodule -> notes/rust-utvecklingsmiljo-pa-workstationen


## [2026-11-17] link | notes/qbittorrent-installerad-pa-workstation-omarchy-arch -> notes/popcorn-time-installerad-som-appimage-i-local


## [2026-11-17] link | notes/popcorn-time-installerad-som-appimage-i-local -> notes/popcorn-time-remote-movies-api-failed-to-respond-orsak-och-fix


## [2026-11-17] link | notes/popcorn-time-remote-movies-api-failed-to-respond-orsak-och-fix -> notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection


## [2026-11-17] link | notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection -> notes/qbittorrent-ar-standardhanterare-for-torrent-och-magnet-lankar


## [2026-11-17] link | notes/qbittorrent-ar-standardhanterare-for-torrent-och-magnet-lankar -> notes/qbittorrent-installerad-pa-workstation-omarchy-arch


## [2026-11-17] link | notes/lightcraft-photocraft-installerade-pa-omarchy-workstation-utan-sudo -> notes/photocraft-agentstyrning-via-control-protokoll-control-7878


## [2026-11-17] link | notes/photocraft-agentstyrning-via-control-protokoll-control-7878 -> notes/photocraft-kamerans-look-varning-vid-raw-oppning-och-hur-den-hanteras


## [2026-11-17] link | notes/photocraft-kamerans-look-varning-vid-raw-oppning-och-hur-den-hanteras -> notes/lightcraft-photocraft-installerade-pa-omarchy-workstation-utan-sudo


## [2026-11-17] link | notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet -> notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet


## [2026-11-17] link | notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken -> notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11


## [2026-11-17] link | notes/pxe-aterstallning-nodspecifika-matchbox-profiler-och-maintenance-mode-flode -> notes/recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten


## [2026-11-17] link | notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc -> notes/rescue-wipe-pa-rock-maste-tomma-hela-systemdisken-wipe-system-esp-filen-ar-det-som-fangar-booten


## [2026-11-17] link | notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first -> notes/rock-4c-4se-talos-kan-aldrig-installeras-till-disk-u-boot-saknar-efi-setvariablert


## [2026-11-17] link | notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen -> notes/rock-4c-4se-utan-batteribackup-rtc-star-pa-2017-tls-mot-factory-talos-dev-misslyckas


## [2026-11-17] link | notes/controlplane-recovery-via-pxe-ar-inte-konfigurerad -> notes/talos-etcd-en-gammal-ephemeral-kopia-pa-controlplanet-ger-en-trasig-raft-log


## [2026-11-17] link | notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument -> notes/talos-hostname-auto-i-en-generisk-config-byter-nodens-namn-och-kastar-ut-den-ur-klustret


## [2026-11-17] link | notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode -> notes/talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod


## [2026-11-17] link | notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode -> notes/talos-noden-kan-hamna-i-maintenance-mode-om-state-volymen-hittas-efter-config-fasen


## [2026-11-17] link | notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc -> notes/talos-v1-14-utan-unattendedinstallconfig-oinstallerad-metal-nod-bootar-om-i-en-loop


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/plan-api-anvander-scalar-for-openapi


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/repo-template-ny-rust-api-src-rust-med-axum-utoipa-scalar


## [2026-11-17] link | notes/tilt-filbevakning-kan-do-starta-om-tilt-up -> notes/repo-template-rust-api-kopplad-till-kubernetes-manifests-tilt


## [2026-11-17] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg


## [2026-11-17] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/ci-saknade-inloggning-mot-ghcr-alla-image-pushar-misslyckades


## [2026-11-17] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/repo-template-ci-gitops-flodet-fixat-bygger-3-images-ghcr-tagging-pr-overlays-per-pr


## [2026-11-17] link | notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas -> notes/boka-enskild-kommunbokning-i-matchi-fran-bookings


## [2026-11-17] link | notes/synk-hallbokningar-kommunens-actor-smartbook-matchi -> notes/gron-markering-av-tider-som-ar-bokade-i-matchi-skrivlasande-jamforelse


## [2026-11-17] link | notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass -> notes/matchi-inloggningen-ar-flaky-klienten-forsoker-om-och-dev-klustret-behover-ombyggnad


## [2026-11-17] link | notes/synk-hallbokningar-kommunens-actor-smartbook-matchi -> notes/query-parametrar-i-actor-smartbook-endpointen-url-avkodas-till-mellanslag


## [2026-11-17] link | notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon -> notes/lunchmenyn-delar-en-modell-mellan-byggskript-och-webblasare


## [2026-11-17] link | notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon -> notes/nuxt-cli-via-bun-run-dev-gar-sonder-i-stk


## [2026-11-17] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/transienta-502-fra-go-api-t-nar-poddar-startar-om-i-dev-klustret


## [2026-11-17] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/stk-dev-miljon-kor-mot-neon-istallet-for-cnpg


## [2026-11-17] link | notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres -> notes/plan-dev-kor-mot-neon-istallet-for-lokal-postgres


## [2026-11-17] link | notes/learn-kunskapssystem-implementerat-i-src-python -> notes/learn-calibre-synk-och-automatisk-import-via-systemd


## [2026-11-17] link | notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas -> notes/learn-designsystem-a3-folio-i-nuxt


## [2026-11-17] link | notes/learn-kunskapsbasen-kan-lagras-i-postgresql -> notes/learn-korsanalys-av-varje-bok-mot-den-samlade-kunskapsbasen


## [2026-11-17] link | notes/learn-kunskapssystemet-portat-fran-python-till-go -> notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas


## [2026-11-17] link | notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo -> notes/kameo-uppgifter-synkas-till-fil-sa-docker-slipper-op-cli


## [2026-11-17] link | notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet -> notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo


## [2026-11-17] link | notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter -> notes/mailmaster-api-v1-gav-failed-to-fetch-go-servern-panickade-och-saknade-cors


## [2026-11-17] link | notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa -> notes/mailmaster-gammal-container-build-korde-process-och-arkiverade-mejl


## [2026-11-17] link | notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan -> notes/mailmaster-granskning-av-ai-atgarder-fore-utforande


## [2026-11-17] link | notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema -> notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan


## [2026-11-17] link | notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo -> notes/pappa-hemlighet-i-build-context-env-i-docker-imagen


## [2026-11-17] link | notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades -> notes/plan-cachad-spa-skal-gjorde-att-webblasaren-korde-gammalt-js-bundle-med-ikon-provider-none


## [2026-11-17] link | notes/sortering-av-grundmal-pa-goals-med-piltangenterna -> notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan


## [2026-11-17] link | notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink -> notes/ipv6-ingress-till-envoy-designlage-2026-11-15


## [2026-11-17] link | notes/ipv6-ingress-till-envoy-designlage-2026-11-15 -> notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink


## [2026-11-17] link | notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe -> notes/u7-lite-94-2a-6f-7a-4b-33-adopterar-inte-saknar-management-ip-och-nattverkssegment-pa-switchen


## [2026-11-17] link | notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen -> notes/twingate-ressurserna-forsvann-pa-arbetsstationen-2026-11-04-hemnatet-10-1-1-1-och-datacenter-10-10-10-0-24-oppnarbara


## [2026-11-17] link | notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen -> notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02


## [2026-11-17] link | notes/accuracy-fel-omstart -> notes/median-i-stallet-for-snitt-for-wpm


## [2026-11-17] link | notes/kunskaps-wiki -> notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin


## [2026-11-17] link | notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent -> notes/simonbrundin-gmail-com-inkorgen-kan-lasas-via-imap-begbot-losenordet


## [2026-11-17] ingest | create notes/mal-kan-ha-due-datum-deadline
- [create] [[notes/mal-kan-ha-due-datum-deadline]] — ny sida från inboxen
- [archive] 2026-11-17-mal-kan-ha-due-datum-deadline.md


## [2026-11-17] ingest | create notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer
- [create] [[notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer]] — ny sida från inboxen
- [archive] 2026-11-17-relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer.md


## [2026-11-17] link | notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer -> notes/redigerare-och-agarskap-for-relationer-i-pappa


## [2026-11-17] link | notes/redigerare-och-agarskap-for-relationer-i-pappa -> notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer


## [2026-11-17] link | notes/eviga-mal-i-plan-kan-inte-slutforas -> notes/mal-kan-ha-due-datum-deadline


## [2026-11-17] link | notes/manuell-sortering-av-mal-pa-priority -> notes/mal-kan-ha-due-datum-deadline


## [2026-11-17] link | notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat -> notes/mal-kan-ha-due-datum-deadline


## [2026-11-17] ingest | create notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage
- [create] [[notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage]] — ny sida från inboxen
- [archive] 2026-11-17-relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage.md


## [2026-11-17] link | notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage -> notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer


## [2026-11-17] link | notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer -> notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage


## [2026-11-17] ingest | create notes/losning-1password-backed-env-variabel-for-mcp-tokens
- [create] [[notes/losning-1password-backed-env-variabel-for-mcp-tokens]] — ny sida från inboxen
- [archive] 2026-11-17-losning-1password-backed-env-variabel-for-mcp-tokens.md


## [2026-11-17] link | notes/losning-1password-backed-env-variabel-for-mcp-tokens -> notes/neon-api-token-hamnade-i-dotfiles-git-historiken


## [2026-11-17] link | notes/neon-api-token-hamnade-i-dotfiles-git-historiken -> notes/losning-1password-backed-env-variabel-for-mcp-tokens


## [2026-11-17] ingest | create notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis
- [create] [[notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis]] — ny sida från inboxen
- [archive] 2026-11-17-kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis.md


## [2026-11-17] link | notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis -> notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning


## [2026-11-17] link | notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis -> notes/vertikal-kalender-med-kontinuerlig-zoom


## [2026-11-17] link | notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning -> notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis


## [2026-11-17] link | notes/vertikal-kalender-med-kontinuerlig-zoom -> notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis


## [2026-11-17] ingest | create notes/parallella-agent-sessioner-i-samma-repo-och-databas
- [create] [[notes/parallella-agent-sessioner-i-samma-repo-och-databas]] — ny sida från inboxen
- [archive] 2026-11-17-parallella-agent-sessioner-i-samma-repo-och-databas.md


## [2026-11-17] link | notes/parallella-agent-sessioner-i-samma-repo-och-databas -> notes/vertikal-kalender-med-kontinuerlig-zoom


## [2026-11-17] link | notes/vertikal-kalender-med-kontinuerlig-zoom -> notes/parallella-agent-sessioner-i-samma-repo-och-databas


## [2026-11-17] link | notes/parallella-agent-sessioner-i-samma-repo-och-databas -> notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis


## [2026-11-17] ingest | create notes/kandidatlistan-heter-nu-personer
- [create] [[notes/kandidatlistan-heter-nu-personer]] — ny sida från inboxen
- [create] [[notes/personer-kan-laggas-in-manuellt-i-pappa]] — ny sida från inboxen
- [archive] 2026-11-17-kandidatlistan-heter-nu-personer.md
- [archive] 2026-11-17-personer-kan-laggas-in-manuellt-i-pappa.md
## [2026-11-17] ingest | create notes/personer-kan-laggas-in-manuellt-i-pappa


## [2026-11-17] link | notes/kandidatlistan-heter-nu-personer -> notes/nuxt-sidomeny-kandidater-som-startsida-i-pappa


## [2026-11-17] link | notes/personer-kan-laggas-in-manuellt-i-pappa -> notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning


## [2026-11-17] link | notes/personer-kan-laggas-in-manuellt-i-pappa -> notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer


## [2026-11-17] ingest | create notes/min-kontaktlista-jag-som-nod-i-grafen-pappa
- [create] [[notes/min-kontaktlista-jag-som-nod-i-grafen-pappa]] — ny sida från inboxen
- [archive] 2026-11-17-min-kontaktlista-jag-som-nod-i-grafen-pappa.md


## [2026-11-17] link | notes/min-kontaktlista-jag-som-nod-i-grafen-pappa -> notes/redigerare-och-agarskap-for-relationer-i-pappa


## [2026-11-17] link | notes/redigerare-och-agarskap-for-relationer-i-pappa -> notes/min-kontaktlista-jag-som-nod-i-grafen-pappa


## [2026-11-17] link | notes/nuxt-sidomeny-kandidater-som-startsida-i-pappa -> notes/kandidatlistan-heter-nu-personer


## [2026-11-17] link | notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning -> notes/personer-kan-laggas-in-manuellt-i-pappa


## [2026-11-17] ingest | create notes/kalenderns-tidsaxelrader-animeras-vid-zoom
- [create] [[notes/kalenderns-tidsaxelrader-animeras-vid-zoom]] — ny sida från inboxen
- [archive] 2026-11-17-kalenderns-tidsaxelrader-animeras-vid-zoom.md


## [2026-11-17] link | notes/kalenderns-tidsaxelrader-animeras-vid-zoom -> notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis


## [2026-11-17] link | notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis -> notes/kalenderns-tidsaxelrader-animeras-vid-zoom


## [2026-11-17] ingest | create notes/kalenderhandelser-kan-ha-egna-farger
- [create] [[notes/kalenderhandelser-kan-ha-egna-farger]] — ny sida från inboxen
- [create] [[notes/kalendern-begransar-aterkommande-forekomster-i-langtidvyer]] — ny sida från inboxen
- [create] [[notes/kalendern-begransas-till-tre-manader]] — ny sida från inboxen
- [create] [[notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt]] — ny sida från inboxen
- [create] [[notes/kalenderns-scroll-axlar-bytta-lodrat-zoomar-vagratt-panorerar]] — ny sida från inboxen
- [create] [[notes/michael-brundin-blir-en-personnod-i-relationsgrafen]] — ny sida från inboxen
- [create] [[notes/relationsgrafen-anvander-fast-michael-id-utan-person-kind]] — ny sida från inboxen
- [create] [[notes/soulmate-teamet-i-sundsvall-importerades]] — ny sida från inboxen
- [create] [[notes/tydligare-hjalpmodal-for-kalenders-kortkommandon]] — ny sida från inboxen
- [archive] 2026-11-17-kalenderhandelser-kan-ha-egna-farger.md
- [archive] 2026-11-17-kalendern-begransar-aterkommande-forekomster-i-langtidvyer.md
- [archive] 2026-11-17-kalendern-begransas-till-tre-manader.md
- [archive] 2026-11-17-kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt.md
- [archive] 2026-11-17-kalenderns-scroll-axlar-bytta-lodrat-zoomar-vagratt-panorerar.md
- [archive] 2026-11-17-michael-brundin-blir-en-personnod-i-relationsgrafen.md
- [archive] 2026-11-17-relationsgrafen-anvander-fast-michael-id-utan-person-kind.md
- [archive] 2026-11-17-soulmate-teamet-i-sundsvall-importerades.md
- [archive] 2026-11-17-tydligare-hjalpmodal-for-kalenders-kortkommandon.md
## [2026-11-17] ingest | create notes/kalendern-begransar-aterkommande-forekomster-i-langtidvyer
## [2026-11-17] ingest | create notes/kalendern-begransas-till-tre-manader
## [2026-11-17] ingest | create notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt
## [2026-11-17] ingest | create notes/kalenderns-scroll-axlar-bytta-lodrat-zoomar-vagratt-panorerar
## [2026-11-17] ingest | create notes/michael-brundin-blir-en-personnod-i-relationsgrafen
## [2026-11-17] ingest | create notes/relationsgrafen-anvander-fast-michael-id-utan-person-kind
## [2026-11-17] ingest | create notes/soulmate-teamet-i-sundsvall-importerades
## [2026-11-17] ingest | create notes/tydligare-hjalpmodal-for-kalenders-kortkommandon


## [2026-11-17] link | notes/vertikal-kalender-med-kontinuerlig-zoom -> notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt


## [2026-11-17] link | notes/kalendern-begransar-aterkommande-forekomster-i-langtidvyer -> notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt


## [2026-11-17] ingest | create notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva
- [create] [[notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva]] — ny sida från inboxen
- [archive] 2026-11-17-aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva.md


## [2026-11-17] link | notes/kalendern-begransar-aterkommande-forekomster-i-langtidvyer -> notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva


## [2026-11-17] ingest | create notes/lokal-expansion-kan-verifieras-mot-serverns-expansion
- [create] [[notes/lokal-expansion-kan-verifieras-mot-serverns-expansion]] — ny sida från inboxen
- [archive] 2026-11-17-lokal-expansion-kan-verifieras-mot-serverns-expansion.md


## [2026-11-17] link | notes/lokal-expansion-kan-verifieras-mot-serverns-expansion -> notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt


## [2026-11-17] ingest | create notes/flera-kalendrar-med-egna-farger-i-plan
- [create] [[notes/flera-kalendrar-med-egna-farger-i-plan]] — ny sida från inboxen
- [archive] 2026-11-17-flera-kalendrar-med-egna-farger-i-plan.md


## [2026-11-17] link | notes/flera-kalendrar-med-egna-farger-i-plan -> notes/vertikal-kalender-med-ikonstod-i-plan


## [2026-11-17] link | notes/vertikal-kalender-med-ikonstod-i-plan -> notes/flera-kalendrar-med-egna-farger-i-plan


## [2026-11-18] ingest | create notes/standardkalender-i-plan
- [create] [[notes/standardkalender-i-plan]] — ny sida från inboxen
- [archive] 2026-11-18-standardkalender-i-plan.md


## [2026-11-18] link | notes/flera-kalendrar-med-egna-farger-i-plan -> notes/standardkalender-i-plan


## [2026-11-18] link | notes/standardkalender-i-plan -> notes/flera-kalendrar-med-egna-farger-i-plan


## [2026-11-18] ingest | create notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret
- [create] [[notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret]] — ny sida från inboxen
- [create] [[notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta]] — ny sida från inboxen
- [archive] 2026-11-18-en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret.md
- [archive] 2026-11-18-uttunnade-serier-vags-sa-att-antalen-blir-exakta.md
## [2026-11-18] ingest | create notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta


## [2026-11-18] link | notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta -> notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva


## [2026-11-18] link | notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret -> notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta


## [2026-11-18] link | notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva -> notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret


## [2026-11-18] ingest | create notes/handelsen-har-ingen-egen-farg
- [create] [[notes/handelsen-har-ingen-egen-farg]] — ny sida från inboxen
- [archive] 2026-11-18-handelsen-har-ingen-egen-farg.md


## [2026-11-18] link | notes/handelsen-har-ingen-egen-farg -> notes/flera-kalendrar-med-egna-farger-i-plan


## [2026-11-18] link | notes/flera-kalendrar-med-egna-farger-i-plan -> notes/handelsen-har-ingen-egen-farg


## [2026-11-18] link | notes/handelsen-har-ingen-egen-farg -> notes/standardkalender-i-plan


## [2026-11-18] ingest | create notes/reglerna-i-pappa-lag-i-containerns-filysystem-sparade-regler-forsvann-vid-omstart
- [create] [[notes/reglerna-i-pappa-lag-i-containerns-filysystem-sparade-regler-forsvann-vid-omstart]] — ny sida från inboxen
- [archive] 2026-11-18-reglerna-i-pappa-lag-i-containerns-filysystem-sparade-regler-forsvann-vid-omstart.md


## [2026-11-18] ingest | create notes/valfri-fargkod-pa-kalendrar
- [create] [[notes/valfri-fargkod-pa-kalendrar]] — ny sida från inboxen
- [archive] 2026-11-18-valfri-fargkod-pa-kalendrar.md


## [2026-11-18] link | notes/valfri-fargkod-pa-kalendrar -> notes/flera-kalendrar-med-egna-farger-i-plan


## [2026-11-18] link | notes/flera-kalendrar-med-egna-farger-i-plan -> notes/valfri-fargkod-pa-kalendrar


## [2026-11-18] ingest | create notes/imap-sessionen-i-mailagenten-ar-inte-tradsaker-parallella-listanrop-forstor-anslutningen
- [create] [[notes/imap-sessionen-i-mailagenten-ar-inte-tradsaker-parallella-listanrop-forstor-anslutningen]] — ny sida från inboxen
- [create] [[notes/mailagenten-expanderar-bifogade-eml-till-egna-mejl]] — ny sida från inboxen
- [archive] 2026-11-18-imap-sessionen-i-mailagenten-ar-inte-tradsaker-parallella-listanrop-forstor-anslutningen.md
- [archive] 2026-11-18-mailagenten-expanderar-bifogade-eml-till-egna-mejl.md
## [2026-11-18] ingest | create notes/mailagenten-expanderar-bifogade-eml-till-egna-mejl


## [2026-11-18] ingest | create notes/marknadslage-panel-hogst-upp-pa-personer-i-pappa
- [create] [[notes/marknadslage-panel-hogst-upp-pa-personer-i-pappa]] — ny sida från inboxen
- [archive] 2026-11-18-marknadslage-panel-hogst-upp-pa-personer-i-pappa.md


## [2026-11-18] ingest | create notes/visionite-och-kvadrat-i-sundsvall-kartlagt-via-bolagsregister-visionites-kontaktsida-och-linkedin
- [create] [[notes/visionite-och-kvadrat-i-sundsvall-kartlagt-via-bolagsregister-visionites-kontaktsida-och-linkedin]] — ny sida från inboxen
- [archive] 2026-11-18-visionite-och-kvadrat-i-sundsvall-kartlagt-via-bolagsregister-visionites-kontaktsida-och-linkedin.md


## [2026-11-18] ingest | create notes/flerval-parkera-flera-kandidater-samtidigt-i-personer
- [create] [[notes/flerval-parkera-flera-kandidater-samtidigt-i-personer]] — ny sida från inboxen
- [archive] 2026-11-18-flerval-parkera-flera-kandidater-samtidigt-i-personer.md


## [2026-11-18] ingest | create notes/kalenderhandelser-kan-tas-bort-och-dupliceras
- [create] [[notes/kalenderhandelser-kan-tas-bort-och-dupliceras]] — ny sida från inboxen
- [archive] 2026-11-18-kalenderhandelser-kan-tas-bort-och-dupliceras.md


## [2026-11-18] ingest | create notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag
- [create] [[notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag]] — ny sida från inboxen
- [archive] 2026-11-18-konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag.md


## [2026-11-18] link | notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag -> notes/marknadslage-panel-hogst-upp-pa-personer-i-pappa


## [2026-11-18] link | notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag -> notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret


## [2026-11-18] link | notes/marknadslage-panel-hogst-upp-pa-personer-i-pappa -> notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret


## [2026-11-18] ingest | create notes/expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append
- [create] [[notes/expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append]] — ny sida från inboxen
- [archive] 2026-11-18-expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append.md


## [2026-11-18] link | notes/konsulter-i-sundsvall-antalet-gar-inte-att-mata-det-som-gar-att-saga -> notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret


## [2026-11-18] link | notes/konsulter-i-sundsvall-antalet-gar-inte-att-mata-det-som-gar-att-saga -> notes/marknadslage-panel-hogst-upp-pa-personer-i-pappa


## [2026-11-18] link | notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag -> notes/konsulter-i-sundsvall-antalet-gar-inte-att-mata-det-som-gar-att-saga


## [2026-11-18] ingest | create notes/linkedin-relationer-i-pappa-422-svaga-samband-inlagda-och-linkedin-bot-sp-u00e4rr-kontot
- [create] [[notes/linkedin-relationer-i-pappa-422-svaga-samband-inlagda-och-linkedin-bot-sp-u00e4rr-kontot]] — ny sida från inboxen
- [archive] 2026-11-18-linkedin-relationer-i-pappa-422-svaga-samband-inlagda-och-linkedin-bot-sp-u00e4rr-kontot.md


## [2026-11-18] ingest | create notes/sessionen-saknade-executor-varje-godkand-atgard-kraschade-med-502
- [create] [[notes/sessionen-saknade-executor-varje-godkand-atgard-kraschade-med-502]] — ny sida från inboxen
- [archive] 2026-11-18-sessionen-saknade-executor-varje-godkand-atgard-kraschade-med-502.md


## [2026-11-18] ingest | create notes/kalendern-visar-handelser-som-rader-aldrig-som-prickar
- [create] [[notes/kalendern-visar-handelser-som-rader-aldrig-som-prickar]] — ny sida från inboxen
- [archive] 2026-11-18-kalendern-visar-handelser-som-rader-aldrig-som-prickar.md


## [2026-11-18] ingest | create notes/kalendern-en-kolumn-dolda-serier-och-en-installningssida
- [create] [[notes/kalendern-en-kolumn-dolda-serier-och-en-installningssida]] — ny sida från inboxen
- [archive] 2026-11-18-kalendern-en-kolumn-dolda-serier-och-en-installningssida.md


## [2026-11-18] ingest | create notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev
- [create] [[notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev]] — ny sida från inboxen
- [archive] 2026-11-18-agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev.md


## [2026-11-18] ingest | create notes/platser-pa-mal-och-kalenderhandelser-openspec-change-add-places
- [create] [[notes/platser-pa-mal-och-kalenderhandelser-openspec-change-add-places]] — ny sida från inboxen
- [archive] 2026-11-18-platser-pa-mal-och-kalenderhandelser-openspec-change-add-places.md


## [2026-11-18] ingest | create notes/aterkommande-mal-samma-malobjekt-med-upprepningsregel
- [create] [[notes/aterkommande-mal-samma-malobjekt-med-upprepningsregel]] — ny sida från inboxen
- [archive] 2026-11-18-aterkommande-mal-samma-malobjekt-med-upprepningsregel.md


## [2026-11-19] ingest | create notes/linkedin-lankar-for-217-av-230-personer-13-kvarvarande-oroliga
- [create] [[notes/linkedin-lankar-for-217-av-230-personer-13-kvarvarande-oroliga]] — ny sida från inboxen
- [archive] 2026-11-19-linkedin-lankar-for-217-av-230-personer-13-kvarvarande-oroliga.md


## [2026-11-19] ingest | create notes/kalendern-bara-etikettbandet-reserverar-plats-i-kolumnen
- [create] [[notes/kalendern-bara-etikettbandet-reserverar-plats-i-kolumnen]] — ny sida från inboxen
- [archive] 2026-11-19-kalendern-bara-etikettbandet-reserverar-plats-i-kolumnen.md


## [2026-11-19] ingest | create notes/moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in
- [create] [[notes/moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in]] — ny sida från inboxen
- [archive] 2026-11-19-moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in.md


## [2026-11-19] link | notes/moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in -> notes/expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append


## [2026-11-19] ingest | create notes/popup-editor-i-klickbar-rad-klicken-bubblar-och-navigerar-bort
- [create] [[notes/popup-editor-i-klickbar-rad-klicken-bubblar-och-navigerar-bort]] — ny sida från inboxen
- [archive] 2026-11-19-popup-editor-i-klickbar-rad-klicken-bubblar-och-navigerar-bort.md


## [2026-11-19] ingest | create notes/kalendern-urval-istallet-for-dold-serie
- [create] [[notes/kalendern-urval-istallet-for-dold-serie]] — ny sida från inboxen
- [archive] 2026-11-19-kalendern-urval-istallet-for-dold-serie.md


## [2026-11-19] ingest | create notes/kalendern-urvalet-togs-bort-serien-doljs-som-helhet
- [create] [[notes/kalendern-urvalet-togs-bort-serien-doljs-som-helhet]] — ny sida från inboxen
- [archive] 2026-11-19-kalendern-urvalet-togs-bort-serien-doljs-som-helhet.md


## [2026-11-19] ingest | create notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor
- [create] [[notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor]] — ny sida från inboxen
- [create] [[notes/mailgranskningen-visar-ett-mejl-i-taget]] — ny sida från inboxen
- [archive] 2026-11-19-leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor.md
- [archive] 2026-11-19-mailgranskningen-visar-ett-mejl-i-taget.md
## [2026-11-19] ingest | create notes/mailgranskningen-visar-ett-mejl-i-taget


## [2026-11-19] link | notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor -> notes/mellanslag-blockerades-i-malbeskrivningen


## [2026-11-19] link | notes/mellanslag-blockerades-i-malbeskrivningen -> notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor


## [2026-11-19] ingest | create notes/feedback-loop-ai-uppdaterar-mailreglerna-och-analyserar-om-mejlet
- [create] [[notes/feedback-loop-ai-uppdaterar-mailreglerna-och-analyserar-om-mejlet]] — ny sida från inboxen
- [create] [[notes/kalendern-sammanfattningen-av-serier-togs-bort]] — ny sida från inboxen
- [create] [[notes/mejlbrodtext-renderas-som-markdown-i-granskningsvyn]] — ny sida från inboxen
- [archive] 2026-11-19-feedback-loop-ai-uppdaterar-mailreglerna-och-analyserar-om-mejlet.md
- [archive] 2026-11-19-kalendern-sammanfattningen-av-serier-togs-bort.md
- [archive] 2026-11-19-mejlbrodtext-renderas-som-markdown-i-granskningsvyn.md
## [2026-11-19] ingest | create notes/kalendern-sammanfattningen-av-serier-togs-bort
## [2026-11-19] ingest | create notes/mejlbrodtext-renderas-som-markdown-i-granskningsvyn


## [2026-11-20] ingest | create notes/brevladelistan-maste-hamtas-innan-granskningslaget-startar
- [create] [[notes/brevladelistan-maste-hamtas-innan-granskningslaget-startar]] — ny sida från inboxen
- [create] [[notes/go-imap-uidstore-istallet-for-store-med-uid-mangd]] — ny sida från inboxen
- [create] [[notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder]] — ny sida från inboxen
- [create] [[notes/ny-rs232-hardvara-ak-co-02-db9-breakout-och-kopplingen-till-kvm-ns-3-pin]] — ny sida från inboxen
- [create] [[notes/ny-terminal-i-tmux-fragade-efter-datorns-losenord]] — ny sida från inboxen
- [create] [[notes/serveragd-ko-for-stegvis-mailgranskning]] — ny sida från inboxen
- [create] [[notes/windows-1252-avkodning-av-mejlens-amnesrad]] — ny sida från inboxen
- [archive] 2026-11-20-brevladelistan-maste-hamtas-innan-granskningslaget-startar.md
- [archive] 2026-11-20-go-imap-uidstore-istallet-for-store-med-uid-mangd.md
- [archive] 2026-11-20-malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder.md
- [archive] 2026-11-20-ny-rs232-hardvara-ak-co-02-db9-breakout-och-kopplingen-till-kvm-ns-3-pin.md
- [archive] 2026-11-20-ny-terminal-i-tmux-fragade-efter-datorns-losenord.md
- [archive] 2026-11-20-serveragd-ko-for-stegvis-mailgranskning.md
- [archive] 2026-11-20-windows-1252-avkodning-av-mejlens-amnesrad.md
## [2026-11-20] ingest | create notes/go-imap-uidstore-istallet-for-store-med-uid-mangd
## [2026-11-20] ingest | create notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder
## [2026-11-20] ingest | create notes/ny-rs232-hardvara-ak-co-02-db9-breakout-och-kopplingen-till-kvm-ns-3-pin
## [2026-11-20] ingest | create notes/ny-terminal-i-tmux-fragade-efter-datorns-losenord
## [2026-11-20] ingest | create notes/serveragd-ko-for-stegvis-mailgranskning
## [2026-11-20] ingest | create notes/windows-1252-avkodning-av-mejlens-amnesrad


## [2026-11-20] link | notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder -> notes/platser-pa-mal-och-kalenderhandelser-openspec-change-add-places


## [2026-11-20] link | notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder -> notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan


## [2026-11-20] link | notes/platser-pa-mal-och-kalenderhandelser-openspec-change-add-places -> notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder


## [2026-11-20] ingest | create notes/kvm-portmappning-worker-n-port-n-den-deployade-configen-pa-pi-n-var-fel
- [create] [[notes/kvm-portmappning-worker-n-port-n-den-deployade-configen-pa-pi-n-var-fel]] — ny sida från inboxen
- [create] [[notes/worker-3-bootade-via-pxe-men-talos-installen-nollstallde-de-forsta-11-mb-pa-microsd]] — ny sida från inboxen
- [archive] 2026-11-20-kvm-portmappning-worker-n-port-n-den-deployade-configen-pa-pi-n-var-fel.md
- [archive] 2026-11-20-worker-3-bootade-via-pxe-men-talos-installen-nollstallde-de-forsta-11-mb-pa-microsd.md
## [2026-11-20] ingest | create notes/worker-3-bootade-via-pxe-men-talos-installen-nollstallde-de-forsta-11-mb-pa-microsd


## [2026-11-20] ingest | create notes/en-talos-efi-partition-pa-kortet-blockerar-pxe-booten-bootflow-loaded-fore-natverket
- [create] [[notes/en-talos-efi-partition-pa-kortet-blockerar-pxe-booten-bootflow-loaded-fore-natverket]] — ny sida från inboxen
- [create] [[notes/talos-installen-nollstaller-de-forsta-11-mb-pa-microsd-loadern-maste-aterstallas]] — ny sida från inboxen
- [archive] 2026-11-20-en-talos-efi-partition-pa-kortet-blockerar-pxe-booten-bootflow-loaded-fore-natverket.md
- [archive] 2026-11-20-talos-installen-nollstaller-de-forsta-11-mb-pa-microsd-loadern-maste-aterstallas.md
## [2026-11-20] ingest | create notes/talos-installen-nollstaller-de-forsta-11-mb-pa-microsd-loadern-maste-aterstallas


## [2026-11-20] ingest | create notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium
- [create] [[notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium]] — ny sida från inboxen
- [create] [[notes/undo-stapel-i-plan-z-angrar-senaste-malhandling]] — ny sida från inboxen
- [archive] 2026-11-20-radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium.md
- [archive] 2026-11-20-undo-stapel-i-plan-z-angrar-senaste-malhandling.md
## [2026-11-20] ingest | create notes/undo-stapel-i-plan-z-angrar-senaste-malhandling


## [2026-11-20] link | notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium -> notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran


## [2026-11-20] link | notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium -> notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode


## [2026-11-20] link | notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium -> notes/talos-installen-nollstaller-de-forsta-11-mb-pa-microsd-loadern-maste-aterstallas


## [2026-11-20] link | notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium -> notes/en-talos-efi-partition-pa-kortet-blockerar-pxe-booten-bootflow-loaded-fore-natverket


## [2026-11-20] link | notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium -> notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat


## [2026-11-20] ingest | create notes/tidsatgang-pa-mal-i-plan-estimated-minutes
- [create] [[notes/tidsatgang-pa-mal-i-plan-estimated-minutes]] — ny sida från inboxen
- [archive] 2026-11-20-tidsatgang-pa-mal-i-plan-estimated-minutes.md


## [2026-11-20] link | notes/tidsatgang-pa-mal-i-plan-estimated-minutes -> notes/undo-stapel-i-plan-z-angrar-senaste-malhandling


## [2026-11-20] link | notes/tidsatgang-pa-mal-i-plan-estimated-minutes -> notes/plan-dev-kor-mot-neon-istallet-for-lokal-postgres


## [2026-11-20] link | notes/undo-stapel-i-plan-z-angrar-senaste-malhandling -> notes/tidsatgang-pa-mal-i-plan-estimated-minutes


## [2026-11-20] ingest | create notes/agentmarke-pa-mal-tangenten-a-ar-reserverad
- [create] [[notes/agentmarke-pa-mal-tangenten-a-ar-reserverad]] — ny sida från inboxen
- [create] [[notes/dold-skrivbordsmeny-med-leader-kommando]] — ny sida från inboxen
- [create] [[notes/kalender-som-permanent-hogerspalt-pa-desktop]] — ny sida från inboxen
- [create] [[notes/kalenderhandelser-aterstallda-fran-neon-wal-historik]] — ny sida från inboxen
- [create] [[notes/kalenderns-tangentbordskontroller-ska-inte-stora-malnavigering]] — ny sida från inboxen
- [create] [[notes/pxe-flash-hanterar-divergerat-infrastructure-repo]] — ny sida från inboxen
- [create] [[notes/talos-varningar-pa-controlplanes-90-kommer-fra-talos-discovery-service-dns]] — ny sida från inboxen
- [create] [[notes/worker-3-poe-omstart-och-talos-konfiguration-2026-11-20]] — ny sida från inboxen
- [archive] 2026-11-20-agentmarke-pa-mal-tangenten-a-ar-reserverad.md
- [archive] 2026-11-20-dold-skrivbordsmeny-med-leader-kommando.md
- [archive] 2026-11-20-kalender-som-permanent-hogerspalt-pa-desktop.md
- [archive] 2026-11-20-kalenderhandelser-aterstallda-fran-neon-wal-historik.md
- [archive] 2026-11-20-kalenderns-tangentbordskontroller-ska-inte-stora-malnavigering.md
- [archive] 2026-11-20-pxe-flash-hanterar-divergerat-infrastructure-repo.md
- [archive] 2026-11-20-talos-varningar-pa-controlplanes-90-kommer-fra-talos-discovery-service-dns.md
- [archive] 2026-11-20-worker-3-poe-omstart-och-talos-konfiguration-2026-11-20.md
## [2026-11-20] ingest | create notes/dold-skrivbordsmeny-med-leader-kommando
## [2026-11-20] ingest | create notes/kalender-som-permanent-hogerspalt-pa-desktop
## [2026-11-20] ingest | create notes/kalenderhandelser-aterstallda-fran-neon-wal-historik
## [2026-11-20] ingest | create notes/kalenderns-tangentbordskontroller-ska-inte-stora-malnavigering
## [2026-11-20] ingest | create notes/pxe-flash-hanterar-divergerat-infrastructure-repo
## [2026-11-20] ingest | create notes/talos-varningar-pa-controlplanes-90-kommer-fra-talos-discovery-service-dns
## [2026-11-20] ingest | create notes/worker-3-poe-omstart-och-talos-konfiguration-2026-11-20


## [2026-11-20] ingest | create notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard
- [create] [[notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard]] — ny sida från inboxen
- [archive] 2026-11-20-simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard.md


## [2026-11-20] link | notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard -> notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm


## [2026-11-20] link | notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard -> notes/kvm-portmappning-worker-n-port-n-den-deployade-configen-pa-pi-n-var-fel


## [2026-11-20] link | notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard -> entities/hdmi-kvm-switch-16-port-labbet


## [2026-11-20] link | notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm -> notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard


## [2026-11-20] ingest | create notes/atta-feature-commits-pushade-till-det-arkiverade-mailmaster-repot
- [create] [[notes/atta-feature-commits-pushade-till-det-arkiverade-mailmaster-repot]] — ny sida från inboxen
- [archive] 2026-11-20-atta-feature-commits-pushade-till-det-arkiverade-mailmaster-repot.md


## [2026-11-20] ingest | create notes/barnordning-foljer-beroenden-i-malsidan
- [create] [[notes/barnordning-foljer-beroenden-i-malsidan]] — ny sida från inboxen
- [archive] 2026-11-20-barnordning-foljer-beroenden-i-malsidan.md


## [2026-11-20] ingest | create notes/go-integrationstest-tomde-kalenderhandelserna-i-neon-igen
- [create] [[notes/go-integrationstest-tomde-kalenderhandelserna-i-neon-igen]] — ny sida från inboxen
- [archive] 2026-11-20-go-integrationstest-tomde-kalenderhandelserna-i-neon-igen.md


## [2026-11-20] link | notes/go-integrationstest-tomde-kalenderhandelserna-i-neon-igen -> notes/kalenderhandelser-aterstallda-fran-neon-wal-historik


## [2026-11-20] link | notes/go-integrationstest-tomde-kalenderhandelserna-i-neon-igen -> notes/parallella-agent-sessioner-i-samma-repo-och-databas


## [2026-11-20] ingest | create notes/mal-kan-ha-startdatum-syns-i-priority-forst-nar-datumet-passerat
- [create] [[notes/mal-kan-ha-startdatum-syns-i-priority-forst-nar-datumet-passerat]] — ny sida från inboxen
- [archive] 2026-11-20-mal-kan-ha-startdatum-syns-i-priority-forst-nar-datumet-passerat.md


## [2026-11-20] link | notes/mal-kan-ha-due-datum-deadline -> notes/mal-kan-ha-startdatum-syns-i-priority-forst-nar-datumet-passerat


## [2026-11-20] link | notes/mal-kan-ha-startdatum-syns-i-priority-forst-nar-datumet-passerat -> notes/mal-kan-ha-due-datum-deadline


## [2026-11-20] page | notes/worker-3-poe-omstart-och-talos-konfiguration-2026-11-20


## [2026-11-20] ingest | create notes/workmux-merge-misslyckas-i-huvud-worktree-kora-kommandot-i-feature-worktreet
- [create] [[notes/workmux-merge-misslyckas-i-huvud-worktree-kora-kommandot-i-feature-worktreet]] — ny sida från inboxen
- [archive] 2026-11-20-workmux-merge-misslyckas-i-huvud-worktree-kora-kommandot-i-feature-worktreet.md


## [2026-11-20] ingest | create notes/mal-kan-schemalaggas-och-kalenderhandelser-flyttas
- [create] [[notes/mal-kan-schemalaggas-och-kalenderhandelser-flyttas]] — ny sida från inboxen
- [create] [[notes/talos-dns-fallback-mot-publika-resolvrar]] — ny sida från inboxen
- [archive] 2026-11-20-mal-kan-schemalaggas-och-kalenderhandelser-flyttas.md
- [archive] 2026-11-20-talos-dns-fallback-mot-publika-resolvrar.md
## [2026-11-20] ingest | create notes/talos-dns-fallback-mot-publika-resolvrar


## [2026-11-20] ingest | create notes/leader-d-oppnar-gemensam-datum-popup-for-mal
- [create] [[notes/leader-d-oppnar-gemensam-datum-popup-for-mal]] — ny sida från inboxen
- [archive] 2026-11-20-leader-d-oppnar-gemensam-datum-popup-for-mal.md


## [2026-11-20] ingest | create notes/escape-i-datum-popup-maste-hanteras-fore-leader-listenern
- [create] [[notes/escape-i-datum-popup-maste-hanteras-fore-leader-listenern]] — ny sida från inboxen
- [create] [[notes/kortkommando-hjalpen-i-helpmodal-maste-folja-tangentandringar]] — ny sida från inboxen
- [create] [[notes/p-tangenten-konfliktade-med-viktlaget-pa-priority]] — ny sida från inboxen
- [create] [[notes/snabbtillagg-med-mal-och-valbart-foraldramal]] — ny sida från inboxen
- [create] [[notes/snabbtillaggets-popup-fick-escape-och-visuell-polish]] — ny sida från inboxen
- [create] [[notes/viktning-av-barnmal-med-balanserad-draglinje]] — ny sida från inboxen
- [create] [[notes/viktning-flyttad-till-varje-mal]] — ny sida från inboxen
- [create] [[notes/viktning-som-proportional-bar-under-maltitel]] — ny sida från inboxen
- [create] [[notes/vy-for-avslutade-mal]] — ny sida från inboxen
- [archive] 2026-11-20-escape-i-datum-popup-maste-hanteras-fore-leader-listenern.md
- [archive] 2026-11-20-kortkommando-hjalpen-i-helpmodal-maste-folja-tangentandringar.md
- [archive] 2026-11-20-p-tangenten-konfliktade-med-viktlaget-pa-priority.md
- [archive] 2026-11-20-snabbtillagg-med-mal-och-valbart-foraldramal.md
- [archive] 2026-11-20-snabbtillaggets-popup-fick-escape-och-visuell-polish.md
- [archive] 2026-11-20-viktning-av-barnmal-med-balanserad-draglinje.md
- [archive] 2026-11-20-viktning-flyttad-till-varje-mal.md
- [archive] 2026-11-20-viktning-som-proportional-bar-under-maltitel.md
- [archive] 2026-11-20-vy-for-avslutade-mal.md
## [2026-11-20] ingest | create notes/kortkommando-hjalpen-i-helpmodal-maste-folja-tangentandringar
## [2026-11-20] ingest | create notes/p-tangenten-konfliktade-med-viktlaget-pa-priority
## [2026-11-20] ingest | create notes/snabbtillagg-med-mal-och-valbart-foraldramal
## [2026-11-20] ingest | create notes/snabbtillaggets-popup-fick-escape-och-visuell-polish
## [2026-11-20] ingest | create notes/viktning-av-barnmal-med-balanserad-draglinje
## [2026-11-20] ingest | create notes/viktning-flyttad-till-varje-mal
## [2026-11-20] ingest | create notes/viktning-som-proportional-bar-under-maltitel
## [2026-11-20] ingest | create notes/vy-for-avslutade-mal


## [2026-11-20] ingest | create notes/vertikalt-viktlage-med-j-k-h-l
- [create] [[notes/vertikalt-viktlage-med-j-k-h-l]] — ny sida från inboxen
- [create] [[notes/worker-4-rock-4se-talos-bootar-lokal-nvme-installation-apid-startar-aldrig-pxe-kedjan-redo-bootordningen-kravs-fysiskt]] — ny sida från inboxen
- [archive] 2026-11-20-vertikalt-viktlage-med-j-k-h-l.md
- [archive] 2026-11-20-worker-4-rock-4se-talos-bootar-lokal-nvme-installation-apid-startar-aldrig-pxe-kedjan-redo-bootordningen-kravs-fysiskt.md
## [2026-11-20] ingest | create notes/worker-4-rock-4se-talos-bootar-lokal-nvme-installation-apid-startar-aldrig-pxe-kedjan-redo-bootordningen-kravs-fysiskt


## [2026-11-20] ingest | create notes/kalenderns-handelser-lag-over-modaler-och-escape-var-dovt
- [create] [[notes/kalenderns-handelser-lag-over-modaler-och-escape-var-dovt]] — ny sida från inboxen
- [archive] 2026-11-20-kalenderns-handelser-lag-over-modaler-och-escape-var-dovt.md


## [2026-11-20] ingest | create notes/hjalpmodalen-for-kortkommandon-fick-sokfilter
- [create] [[notes/hjalpmodalen-for-kortkommandon-fick-sokfilter]] — ny sida från inboxen
- [archive] 2026-11-20-hjalpmodalen-for-kortkommandon-fick-sokfilter.md


## [2026-11-20] ingest | create notes/kalenderrutorna-har-ingen-maxhojd-langre
- [create] [[notes/kalenderrutorna-har-ingen-maxhojd-langre]] — ny sida från inboxen
- [create] [[notes/stabiliserat-vertikalt-viktlage]] — ny sida från inboxen
- [archive] 2026-11-20-kalenderrutorna-har-ingen-maxhojd-langre.md
- [archive] 2026-11-20-stabiliserat-vertikalt-viktlage.md
## [2026-11-20] ingest | create notes/stabiliserat-vertikalt-viktlage


## [2026-11-20] ingest | create notes/dolda-ikoner-i-barnmalsrader-pa-malsidan-visa-vid-hover
- [create] [[notes/dolda-ikoner-i-barnmalsrader-pa-malsidan-visa-vid-hover]] — ny sida från inboxen
- [archive] 2026-11-20-dolda-ikoner-i-barnmalsrader-pa-malsidan-visa-vid-hover.md


## [2026-11-22] ingest | create notes/beroendevaljare-med-tangenten-w
- [create] [[notes/beroendevaljare-med-tangenten-w]] — ny sida från inboxen
- [create] [[notes/beroendevaljaren-kan-ta-bort-befintliga-beroenden]] — ny sida från inboxen
- [create] [[notes/ikonsokning-far-timeout-och-checkbox-fallback-vid-nya-mal]] — ny sida från inboxen
- [create] [[notes/priority-vyn-byter-namn-till-gor-nu]] — ny sida från inboxen
- [create] [[notes/stk-och-sundsvall-padelklubb-styrelsemotesschema-2026-och-var-handlingarna-finns]] — ny sida från inboxen
- [create] [[notes/worker-4-ateransluten-till-cluster1-via-lokalt-sd-kort]] — ny sida från inboxen
- [create] [[notes/worker-4-utan-pxe-talos-installation-pa-microsd-kortet-via-image-factory-grub-layout]] — ny sida från inboxen
- [archive] 2026-11-22-beroendevaljare-med-tangenten-w.md
- [archive] 2026-11-22-beroendevaljaren-kan-ta-bort-befintliga-beroenden.md
- [archive] 2026-11-22-ikonsokning-far-timeout-och-checkbox-fallback-vid-nya-mal.md
- [archive] 2026-11-22-priority-vyn-byter-namn-till-gor-nu.md
- [archive] 2026-11-22-stk-och-sundsvall-padelklubb-styrelsemotesschema-2026-och-var-handlingarna-finns.md
- [archive] 2026-11-22-worker-4-ateransluten-till-cluster1-via-lokalt-sd-kort.md
- [archive] 2026-11-22-worker-4-utan-pxe-talos-installation-pa-microsd-kortet-via-image-factory-grub-layout.md
## [2026-11-22] ingest | create notes/beroendevaljaren-kan-ta-bort-befintliga-beroenden
## [2026-11-22] ingest | create notes/ikonsokning-far-timeout-och-checkbox-fallback-vid-nya-mal
## [2026-11-22] ingest | create notes/priority-vyn-byter-namn-till-gor-nu
## [2026-11-22] ingest | create notes/stk-och-sundsvall-padelklubb-styrelsemotesschema-2026-och-var-handlingarna-finns
## [2026-11-22] ingest | create notes/worker-4-ateransluten-till-cluster1-via-lokalt-sd-kort
## [2026-11-22] ingest | create notes/worker-4-utan-pxe-talos-installation-pa-microsd-kortet-via-image-factory-grub-layout


## [2026-11-22] ingest | create notes/barnmarke-pa-mal-antal-undermal-visas-overallt
- [create] [[notes/barnmarke-pa-mal-antal-undermal-visas-overallt]] — ny sida från inboxen
- [create] [[notes/foraldrar-syns-pa-mal-i-now]] — ny sida från inboxen
- [archive] 2026-11-22-barnmarke-pa-mal-antal-undermal-visas-overallt.md
- [archive] 2026-11-22-foraldrar-syns-pa-mal-i-now.md
## [2026-11-22] ingest | create notes/foraldrar-syns-pa-mal-i-now


## [2026-11-22] ingest | create notes/leader-laget-kunde-inte-oppnas-efter-musklick-pa-en-knapp
- [create] [[notes/leader-laget-kunde-inte-oppnas-efter-musklick-pa-en-knapp]] — ny sida från inboxen
- [create] [[notes/nu-marke-pa-mal-leader-n-styr-gor-nu-listan]] — ny sida från inboxen
- [archive] 2026-11-22-leader-laget-kunde-inte-oppnas-efter-musklick-pa-en-knapp.md
- [archive] 2026-11-22-nu-marke-pa-mal-leader-n-styr-gor-nu-listan.md
## [2026-11-22] ingest | create notes/nu-marke-pa-mal-leader-n-styr-gor-nu-listan


## [2026-11-23] ingest | create notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko
- [create] [[notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko]] — ny sida från inboxen
- [archive] 2026-11-23-prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko.md


## [2026-11-23] link | notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko -> notes/sortering-av-grundmal-pa-goals-med-piltangenterna


## [2026-11-23] link | notes/sortering-av-grundmal-pa-goals-med-piltangenterna -> notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko


## [2026-11-23] ingest | create notes/1password-mcp-environments-pa-arch-omarchy-kraver-setgid-pa-1password-mcp
- [create] [[notes/1password-mcp-environments-pa-arch-omarchy-kraver-setgid-pa-1password-mcp]] — ny sida från inboxen
- [archive] 2026-11-23-1password-mcp-environments-pa-arch-omarchy-kraver-setgid-pa-1password-mcp.md


## [2026-11-23] link | notes/1password-mcp-environments-pa-arch-omarchy-kraver-setgid-pa-1password-mcp -> notes/losning-1password-backed-env-variabel-for-mcp-tokens


## [2026-11-23] ingest | create notes/lunchdeploy-triggas-fran-cronjob-pa-pxe
- [create] [[notes/lunchdeploy-triggas-fran-cronjob-pa-pxe]] — ny sida från inboxen
- [archive] 2026-11-23-lunchdeploy-triggas-fran-cronjob-pa-pxe.md


## [2026-11-23] ingest | create notes/hierarki-i-gor-nu-listan-now
- [create] [[notes/hierarki-i-gor-nu-listan-now]] — ny sida från inboxen
- [create] [[notes/taggar-pa-mal-i-plan-tangenten-g]] — ny sida från inboxen
- [archive] 2026-11-23-hierarki-i-gor-nu-listan-now.md
- [archive] 2026-11-23-taggar-pa-mal-i-plan-tangenten-g.md
## [2026-11-23] ingest | create notes/taggar-pa-mal-i-plan-tangenten-g


## [2026-11-23] ingest | create notes/figma-lokala-fonter-pa-linux-figma-agent-linux
- [create] [[notes/figma-lokala-fonter-pa-linux-figma-agent-linux]] — ny sida från inboxen
- [archive] 2026-11-23-figma-lokala-fonter-pa-linux-figma-agent-linux.md


## [2026-11-23] ingest | create notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs
- [create] [[notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs]] — ny sida från inboxen
- [archive] 2026-11-23-worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs.md


## [2026-11-23] link | notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs -> notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11


## [2026-11-23] link | notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs -> notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken


## [2026-11-23] ingest | create notes/omarchy-idle-lasning-stangs-av-via-shell-json-inte-hypridle-conf
- [create] [[notes/omarchy-idle-lasning-stangs-av-via-shell-json-inte-hypridle-conf]] — ny sida från inboxen
- [create] [[notes/stow-fallgropar-dotfiles-i-repos-dotfiles-stow-path-repos-dotfiles]] — ny sida från inboxen
- [create] [[notes/taggar-via-leader-t-pa-alla-malsidor]] — ny sida från inboxen
- [archive] 2026-11-23-omarchy-idle-lasning-stangs-av-via-shell-json-inte-hypridle-conf.md
- [archive] 2026-11-23-stow-fallgropar-dotfiles-i-repos-dotfiles-stow-path-repos-dotfiles.md
- [archive] 2026-11-23-taggar-via-leader-t-pa-alla-malsidor.md
## [2026-11-23] ingest | create notes/stow-fallgropar-dotfiles-i-repos-dotfiles-stow-path-repos-dotfiles
## [2026-11-23] ingest | create notes/taggar-via-leader-t-pa-alla-malsidor


## [2026-11-23] ingest | create notes/enter-skapar-alltid-nytt-mal-piltangenterna-markerar-befintligt-mal-i-sokfalt
- [create] [[notes/enter-skapar-alltid-nytt-mal-piltangenterna-markerar-befintligt-mal-i-sokfalt]] — ny sida från inboxen
- [archive] 2026-11-23-enter-skapar-alltid-nytt-mal-piltangenterna-markerar-befintligt-mal-i-sokfalt.md


## [2026-11-23] ingest | create notes/klustret-anvander-sops-via-flux-inte-vault-eso
- [create] [[notes/klustret-anvander-sops-via-flux-inte-vault-eso]] — ny sida från inboxen
- [create] [[notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23]] — ny sida från inboxen
- [archive] 2026-11-23-klustret-anvander-sops-via-flux-inte-vault-eso.md
- [archive] 2026-11-23-klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23.md
## [2026-11-23] ingest | create notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23


## [2026-11-23] ingest | create notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876
- [create] [[notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876]] — ny sida från inboxen
- [create] [[notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet]] — ny sida från inboxen
- [archive] 2026-11-23-agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876.md
- [archive] 2026-11-23-artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet.md
## [2026-11-23] ingest | create notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet


## [2026-11-23] link | summaries/llm-wiki-karpathy -> notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet


## [2026-11-23] link | notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev -> notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876


## [2026-11-23] link | notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876 -> notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev


## [2026-11-23] ingest | create notes/verifierade-policy-hal-i-tool-lagret-kommandon-ar-blocklist-inte-allowlist
- [create] [[notes/verifierade-policy-hal-i-tool-lagret-kommandon-ar-blocklist-inte-allowlist]] — ny sida från inboxen
- [archive] 2026-11-23-verifierade-policy-hal-i-tool-lagret-kommandon-ar-blocklist-inte-allowlist.md


## [2026-11-23] link | notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet -> notes/verifierade-policy-hal-i-tool-lagret-kommandon-ar-blocklist-inte-allowlist


## [2026-11-23] link | notes/verifierade-policy-hal-i-tool-lagret-kommandon-ar-blocklist-inte-allowlist -> notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet


## [2026-11-24] ingest | create notes/agentmal-flaggan-blev-levande-mvp
- [create] [[notes/agentmal-flaggan-blev-levande-mvp]] — ny sida från inboxen
- [create] [[notes/autonoma-agentruns-som-kubernetes-jobs-lardomar-fran-e2e-flodet]] — ny sida från inboxen
- [create] [[notes/chattinterface-byggt-i-plan-slack-borttaget]] — ny sida från inboxen
- [create] [[notes/inaktiva-ikon-rader-dolda-helt-ingen-graad-variant-vid-hover]] — ny sida från inboxen
- [create] [[notes/interaktiv-malchatt-behover-sakra-plan-verktyg]] — ny sida från inboxen
- [create] [[notes/malfaserna-ar-gemensamma-for-alla-mal-i-plan]] — ny sida från inboxen
- [archive] 2026-11-24-agentmal-flaggan-blev-levande-mvp.md
- [archive] 2026-11-24-autonoma-agentruns-som-kubernetes-jobs-lardomar-fran-e2e-flodet.md
- [archive] 2026-11-24-chattinterface-byggt-i-plan-slack-borttaget.md
- [archive] 2026-11-24-inaktiva-ikon-rader-dolda-helt-ingen-graad-variant-vid-hover.md
- [archive] 2026-11-24-interaktiv-malchatt-behover-sakra-plan-verktyg.md
- [archive] 2026-11-24-malfaserna-ar-gemensamma-for-alla-mal-i-plan.md
## [2026-11-24] ingest | create notes/autonoma-agentruns-som-kubernetes-jobs-lardomar-fran-e2e-flodet
## [2026-11-24] ingest | create notes/chattinterface-byggt-i-plan-slack-borttaget
## [2026-11-24] ingest | create notes/inaktiva-ikon-rader-dolda-helt-ingen-graad-variant-vid-hover
## [2026-11-24] ingest | create notes/interaktiv-malchatt-behover-sakra-plan-verktyg
## [2026-11-24] ingest | create notes/malfaserna-ar-gemensamma-for-alla-mal-i-plan


## [2026-11-24] ingest | create notes/agentprofiler-som-teknisk-policy-utanfor-prompten
- [create] [[notes/agentprofiler-som-teknisk-policy-utanfor-prompten]] — ny sida från inboxen
- [archive] 2026-11-24-agentprofiler-som-teknisk-policy-utanfor-prompten.md


## [2026-11-24] ingest | create notes/human-gate-som-teknisk-grans-scope-pa-api-nyckeln
- [create] [[notes/human-gate-som-teknisk-grans-scope-pa-api-nyckeln]] — ny sida från inboxen
- [archive] 2026-11-24-human-gate-som-teknisk-grans-scope-pa-api-nyckeln.md


## [2026-11-24] ingest | create notes/sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det
- [create] [[notes/sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det]] — ny sida från inboxen
- [archive] 2026-11-24-sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det.md


## [2026-11-24] ingest | create notes/slutverifiering-av-sammanslagen-plattform-mal-540-genom-hela-kedjan
- [create] [[notes/slutverifiering-av-sammanslagen-plattform-mal-540-genom-hela-kedjan]] — ny sida från inboxen
- [archive] 2026-11-24-slutverifiering-av-sammanslagen-plattform-mal-540-genom-hela-kedjan.md


## [2026-11-24] ingest | create notes/koda-korningar-ar-uttryckliga-begaran-och-en-avslutad-korning-har-alltid-run-completed
- [create] [[notes/koda-korningar-ar-uttryckliga-begaran-och-en-avslutad-korning-har-alltid-run-completed]] — ny sida från inboxen
- [archive] 2026-11-24-koda-korningar-ar-uttryckliga-begaran-och-en-avslutad-korning-har-alltid-run-completed.md


## [2026-11-24] ingest | create notes/jobb-ska-avslutas-nar-korningen-ar-klar-och-en-delad-proxy-rensar-typescript-felen
- [create] [[notes/jobb-ska-avslutas-nar-korningen-ar-klar-och-en-delad-proxy-rensar-typescript-felen]] — ny sida från inboxen
- [archive] 2026-11-24-jobb-ska-avslutas-nar-korningen-ar-klar-och-en-delad-proxy-rensar-typescript-felen.md


## [2026-11-24] ingest | create notes/kostnad-och-budget-for-agentkorningar
- [create] [[notes/kostnad-och-budget-for-agentkorningar]] — ny sida från inboxen
- [archive] 2026-11-24-kostnad-och-budget-for-agentkorningar.md


## [2026-11-24] ingest | create notes/ssh-nyckelatkomst-till-mac-mini-10-10-10-29-worker-9
- [create] [[notes/ssh-nyckelatkomst-till-mac-mini-10-10-10-29-worker-9]] — ny sida från inboxen
- [archive] 2026-11-24-ssh-nyckelatkomst-till-mac-mini-10-10-10-29-worker-9.md


## [2026-11-24] ingest | create notes/mac-mini-worker-9-har-en-bootbar-talos-usb-och-dog-under-inspektion-2026-11-24
- [create] [[notes/mac-mini-worker-9-har-en-bootbar-talos-usb-och-dog-under-inspektion-2026-11-24]] — ny sida från inboxen
- [archive] 2026-11-24-mac-mini-worker-9-har-en-bootbar-talos-usb-och-dog-under-inspektion-2026-11-24.md


## [2026-11-24] ingest | create notes/worker-9-mac-mini-bootade-talos-fran-usb-men-talos-api-t-50000-startar-inte-noden-ar-en-zombie
- [create] [[notes/worker-9-mac-mini-bootade-talos-fran-usb-men-talos-api-t-50000-startar-inte-noden-ar-en-zombie]] — ny sida från inboxen
- [archive] 2026-11-24-worker-9-mac-mini-bootade-talos-fran-usb-men-talos-api-t-50000-startar-inte-noden-ar-en-zombie.md


## [2026-11-24] ingest | create notes/mac-mini-crashar-varje-gang-man-monterar-den-externa-usb-stickan-msdosfs-2026-11-24
- [create] [[notes/mac-mini-crashar-varje-gang-man-monterar-den-externa-usb-stickan-msdosfs-2026-11-24]] — ny sida från inboxen
- [archive] 2026-11-24-mac-mini-crashar-varje-gang-man-monterar-den-externa-usb-stickan-msdosfs-2026-11-24.md


## [2026-11-24] ingest | create notes/crd-beroende-resurser-maste-synkas-i-ett-andra-steg
- [create] [[notes/crd-beroende-resurser-maste-synkas-i-ett-andra-steg]] — ny sida från inboxen
- [create] [[notes/flux-i-prod-klonade-ett-1-3-gb-stort-repo-och-dog]] — ny sida från inboxen
- [create] [[notes/kubeletts-logg-api-svarade-med-tls-internal-error-i-prod]] — ny sida från inboxen
- [archive] 2026-11-24-crd-beroende-resurser-maste-synkas-i-ett-andra-steg.md
- [archive] 2026-11-24-flux-i-prod-klonade-ett-1-3-gb-stort-repo-och-dog.md
- [archive] 2026-11-24-kubeletts-logg-api-svarade-med-tls-internal-error-i-prod.md
## [2026-11-24] ingest | create notes/flux-i-prod-klonade-ett-1-3-gb-stort-repo-och-dog
## [2026-11-24] ingest | create notes/kubeletts-logg-api-svarade-med-tls-internal-error-i-prod


## [2026-11-25] ingest | create notes/agents-repot-arkiverat-och-mergens-mal-kompletterade
- [create] [[notes/agents-repot-arkiverat-och-mergens-mal-kompletterade]] — ny sida från inboxen
- [create] [[notes/mac-mini-worker-9-kraschar-vid-tyngre-usb-i-o-pa-talos-stickan]] — ny sida från inboxen
- [create] [[notes/plan-goals-gav-503-go-api-startade-i-mock-lage-nar-dns-inte-var-uppe]] — ny sida från inboxen
- [archive] 2026-11-25-agents-repot-arkiverat-och-mergens-mal-kompletterade.md
- [archive] 2026-11-25-mac-mini-worker-9-kraschar-vid-tyngre-usb-i-o-pa-talos-stickan.md
- [archive] 2026-11-25-plan-goals-gav-503-go-api-startade-i-mock-lage-nar-dns-inte-var-uppe.md
## [2026-11-25] ingest | create notes/mac-mini-worker-9-kraschar-vid-tyngre-usb-i-o-pa-talos-stickan
## [2026-11-25] ingest | create notes/plan-goals-gav-503-go-api-startade-i-mock-lage-nar-dns-inte-var-uppe


## [2026-11-25] ingest | create notes/ci-gates-kanonisk-agentrun-livscykel-och-versionerat-plan-agent-kontrakt
- [create] [[notes/ci-gates-kanonisk-agentrun-livscykel-och-versionerat-plan-agent-kontrakt]] — ny sida från inboxen
- [archive] 2026-11-25-ci-gates-kanonisk-agentrun-livscykel-och-versionerat-plan-agent-kontrakt.md


## [2026-11-25] ingest | create notes/home-assistant-krokviken-ssh-via-terminal-ssh-addon
- [create] [[notes/home-assistant-krokviken-ssh-via-terminal-ssh-addon]] — ny sida från inboxen
- [archive] 2026-11-25-home-assistant-krokviken-ssh-via-terminal-ssh-addon.md


## [2026-11-25] page | entities/home-assistant


## [2026-11-25] link | notes/home-assistant-krokviken-ssh-via-terminal-ssh-addon -> entities/home-assistant


## [2026-11-25] link | notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02 -> entities/home-assistant


## [2026-11-25] ingest | create notes/ha-krokviken-robotdammsugare-startar-stoppar-pa-person-entiteter
- [create] [[notes/ha-krokviken-robotdammsugare-startar-stoppar-pa-person-entiteter]] — ny sida från inboxen
- [archive] 2026-11-25-ha-krokviken-robotdammsugare-startar-stoppar-pa-person-entiteter.md


## [2026-11-25] link | notes/ha-krokviken-robotdammsugare-startar-stoppar-pa-person-entiteter -> entities/home-assistant


## [2026-11-25] ingest | create notes/agent-scope-i-prod-allowlistan-maste-matcha-kontraktet
- [create] [[notes/agent-scope-i-prod-allowlistan-maste-matcha-kontraktet]] — ny sida från inboxen
- [create] [[notes/prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte]] — ny sida från inboxen
- [archive] 2026-11-25-agent-scope-i-prod-allowlistan-maste-matcha-kontraktet.md
- [archive] 2026-11-25-prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte.md
## [2026-11-25] ingest | create notes/prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte


## [2026-11-25] link | notes/agent-scope-i-prod-allowlistan-maste-matcha-kontraktet -> notes/ci-gates-kanonisk-agentrun-livscykel-och-versionerat-plan-agent-kontrakt


## [2026-11-25] link | notes/agent-scope-i-prod-allowlistan-maste-matcha-kontraktet -> concepts/kubernetes


## [2026-11-25] link | notes/prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte -> entities/kubernetes


## [2026-11-25] link | notes/prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte -> notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23


## [2026-11-25] link | notes/ci-gates-kanonisk-agentrun-livscykel-och-versionerat-plan-agent-kontrakt -> notes/agent-scope-i-prod-allowlistan-maste-matcha-kontraktet


## [2026-11-25] link | notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23 -> notes/prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte


## [2026-11-25] ingest | create notes/ha-tuya-tesvor-s6-anslutningen-fungerar-men-statusfalt-saknas
- [create] [[notes/ha-tuya-tesvor-s6-anslutningen-fungerar-men-statusfalt-saknas]] — ny sida från inboxen
- [archive] 2026-11-25-ha-tuya-tesvor-s6-anslutningen-fungerar-men-statusfalt-saknas.md


## [2026-11-25] link | notes/ha-tuya-tesvor-s6-anslutningen-fungerar-men-statusfalt-saknas -> entities/home-assistant


## [2026-11-25] ingest | create notes/aterstaende-funktioner-i-agentplattformen
- [create] [[notes/aterstaende-funktioner-i-agentplattformen]] — ny sida från inboxen
- [create] [[notes/ctrl-backspace-i-foot-skickar-ctrl-w-till-pi]] — ny sida från inboxen
- [create] [[notes/foot-bindningen-kraver-omstart-av-foot]] — ny sida från inboxen
- [archive] 2026-11-25-aterstaende-funktioner-i-agentplattformen.md
- [archive] 2026-11-25-ctrl-backspace-i-foot-skickar-ctrl-w-till-pi.md
- [archive] 2026-11-25-foot-bindningen-kraver-omstart-av-foot.md
## [2026-11-25] ingest | create notes/ctrl-backspace-i-foot-skickar-ctrl-w-till-pi
## [2026-11-25] ingest | create notes/foot-bindningen-kraver-omstart-av-foot


## [2026-11-25] ingest | create notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna
- [create] [[notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna]] — ny sida från inboxen
- [archive] 2026-11-25-kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna.md


## [2026-11-25] ingest | create notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard
- [create] [[notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard]] — ny sida från inboxen
- [archive] 2026-11-25-falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard.md


## [2026-11-25] link | notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna -> notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard


## [2026-11-25] link | notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard -> notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna


## [2026-11-25] ingest | create notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat
- [create] [[notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat]] — ny sida från inboxen
- [archive] 2026-11-25-oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat.md


## [2026-11-25] link | notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat -> notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna


## [2026-11-25] link | notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat -> notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard


## [2026-11-25] link | notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna -> notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat


## [2026-11-25] link | notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard -> notes/oidc-fix-jwt-konfiguration-applicerad-pa-controlplanes-och-zitadel-rbac-applicerat


## [2026-11-26] page | concepts/anvanda-agenter-i-plan


## [2026-11-26] link | concepts/anvanda-agenter-i-plan -> concepts/sops


## [2026-11-26] unlink | concepts/anvanda-agenter-i-plan -> concepts/sops


## [2026-11-26] link | concepts/anvanda-agenter-i-plan -> notes/human-gate-som-teknisk-grans-scope-pa-api-nyckeln


## [2026-11-26] page | concepts/anvanda-agenter-i-plan


## [2026-11-26] ingest | create notes/mal-540-slutfort-agentens-readme-andring-granskad-och-mergead
- [create] [[notes/mal-540-slutfort-agentens-readme-andring-granskad-och-mergead]] — ny sida från inboxen
- [create] [[notes/pr-preview-yaml-bygger-ogiltig-image-tagg-alla-pr-preview-floden-faller]] — ny sida från inboxen
- [archive] 2026-11-26-mal-540-slutfort-agentens-readme-andring-granskad-och-mergead.md
- [archive] 2026-11-26-pr-preview-yaml-bygger-ogiltig-image-tagg-alla-pr-preview-floden-faller.md
## [2026-11-26] ingest | create notes/pr-preview-yaml-bygger-ogiltig-image-tagg-alla-pr-preview-floden-faller


## [2026-11-26] ingest | create notes/mal-568-verifierad-tva-samtidiga-agentruns-i-dev
- [create] [[notes/mal-568-verifierad-tva-samtidiga-agentruns-i-dev]] — ny sida från inboxen
- [archive] 2026-11-26-mal-568-verifierad-tva-samtidiga-agentruns-i-dev.md


## [2026-11-26] link | notes/mal-568-verifierad-tva-samtidiga-agentruns-i-dev -> concepts/anvanda-agenter-i-plan


## [2026-11-26] link | notes/mal-568-verifierad-tva-samtidiga-agentruns-i-dev -> notes/aterstaende-funktioner-i-agentplattformen


## [2026-11-26] ingest | create notes/dokumentationssidor-i-granssnittet-docs
- [create] [[notes/dokumentationssidor-i-granssnittet-docs]] — ny sida från inboxen
- [archive] 2026-11-26-dokumentationssidor-i-granssnittet-docs.md


## [2026-11-26] link | concepts/anvanda-agenter-i-plan -> notes/dokumentationssidor-i-granssnittet-docs


## [2026-11-26] link | notes/arkitektur -> notes/dokumentationssidor-i-granssnittet-docs


## [2026-11-26] ingest | create notes/agent-dock-pa-malsiden-ersatter-de-tre-agentpanelerna
- [create] [[notes/agent-dock-pa-malsiden-ersatter-de-tre-agentpanelerna]] — ny sida från inboxen
- [archive] 2026-11-26-agent-dock-pa-malsiden-ersatter-de-tre-agentpanelerna.md


## [2026-11-26] ingest | create notes/pr-preview-floden-var-trasig-i-fem-steg-nu-fungerar-den
- [create] [[notes/pr-preview-floden-var-trasig-i-fem-steg-nu-fungerar-den]] — ny sida från inboxen
- [archive] 2026-11-26-pr-preview-floden-var-trasig-i-fem-steg-nu-fungerar-den.md


## [2026-11-26] link | notes/pr-preview-floden-var-trasig-i-fem-steg-nu-fungerar-den -> notes/pr-preview-yaml-bygger-ogiltig-image-tagg-alla-pr-preview-floden-faller


## [2026-11-26] ingest | create notes/aa-data-api-v2-data-pensionerad-dashboard-migrerar-till-api-v2-language-models
- [create] [[notes/aa-data-api-v2-data-pensionerad-dashboard-migrerar-till-api-v2-language-models]] — ny sida från inboxen
- [create] [[notes/controlplane-1-ateransluten-via-temporar-pxe-pa-controlplane-3]] — ny sida från inboxen
- [archive] 2026-11-26-aa-data-api-v2-data-pensionerad-dashboard-migrerar-till-api-v2-language-models.md
- [archive] 2026-11-26-controlplane-1-ateransluten-via-temporar-pxe-pa-controlplane-3.md
## [2026-11-26] ingest | create notes/controlplane-1-ateransluten-via-temporar-pxe-pa-controlplane-3


## [2026-11-26] ingest | create notes/src-strukturen-stadades-src-go-blev-src-api-kommandon-flyttade-till-src-agent-commands
- [create] [[notes/src-strukturen-stadades-src-go-blev-src-api-kommandon-flyttade-till-src-agent-commands]] — ny sida från inboxen
- [create] [[notes/worker-8-ater-ready-worker-1-och-worker-2-saknar-fysisk-lank]] — ny sida från inboxen
- [archive] 2026-11-26-src-strukturen-stadades-src-go-blev-src-api-kommandon-flyttade-till-src-agent-commands.md
- [archive] 2026-11-26-worker-8-ater-ready-worker-1-och-worker-2-saknar-fysisk-lank.md
## [2026-11-26] ingest | create notes/worker-8-ater-ready-worker-1-och-worker-2-saknar-fysisk-lank


## [2026-11-26] link | notes/src-strukturen-stadades-src-go-blev-src-api-kommandon-flyttade-till-src-agent-commands -> notes/sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det


## [2026-11-26] link | notes/src-strukturen-stadades-src-go-blev-src-api-kommandon-flyttade-till-src-agent-commands -> notes/ci-gates-kanonisk-agentrun-livscykel-och-versionerat-plan-agent-kontrakt


## [2026-11-26] link | notes/sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det -> notes/src-strukturen-stadades-src-go-blev-src-api-kommandon-flyttade-till-src-agent-commands


## [2026-11-26] ingest | create notes/dashboard-ranking-anvander-benchmark-matchade-kostnader
- [create] [[notes/dashboard-ranking-anvander-benchmark-matchade-kostnader]] — ny sida från inboxen
- [create] [[notes/tilt-dev-server-ska-koras-som-serve-cmd]] — ny sida från inboxen
- [archive] 2026-11-26-dashboard-ranking-anvander-benchmark-matchade-kostnader.md
- [archive] 2026-11-26-tilt-dev-server-ska-koras-som-serve-cmd.md
## [2026-11-26] ingest | create notes/tilt-dev-server-ska-koras-som-serve-cmd


## [2026-11-26] page | notes/dashboard-ranking-anvander-benchmark-matchade-kostnader


## [2026-11-26] page | notes/tilt-dev-server-ska-koras-som-serve-cmd


## [2026-11-26] link | notes/dashboard-ranking-anvander-benchmark-matchade-kostnader -> notes/aa-data-api-v2-data-pensionerad-dashboard-migrerar-till-api-v2-language-models


## [2026-11-26] link | notes/tilt-dev-server-ska-koras-som-serve-cmd -> notes/tilt-filbevakning-kan-do-starta-om-tilt-up


## [2026-11-26] link | notes/aa-data-api-v2-data-pensionerad-dashboard-migrerar-till-api-v2-language-models -> notes/dashboard-ranking-anvander-benchmark-matchade-kostnader


## [2026-11-26] link | notes/tilt-filbevakning-kan-do-starta-om-tilt-up -> notes/tilt-dev-server-ska-koras-som-serve-cmd


## [2026-11-26] page | notes/flux-i-prod-klonade-ett-1-3-gb-stort-repo-och-dog


## [2026-11-26] link | concepts/flux-operator -> notes/flux-i-prod-klonade-ett-1-3-gb-stort-repo-och-dog


## [2026-11-26] page | notes/dashboard-ranking-anvander-benchmark-matchade-kostnader


## [2026-11-26] ingest | create notes/dashboard-ranking-utokad-med-benchmark-och-kostnadstackning
- [create] [[notes/dashboard-ranking-utokad-med-benchmark-och-kostnadstackning]] — ny sida från inboxen
- [create] [[notes/flannel-vxlan-mac-kollision-blockerar-controlplane-2-webhooks]] — ny sida från inboxen
- [create] [[notes/flux-felsokning-avbruten-efter-exponerade-hemligheter]] — ny sida från inboxen
- [create] [[notes/plan-api-och-agent-delar-go-modul-men-behaller-tjanstegransen]] — ny sida från inboxen
- [archive] 2026-11-26-dashboard-ranking-utokad-med-benchmark-och-kostnadstackning.md
- [archive] 2026-11-26-flannel-vxlan-mac-kollision-blockerar-controlplane-2-webhooks.md
- [archive] 2026-11-26-flux-felsokning-avbruten-efter-exponerade-hemligheter.md
- [archive] 2026-11-26-plan-api-och-agent-delar-go-modul-men-behaller-tjanstegransen.md
## [2026-11-26] ingest | create notes/flannel-vxlan-mac-kollision-blockerar-controlplane-2-webhooks
## [2026-11-26] ingest | create notes/flux-felsokning-avbruten-efter-exponerade-hemligheter
## [2026-11-26] ingest | create notes/plan-api-och-agent-delar-go-modul-men-behaller-tjanstegransen


## [2026-11-26] ingest | create notes/dashboard-importerar-hela-aa-katalogen-766-modeller
- [create] [[notes/dashboard-importerar-hela-aa-katalogen-766-modeller]] — ny sida från inboxen
- [create] [[notes/sops-age-nyckelroterad-privat-nyckel-i-1password-item-sops]] — ny sida från inboxen
- [archive] 2026-11-26-dashboard-importerar-hela-aa-katalogen-766-modeller.md
- [archive] 2026-11-26-sops-age-nyckelroterad-privat-nyckel-i-1password-item-sops.md
## [2026-11-26] ingest | create notes/sops-age-nyckelroterad-privat-nyckel-i-1password-item-sops


## [2026-11-26] ingest | create notes/dashboard-gratismodeller-saknade-uppskattad-kostnad-och-knapparna-lovade-fel
- [create] [[notes/dashboard-gratismodeller-saknade-uppskattad-kostnad-och-knapparna-lovade-fel]] — ny sida från inboxen
- [create] [[notes/flux-source-controller-flyttad-fran-worker-3-efter-livenessfel]] — ny sida från inboxen
- [archive] 2026-11-26-dashboard-gratismodeller-saknade-uppskattad-kostnad-och-knapparna-lovade-fel.md
- [archive] 2026-11-26-flux-source-controller-flyttad-fran-worker-3-efter-livenessfel.md
## [2026-11-26] ingest | create notes/flux-source-controller-flyttad-fran-worker-3-efter-livenessfel


## [2026-11-26] ingest | create notes/dashboard-en-synk-knapp-i-stallet-for-tre-forvirrande
- [create] [[notes/dashboard-en-synk-knapp-i-stallet-for-tre-forvirrande]] — ny sida från inboxen
- [archive] 2026-11-26-dashboard-en-synk-knapp-i-stallet-for-tre-forvirrande.md


## [2026-11-26] ingest | create notes/aa-intelligence-index-api-och-webb-ar-samsta-aldre-index-finns-bara-i-krypterade-chartfiler
- [create] [[notes/aa-intelligence-index-api-och-webb-ar-samsta-aldre-index-finns-bara-i-krypterade-chartfiler]] — ny sida från inboxen
- [create] [[notes/bulkmarkering-av-mal]] — ny sida från inboxen
- [archive] 2026-11-26-aa-intelligence-index-api-och-webb-ar-samsta-aldre-index-finns-bara-i-krypterade-chartfiler.md
- [archive] 2026-11-26-bulkmarkering-av-mal.md
## [2026-11-26] ingest | create notes/bulkmarkering-av-mal


## [2026-11-26] ingest | create notes/bulkmarkering-pa-malsidor
- [create] [[notes/bulkmarkering-pa-malsidor]] — ny sida från inboxen
- [create] [[notes/dashboard-budget-finder-och-prenumerationsjamforelse]] — ny sida från inboxen
- [archive] 2026-11-26-bulkmarkering-pa-malsidor.md
- [archive] 2026-11-26-dashboard-budget-finder-och-prenumerationsjamforelse.md
## [2026-11-26] ingest | create notes/dashboard-budget-finder-och-prenumerationsjamforelse


## [2026-11-26] ingest | create notes/dashboard-tydlighetslyft-av-jamforelserna
- [create] [[notes/dashboard-tydlighetslyft-av-jamforelserna]] — ny sida från inboxen
- [archive] 2026-11-26-dashboard-tydlighetslyft-av-jamforelserna.md


## [2026-11-26] ingest | create notes/cluster1-longhorn-worker-4-aterstalld-och-flux-stabiliserad
- [create] [[notes/cluster1-longhorn-worker-4-aterstalld-och-flux-stabiliserad]] — ny sida från inboxen
- [create] [[notes/dashboard-rekommendationsvyn-som-enda-forstavy]] — ny sida från inboxen
- [create] [[notes/tillfalliga-felsokningspoddar-rensade]] — ny sida från inboxen
- [archive] 2026-11-26-cluster1-longhorn-worker-4-aterstalld-och-flux-stabiliserad.md
- [archive] 2026-11-26-dashboard-rekommendationsvyn-som-enda-forstavy.md
- [archive] 2026-11-26-tillfalliga-felsokningspoddar-rensade.md
## [2026-11-26] ingest | create notes/dashboard-rekommendationsvyn-som-enda-forstavy
## [2026-11-26] ingest | create notes/tillfalliga-felsokningspoddar-rensade


## [2026-11-26] ingest | create notes/dashboard-plans-visar-och-raknar-anvandningskvoter
- [create] [[notes/dashboard-plans-visar-och-raknar-anvandningskvoter]] — ny sida från inboxen
- [create] [[notes/minimax-m-plan-inlagd-i-jamforelsen-m2-7-syns-som-uppskattad]] — ny sida från inboxen
- [create] [[notes/talos-certifikat-kortlivade-servercertifikat-roteras-automatiskt]] — ny sida från inboxen
- [create] [[notes/varfor-src-har-tva-cmd-kataloger-go-internal-regeln-sparrar-gemensam-cmd]] — ny sida från inboxen
- [archive] 2026-11-26-dashboard-plans-visar-och-raknar-anvandningskvoter.md
- [archive] 2026-11-26-minimax-m-plan-inlagd-i-jamforelsen-m2-7-syns-som-uppskattad.md
- [archive] 2026-11-26-talos-certifikat-kortlivade-servercertifikat-roteras-automatiskt.md
- [archive] 2026-11-26-varfor-src-har-tva-cmd-kataloger-go-internal-regeln-sparrar-gemensam-cmd.md
## [2026-11-26] ingest | create notes/minimax-m-plan-inlagd-i-jamforelsen-m2-7-syns-som-uppskattad
## [2026-11-26] ingest | create notes/talos-certifikat-kortlivade-servercertifikat-roteras-automatiskt
## [2026-11-26] ingest | create notes/varfor-src-har-tva-cmd-kataloger-go-internal-regeln-sparrar-gemensam-cmd


## [2026-11-26] ingest | create notes/envoy-vip-10-10-10-101-stulen-av-agentgateway-admin-plan-simonbrundin-com-onabar
- [create] [[notes/envoy-vip-10-10-10-101-stulen-av-agentgateway-admin-plan-simonbrundin-com-onabar]] — ny sida från inboxen
- [archive] 2026-11-26-envoy-vip-10-10-10-101-stulen-av-agentgateway-admin-plan-simonbrundin-com-onabar.md


## [2026-11-26] ingest | create notes/cloudflare-tunnel-aktiverat-plan-simonbrundin-com-publikt-nabar
- [create] [[notes/cloudflare-tunnel-aktiverat-plan-simonbrundin-com-publikt-nabar]] — ny sida från inboxen
- [archive] 2026-11-26-cloudflare-tunnel-aktiverat-plan-simonbrundin-com-publikt-nabar.md


## [2026-11-26] ingest | create notes/wildcard-dns-via-cloudflare-tunnel-alla-simonbrundin-com-gar-genom-tunneln-till-envoy
- [create] [[notes/wildcard-dns-via-cloudflare-tunnel-alla-simonbrundin-com-gar-genom-tunneln-till-envoy]] — ny sida från inboxen
- [archive] 2026-11-26-wildcard-dns-via-cloudflare-tunnel-alla-simonbrundin-com-gar-genom-tunneln-till-envoy.md


## [2026-11-26] ingest | create notes/platsvaljaren-flyttade-till-leader-p
- [create] [[notes/platsvaljaren-flyttade-till-leader-p]] — ny sida från inboxen
- [archive] 2026-11-26-platsvaljaren-flyttade-till-leader-p.md


## [2026-11-26] ingest | create notes/kvotmodell-anropskvoter-och-mjuka-granser-kan-inte-raknas-till-dollar
- [create] [[notes/kvotmodell-anropskvoter-och-mjuka-granser-kan-inte-raknas-till-dollar]] — ny sida från inboxen
- [archive] 2026-11-26-kvotmodell-anropskvoter-och-mjuka-granser-kan-inte-raknas-till-dollar.md


## [2026-11-26] ingest | create notes/enhet-per-lyckad-uppgift-prenumerationer-kostar-per-forsok-och-maste-omraknas-med-pass-rate
- [create] [[notes/enhet-per-lyckad-uppgift-prenumerationer-kostar-per-forsok-och-maste-omraknas-med-pass-rate]] — ny sida från inboxen
- [archive] 2026-11-26-enhet-per-lyckad-uppgift-prenumerationer-kostar-per-forsok-och-maste-omraknas-med-pass-rate.md


## [2026-11-26] ingest | create notes/agentens-maldefinition-gar-via-human-gate
- [create] [[notes/agentens-maldefinition-gar-via-human-gate]] — ny sida från inboxen
- [archive] 2026-11-26-agentens-maldefinition-gar-via-human-gate.md


## [2026-11-26] ingest | create notes/anvandningsgranser-ar-tier-bundna-och-ska-aldrig-visas-utan-sin-tier
- [create] [[notes/anvandningsgranser-ar-tier-bundna-och-ska-aldrig-visas-utan-sin-tier]] — ny sida från inboxen
- [archive] 2026-11-26-anvandningsgranser-ar-tier-bundna-och-ska-aldrig-visas-utan-sin-tier.md


## [2026-11-26] ingest | create notes/plan-migreras-fran-go-till-rust
- [create] [[notes/plan-migreras-fran-go-till-rust]] — ny sida från inboxen
- [archive] 2026-11-26-plan-migreras-fran-go-till-rust.md


## [2026-11-26] ingest | create notes/paritetstest-gatear-rust-migreringen-mot-go-kallan
- [create] [[notes/paritetstest-gatear-rust-migreringen-mot-go-kallan]] — ny sida från inboxen
- [archive] 2026-11-26-paritetstest-gatear-rust-migreringen-mot-go-kallan.md


## [2026-11-26] ingest | create notes/ett-antagande-far-aldrig-vinna-en-jamforelse-mot-en-matning
- [create] [[notes/ett-antagande-far-aldrig-vinna-en-jamforelse-mot-en-matning]] — ny sida från inboxen
- [archive] 2026-11-26-ett-antagande-far-aldrig-vinna-en-jamforelse-mot-en-matning.md


## [2026-11-26] ingest | create notes/manadskostnad-raknas-pa-korningar-inte-pa-losta-uppgifter
- [create] [[notes/manadskostnad-raknas-pa-korningar-inte-pa-losta-uppgifter]] — ny sida från inboxen
- [archive] 2026-11-26-manadskostnad-raknas-pa-korningar-inte-pa-losta-uppgifter.md


## [2026-11-26] ingest | create notes/rust-api-t-far-auth-lager-och-forsta-domanerna
- [create] [[notes/rust-api-t-far-auth-lager-och-forsta-domanerna]] — ny sida från inboxen
- [archive] 2026-11-26-rust-api-t-far-auth-lager-och-forsta-domanerna.md


## [2026-11-26] ingest | create notes/pi-extension-minimax-quota-guard-vantar-ut-5h-fonstret-och-visar-pace
- [create] [[notes/pi-extension-minimax-quota-guard-vantar-ut-5h-fonstret-och-visar-pace]] — ny sida från inboxen
- [archive] 2026-11-26-pi-extension-minimax-quota-guard-vantar-ut-5h-fonstret-och-visar-pace.md


## [2026-11-26] link | notes/pi-extension-minimax-quota-guard-vantar-ut-5h-fonstret-och-visar-pace -> notes/pi-durable-fungerar-med-minimax-coding-plan-anthropic-messages-api


## [2026-11-26] ingest | create notes/malens-id-n-ar-integer-och-tysta-sqlx-fel-ar-farliga
- [create] [[notes/malens-id-n-ar-integer-och-tysta-sqlx-fel-ar-farliga]] — ny sida från inboxen
- [archive] 2026-11-26-malens-id-n-ar-integer-och-tysta-sqlx-fel-ar-farliga.md


## [2026-11-26] ingest | create notes/planpris-delat-per-korning-gar-inte-att-jamfora-mellan-planer
- [create] [[notes/planpris-delat-per-korning-gar-inte-att-jamfora-mellan-planer]] — ny sida från inboxen
- [archive] 2026-11-26-planpris-delat-per-korning-gar-inte-att-jamfora-mellan-planer.md


## [2026-11-26] ingest | create notes/worker-1-ansluter-till-cluster1-2026-11-26-install-config-applicerad-via-pxe-boot-och-noden-gick-med
- [create] [[notes/worker-1-ansluter-till-cluster1-2026-11-26-install-config-applicerad-via-pxe-boot-och-noden-gick-med]] — ny sida från inboxen
- [archive] 2026-11-26-worker-1-ansluter-till-cluster1-2026-11-26-install-config-applicerad-via-pxe-boot-och-noden-gick-med.md


## [2026-11-26] ingest | create notes/worker-1-longhorn-anvager-ephemeral-disken-inte-usb-ssd-n-och-alla-workers-far-samma-behandling
- [create] [[notes/worker-1-longhorn-anvager-ephemeral-disken-inte-usb-ssd-n-och-alla-workers-far-samma-behandling]] — ny sida från inboxen
- [create] [[notes/worker-2-och-worker-9-saknas-fortfarande-som-cluster1-workers-diagnos-2026-11-26]] — ny sida från inboxen
- [archive] 2026-11-26-worker-1-longhorn-anvager-ephemeral-disken-inte-usb-ssd-n-och-alla-workers-far-samma-behandling.md
- [archive] 2026-11-26-worker-2-och-worker-9-saknas-fortfarande-som-cluster1-workers-diagnos-2026-11-26.md
## [2026-11-26] ingest | create notes/worker-2-och-worker-9-saknas-fortfarande-som-cluster1-workers-diagnos-2026-11-26


## [2026-11-26] link | notes/worker-1-longhorn-anvager-ephemeral-disken-inte-usb-ssd-n-och-alla-workers-far-samma-behandling -> concepts/longhorn


## [2026-11-26] link | notes/worker-1-ansluter-till-cluster1-2026-11-26-install-config-applicerad-via-pxe-boot-och-noden-gick-med -> concepts/longhorn


## [2026-11-26] source | 2026-11-26-pinchbench-verifierade-api-fakta-och-datafel-i-value-vyn.md


## [2026-11-26] ingest | create notes/usb-ssd-orna-har-kvar-giltig-longhorn-identitet-ssd-1-med-gamla-replikor-de-ar-bara-inte-registrerade-i-longhorn
- [create] [[notes/usb-ssd-orna-har-kvar-giltig-longhorn-identitet-ssd-1-med-gamla-replikor-de-ar-bara-inte-registrerade-i-longhorn]] — ny sida från inboxen
- [archive] 2026-11-26-usb-ssd-orna-har-kvar-giltig-longhorn-identitet-ssd-1-med-gamla-replikor-de-ar-bara-inte-registrerade-i-longhorn.md


## [2026-11-26] ingest | create notes/pinchbench-egen-value-score-ska-inte-kopieras-rakna-kostnad-per-lyckad-uppgift-sjalv
- [create] [[notes/pinchbench-egen-value-score-ska-inte-kopieras-rakna-kostnad-per-lyckad-uppgift-sjalv]] — ny sida från inboxen
- [archive] 2026-11-26-pinchbench-egen-value-score-ska-inte-kopieras-rakna-kostnad-per-lyckad-uppgift-sjalv.md


## [2026-11-26] link | notes/usb-ssd-orna-har-kvar-giltig-longhorn-identitet-ssd-1-med-gamla-replikor-de-ar-bara-inte-registrerade-i-longhorn -> concepts/longhorn


## [2026-11-26] link | notes/usb-ssd-orna-har-kvar-giltig-longhorn-identitet-ssd-1-med-gamla-replikor-de-ar-bara-inte-registrerade-i-longhorn -> notes/worker-1-longhorn-anvager-ephemeral-disken-inte-usb-ssd-n-och-alla-workers-far-samma-behandling


## [2026-11-26] page | summaries/pinchbench-verifierade-api-fakta-och-datafel-i-value-vyn


## [2026-11-26] link | summaries/pinchbench-verifierade-api-fakta-och-datafel-i-value-vyn -> notes/pinchbench-egen-value-score-ska-inte-kopieras-rakna-kostnad-per-lyckad-uppgift-sjalv


## [2026-11-26] link | notes/pinchbench-egen-value-score-ska-inte-kopieras-rakna-kostnad-per-lyckad-uppgift-sjalv -> summaries/pinchbench-verifierade-api-fakta-och-datafel-i-value-vyn


## [2026-11-26] ingest | create notes/rrule-motor-i-rust-ersatter-rrule-go
- [create] [[notes/rrule-motor-i-rust-ersatter-rrule-go]] — ny sida från inboxen
- [archive] 2026-11-26-rrule-motor-i-rust-ersatter-rrule-go.md


## [2026-11-26] ingest | create notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor
- [create] [[notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor]] — ny sida från inboxen
- [create] [[notes/zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export]] — ny sida från inboxen
- [create] [[notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om]] — ny sida från inboxen
- [create] [[notes/zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in]] — ny sida från inboxen
- [archive] 2026-11-26-plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor.md
- [archive] 2026-11-26-zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export.md
- [archive] 2026-11-26-zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om.md
- [archive] 2026-11-26-zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in.md
## [2026-11-26] ingest | create notes/zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export
## [2026-11-26] ingest | create notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om
## [2026-11-26] ingest | create notes/zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in


## [2026-11-26] link | notes/zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export -> concepts/zitadel


## [2026-11-26] link | concepts/zitadel -> notes/zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export


## [2026-11-26] link | notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om -> concepts/zitadel


## [2026-11-26] link | concepts/zitadel -> notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om


## [2026-11-26] link | notes/zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in -> concepts/zitadel


## [2026-11-26] link | concepts/zitadel -> notes/zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in


## [2026-11-26] link | notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor -> concepts/cloudnative-pg


## [2026-11-26] link | concepts/cloudnative-pg -> notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor


## [2026-11-26] ingest | create notes/fyra-paritetsfallor-i-agentdomanens-port-till-rust
- [create] [[notes/fyra-paritetsfallor-i-agentdomanens-port-till-rust]] — ny sida från inboxen
- [archive] 2026-11-26-fyra-paritetsfallor-i-agentdomanens-port-till-rust.md


## [2026-11-26] ingest | create notes/malkompletteringar-och-auth-i-rust-de-sista-api-vagarna
- [create] [[notes/malkompletteringar-och-auth-i-rust-de-sista-api-vagarna]] — ny sida från inboxen
- [archive] 2026-11-26-malkompletteringar-och-auth-i-rust-de-sista-api-vagarna.md


## [2026-11-27] ingest | create notes/kunskaps-wikin-och-dess-verktyg-har-portats-till-rust
- [create] [[notes/kunskaps-wikin-och-dess-verktyg-har-portats-till-rust]] — ny sida från inboxen
- [archive] 2026-11-27-kunskaps-wikin-och-dess-verktyg-har-portats-till-rust.md
- [merge?] notes/kunskaps-wiki + notes/kunskaps-wikin-och-dess-verktyg-har-portats-till-rust (delade taggar: wiki, agent, verktyg)

## [2026-11-27] ingest | create notes/neon-anslutningen-hanger-nar-ipv6-prioriteras
- [create] [[notes/neon-anslutningen-hanger-nar-ipv6-prioriteras]] — ny sida från inboxen
- [create] [[notes/skrivbordswidget-som-visar-gor-nu-malet]] — ny sida från inboxen
- [archive] 2026-11-27-neon-anslutningen-hanger-nar-ipv6-prioriteras.md
- [archive] 2026-11-27-skrivbordswidget-som-visar-gor-nu-malet.md
## [2026-11-27] ingest | create notes/skrivbordswidget-som-visar-gor-nu-malet


## [2026-11-27] link | notes/skrivbordswidget-som-visar-gor-nu-malet -> concepts/anvanda-agenter-i-plan


## [2026-11-27] link | notes/neon-anslutningen-hanger-nar-ipv6-prioriteras -> notes/skrivbordswidget-som-visar-gor-nu-malet


## [2026-11-27] link | notes/skrivbordswidget-som-visar-gor-nu-malet -> notes/neon-anslutningen-hanger-nar-ipv6-prioriteras


## [2026-11-27] ingest | create notes/flyttalsparitet-och-orchestratorns-doda-gate-flode
- [create] [[notes/flyttalsparitet-och-orchestratorns-doda-gate-flode]] — ny sida från inboxen
- [archive] 2026-11-27-flyttalsparitet-och-orchestratorns-doda-gate-flode.md


## [2026-11-27] link | notes/flyttalsparitet-och-orchestratorns-doda-gate-flode -> notes/kunskaps-wiki


## [2026-11-27] link | notes/flyttalsparitet-och-orchestratorns-doda-gate-flode -> concepts/anvanda-agenter-i-plan


## [2026-11-27] link | notes/kunskaps-wiki -> notes/flyttalsparitet-och-orchestratorns-doda-gate-flode

## [2026-11-27] source | 2026-11-27-prod-patch-utan-namespace-bryter-flux-builden.md

## [2026-11-27] ingest | create notes/prod-patch-utan-namespace-bryter-flux-builden
- [create] [[notes/prod-patch-utan-namespace-bryter-flux-builden]] — ny sida från inboxen
- [archive] 2026-11-27-prod-patch-utan-namespace-bryter-flux-builden.md

## [2026-11-27] link | summaries/prod-patch-utan-namespace-bryter-flux-builden -> notes/prod-patch-utan-namespace-bryter-flux-builden
