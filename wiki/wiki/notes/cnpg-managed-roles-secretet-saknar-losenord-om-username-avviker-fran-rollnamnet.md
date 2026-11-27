---
title: CNPG managed roles - sekretets username måste matcha rollnamnet
type: note
summary: Om username i managed role-sekretet skiljer sig från rollnamnet vägrar CNPG synka rollen ("wrong username ... expected ...") och rollen får aldrig något lösenord.
tags: [cnpg, postgresql, secrets, bug, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/cloudnative-pg, notes/cnpg-initdb-agarollen-skapas-fore-postinitsql]
---

# CNPG managed roles - sekretets username måste matcha rollnamnet

Efter att initdb-klustret för `plan-db` gick upp igen saknade applikationsrollen
`plan` fortfarande lösenord, och Atlas-applikationen kunde inte schemalägga:

```
ERROR:  role "plan" does not exist
```

## Varför

`managed.roles` i CNPG synkas mot en Secret per roll. Operatören jämför
`username` i sekretet med rollnamnet i klusterresursen. I Plan-repots hemlighet
`plan-db-user` stod rollnamnet fortfarande som `user` i fälten `username`,
`uri`, `jdbc-uri` och `pgpass` - efter att rollen bytts namn till `plan`.

Felet i operatörens logg var:

```
username "user" in secret "plan-db-user" does not match role name "plan"
```

Sekretet synkades aldrig, så `plan` hade inget lösenord och Atlas kunde inte
använda rollen.

## Lösning

Alla fält i sekretet ska använda samma rollnamn som i klusterresursen:

```yaml
stringData:
  username: plan
  password: <hemlighet>
  uri: postgresql://plan:<hemlighet>@plan-db-rw.plan.svc.cluster.local:5432/plan
  jdbc-uri: jdbc:postgresql://plan:<hemlighet>@plan-db-rw.plan.svc.cluster.local:5432/plan
  pgpass: <hemlighet>
```

Kontrollera efter en ändring att operatören verkligen synkade rollen:

```bash
kubectl get cluster plan-db -n plan -o jsonpath='{.status.managedRoles}'
kubectl exec -n plan plan-db-2 -- psql -c "\du"
```

## Kopplingar

- [[concepts/cloudnative-pg]]
- [[notes/cnpg-initdb-agarollen-skapas-fore-postinitsql]]
