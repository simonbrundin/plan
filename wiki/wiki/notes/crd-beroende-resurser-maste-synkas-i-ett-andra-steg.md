---
title: CRD-beroende resurser måste synkas i ett andra steg
type: note
tags: [flux, kustomize, gateway-api, projekt:infrastructure]
created: 2026-11-24
updated: 2026-11-24
status: active
---

Kustomex och Flux kan inte installera en CRD och använda den i samma synkning. Flux gör en server-side apply med dry-run av hela kustomizationen, och en resurs av okänd sorts (t.ex. GatewayClass, HTTPRoute, SecurityPolicy, AgentgatewayModel) får hela synkningen att fallera. Det går alltså inte att installera Envoy Gateway-charten och samtidigt applicera Gateway API-resurser.

Lösningen i infrastrukturrepot: Gateway API-resurserna ligger nu i infrastructure-flux/gateway/ och synkas av en egen FluxKustomization (apps/gateway) som har dependsOn på rotkustomizationen. Rotkustomizationen installerar charten och blir Ready först när HelmRelease är frisk, och då finns CRD:erna. Andra steget applicerar sedan GatewayClass, Gateway, HTTPRoute, ReferenceGrant, SecurityPolicy och AgentgatewayModel.

Samma regel gäller alla CRD:er från en chart som också har egna resurser i samma kustomization. Här gällde det agentgateway-charten (AgentgatewayModel och AgentgatewayParameters).
