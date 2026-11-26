# Plan för Kubernetes Problem

Skapa en ny plan i markdown med det specificerade `Plan Format`. Följ
`Instruktioner` för att skapa planen och använd `Relevanta filer` för att
fokusera på rätt filer.

## Instruktioner

- **VIKTIGT**: Du skriver en plan för hur du ska felsöka och åtgärda ett problem
  i Kubernetes. Du skriver alltså en plan för hur problemet ska felsökas och
  åtgärdas, inte den faktiska lösningen. Planen ska vara skriven i
  `Plan Format`.
- VIKTIGT: Skapa en markdown fil för planen i
  `.agent/plans/<beskrivande-namn-på-problemet>.md`

### Fas 1: Identifiera symptom

1. Dokumentera exakt vad som inte fungerar
2. Fråga: Vad ser användaren?
   - Webbsidan svarar inte
   - Pod startar inte
   - PVC fastnar i Pending
   - Ingress ger 404
   - Certifikat utfärdas inte
   - DNS fungerar inte

### Fas 2: Hitta första trasiga resursen

Undersök resurskedjan uppifrån och ner:

```
User
  ↓
Ingress/Gateway
  ↓
HTTPRoute
  ↓
Service
  ↓
Endpoints
  ↓
Pod
  ↓
Container
```

Kör `kubectl get` och `kubectl describe` på resurser i denna ordning. Första
resursen som inte ser korrekt ut är ofta där problemet börjar.

### Fas 3: Klassificera problemet efter lager

Identifiera vilket lager problemet tillhör:

| Lager          | Symptom                           | Vanliga orsaker                                 |
| -------------- | --------------------------------- | ----------------------------------------------- |
| **Scheduling** | Pod Pending                       | CPU/RAM slut, NodeSelector, Affinity, Taints    |
| **Container**  | Pod Running men app fungerar inte | Fel env vars, fel config, databas saknas        |
| **Nätverk**    | Kan inte nå app                   | Fel selector, fel port, DNS, TLS, NetworkPolicy |
| **Lagring**    | PVC Pending                       | StorageClass saknas, provisioner fungerar inte  |
| **GitOps**     | Manifest deployas inte            | Flux/ArgoCD reconciliation misslyckas           |
| **Certifikat** | HTTPS fungerar inte               | cert-manager problem                            |

### Fas 4: Hitta Root Cause (5 Varför)

Fråga "Varför?" tills du når grundorsaken:

```
Pod Pending
  ↓ Varför?
PVC Pending
  ↓ Varför?
StorageClass saknas
  ↓ Varför?
Fel namn i manifest
  ↓ Varför?
Root Cause hittad ✓
```

### Fas 5: Verifiera hypotes

Innan ändring, formulera en hypotes:

> "Om jag ändrar X, förväntar jag mig att Y händer."

### Fas 6: Ställ alltid dessa frågor

1. Vilken resurs är första som är trasig?
2. Vilken controller äger resursen?
3. Vad säger Events?
4. Vad säger Logs?
5. Vad är root cause?

### Fas 7: Implementera minsta möjliga fix

Undvik:

- Restarta allt
- Ta bort namespace
- Installera om

Föredra:

- Minsta ändring
- Tydlig rollback
- En komponent åt gången

### Fas 8: Verifiera hela kedjan

```
Git (desired state)
  ↓
Kubernetes Resource
  ↓
Pod
  ↓
Service
  ↓
Gateway
  ↓
User
```

## Relevanta filer

### Infrastructure components

Flux-hanterade Kubernetes-resurser:

```
~/repos/infrastructure/infrastructure-flux/
├── apps/                        # Flux-applikationer
├── clusters/production/         # Produktionskluster-konfiguration
├── components/                  # Återanvändbara komponenter
│   ├── cert-manager/            # Certifikathantering
│   ├── cloudnative-pg/           # PostgreSQL (CNPG)
│   ├── flux/                     # Flux bootstrap & konfiguration
│   ├── kube-prometheus-stack/   # Monitoring & logging
│   ├── longhorn-system/         # Blocklagring
│   ├── metallb/                  # Lastbalansering
│   └── secrets/                 # SOPS-krypterade hemligheter
└── flux-system/                 # Flux systemkonfiguration
```

### Vanliga felsökningsresurser

