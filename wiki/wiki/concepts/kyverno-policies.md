---
title: Kyverno Policies
type: concept
summary: Kyverno är policy-engine som blockerar osäkra resurser i klustret - NodePort, latest-tag, skyddade namespaces och saknade resursgränser.
tags: [kyverno, policy, kubernetes, security, validation]
sources: [infrastructure-flux/components/kyverno/]
created: 2026-06-03
updated: 2026-10-24
---

# Kyverno Policies

## Definition

Kyverno är en policy-engine för Kubernetes som möjliggör deklarativ
policyhantering utan separat programspråk. Policyerna ligger i
`infrastructure-flux/components/kyverno/policies/` och tillämpas av
Kyverno i namespace `kyverno` (chart version 3.8.1).

Varje policy är ett `ClusterPolicy` med
`validationFailureAction: Enforce` eller `Audit`.

## Aktiva policies

| Policy | Action | Regel |
|--------|--------|-------|
| `block-nodeport` | Enforce | Services av typ NodePort tillåts inte - använd ClusterIP eller LoadBalancer |
| `disallow-latest-tag` | Enforce | Images får inte använda `:latest` - använd specifik tagg eller digest |
| `protect-system-namespaces` | Enforce | `longhorn-system`, `kube-system` och `kyverno` får inte raderas |
| `require-requests-limits` | Audit | Podar bör ha CPU/memory requests och memory limits |

### Blockera NodePort

```yaml
spec:
  validationFailureAction: Enforce
  background: true
  rules:
    - name: validate-nodeport
      match:
        any:
          - resources:
              kinds:
                - Service
      validate:
        message: >
          Services of type NodePort are not allowed.
          Use ClusterIP or LoadBalancer instead.
        cel:
          expressions:
            - expression: "has(object.spec) && object.spec.type != 'NodePort'"
```

### Skydda system-namespaces

Kräver både `preconditions` (att det är en DELETE) och ett CEL-uttryck
som alltid är falskt:

```yaml
      preconditions:
        any:
          - key: "{{ request.operation }}"
            operator: In
            value:
              - "DELETE"
      validate:
        message: "Deletion of the longhorn-system namespace is not allowed..."
        cel:
          expressions:
            - expression: "false"
              message: "The longhorn-system namespace cannot be deleted."
```

## CEL-mönstret

Kyverno 3.x stöder CEL inuti `validate.cel.expressions` - det är
standardvalet för nya policies i denna infrastruktur. Se
[[concepts/kyverno-cel]].

## Felsökning

```bash
# Vilka policies finns?
kubectl get clusterpolicies

# Vad blockerar en resurs?
kubectl describe clusterpolicy block-nodeport

# Kyverno-loggen
kubectl logs -n kyverno -l app.kubernetes.io/component=admission-controller
```

## Kopplingar

- [[concepts/kyverno-cel]]
- [[entities/kubernetes]]
- [[concepts/flux-operator]]
