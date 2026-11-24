---
title: Kyverno CEL Policies
type: concept
summary: CEL (Common Expression Language) är standardvalet för alla nya Kyverno-policies i denna infrastruktur.
tags: [kyverno, policy, cel, kubernetes, security]
sources: [infrastructure-flux/components/kyverno/policies/]
created: 2026-06-03
updated: 2026-10-24
---

# Kyverno CEL Policies

CEL (Common Expression Language) är standardvalet för alla nya
Kyverno-policies i denna infrastruktur.

## Varför CEL?

- **Bättre prestanda**: CEL-uttryck kompileras en gång och körs snabbare
- **Standardiserat**: CEL är ett CNCF-projekt och vanligt i Kubernetes-ekosystemet
- **Lättläst**: Mer explicita villkor jämfört med pattern matching
- **Bättre felsökning**: Tydligare felmeddelanden vid valideringsfel

## Rätt syntax

I Kyverno 3.x heter fältet `cel` (inte `CEL`) och listan
`expressions` (inte `conditions`):

```yaml
      validate:
        message: >
          Services of type NodePort are not allowed.
          Use ClusterIP or LoadBalancer instead.
        cel:
          expressions:
            - expression: "has(object.spec) && object.spec.type != 'NodePort'"
```

### Pattern (fortfarande användbart för mönstermatchning)

```yaml
      validate:
        pattern:
          spec:
            containers:
              - image: "!*:latest"
```

Pattern används fortfarande av `require-requests-limits`, där mönstret
är enklare att uttrycka än ett CEL-uttryck.

## Regler vid skapande av nya policies

1. **Använd CEL** (`validate.cel.expressions`) för logik, pattern för
   struktur
2. **Använd `validationFailureAction: Enforce`** för säkerhetskritiska
   policies, `Audit` för rekommendationer
3. **Använd `background: true`** för att validera befintliga resurser
4. **Lägg till annotations** (`policies.kyverno.io/title`, `category`,
   `severity`, `subject`, `description`)
5. **Skriv uttrycket med `object.`** för resursen och `request.` för
   operationen

## Exempel: blockera radering av system-namespace

```yaml
      preconditions:
        any:
          - key: "{{ request.operation }}"
            operator: In
            value:
              - "DELETE"
      validate:
        message: "Deletion of the longhorn-system namespace is not allowed."
        cel:
          expressions:
            - expression: "false"
              message: "The longhorn-system namespace cannot be deleted."
```

## Kopplingar

- [[concepts/kyverno-policies]] - översikt av aktiva policies
- [[entities/kubernetes]]

## Referenser

- [Kyverno CEL Documentation](https://kyverno.io/docs/writing-policies/cel/)