| Resurs            | Kommando                             | Lager                |
| ----------------- | ------------------------------------ | -------------------- |
| Gateway           | `kubectl get gateway -A`             | Nätverk              |
| HTTPRoute         | `kubectl get httproute -A`           | Nätverk              |
| Ingress           | `kubectl get ingress -A`             | Nätverk              |
| Service           | `kubectl get svc -A`                 | Nätverk              |
| Endpoints         | `kubectl get endpoints -A`           | Nätverk              |
| Pod               | `kubectl get pods -A`                | Scheduling/Container |
| PVC               | `kubectl get pvc -A`                 | Lagring              |
| Kustomization     | `kubectl get kustomization -A`       | GitOps               |
| HelmRelease       | `kubectl get hr -A`                  | GitOps               |
| Certificate       | `kubectl get certificate -A`         | Certifikat           |
| Volume (Longhorn) | `kubectl get volumes.longhorn.io -A` | Lagring              |

### Checka relevanta manifest

1. Läs `/home/simon/repos/infrastructure/README.md` för övergripande kontext
2. Leta i `~/repos/infrastructure/infrastructure-flux/components/` efter
   relevant component
3. Använd `flux get all` för att se Flux-synkstatus

---

## Planformat

```md
# Kubernetes Issue: <beskrivande-namn>

## Metadata

issue_number: `{issue_number}` adw_id: `{adw_id}` issue_json: `{issue_json}`

## Symptombeskrivning

<Vad ser användaren? Dokumentera exakt symptom.>

## Fas 2: Första trasiga resurs

<Ange vilken resurs som är första trasig i kedjan och vad som är fel med den.>

### Resurskedja (uppifrån och ner)

- [ ] User
- [ ] Gateway: `<status>`
- [ ] HTTPRoute: `<status>`
- [ ] Ingress: `<status>`
- [ ] Service: `<status>`
- [ ] Endpoints: `<status>`
- [ ] Pod: `<status>`

## Fas 3: Problemlager

<Ange vilket lager problemet tillhör: Scheduling / Container / Nätverk / Lagring
/ GitOps / Certifikat>

### Symptom

<Det exakta symtomet>

### Vanliga orsaker att undersöka

<Checkbox-lista med undersökta orsaker>

- [ ] Orsak 1
- [ ] Orsak 2

## Fas 4: Root Cause (5 Varför)

<Utför 5 Varför-analysen tills grundorsaken hittas.>

Fråga: <problem> ↓ Varför? Svar: <orsak> ↓ Varför? Svar: <orsak> ↓ Varför? Svar:
<orsak> ↓ Varför? Root Cause: <grundorsak>

## Fas 5: Hypotes

<Innan ändring: "Om jag ändrar X, förväntar jag mig att Y händer.">

## Lösningsbeskrivning

<Beskriv det föreslagna lösningssättet för att åtgärda problemet.>

## Relevanta filer

<Lista filerna som är relevanta för att åtgärda problemet.>

### Befintliga filer

| Fil        | Anledning           |
| ---------- | ------------------- |
| `<sökväg>` | `<varför relevant>` |

### Nya filer (om några)

| Fil        | Syfte                  |
| ---------- | ---------------------- |
| `<sökväg>` | `<vad filen ska göra>` |

## Steg-för-steg-uppgifter

### Steg 1: Undersök aktuell status

<Lista kubectl-kommandon för att verifiera problemet.>

### Steg 2: Verifiera hypotes

<Lista steg för att verifiera att hypotesen är korrekt.>

### Steg 3: Implementera fix

<Lista konkreta ändringar som ska göras.>

### Steg 4: Verifiera hela kedjan

Git → Kubernetes → Pod → Service → Gateway → User

<Lista verifieringssteg för varje del av kedjan.>

## Valideringskommandon

<Lista kubectl-kommandon och andra kommandon för att verifiera att problemet är
löst.>

- `kubectl get <resurs> -A` - Verifiera att resursen är korrekt
- `kubectl describe <resurs> <namn>` - Verifiera events
- `kubectl logs <pod>` - Verifiera apploggar
- `flux get all` - Verifiera Flux-synkstatus

## Rollback-plan

<Vad ska göras om fixen inte fungerar?>

## Anteckningar

<Valfritt: ytterligare anteckningar eller kontext.>
```

## Problemet

Problemet är det första argumentet `$1` som användes när du körde igång.

## Rapport

- **VIKTIGT**: Returnera uteslutande sökvägen till den skapade planfilen och
  inget annat. `.agent/plans/<beskrivande-namn-på-problemet>.md`
