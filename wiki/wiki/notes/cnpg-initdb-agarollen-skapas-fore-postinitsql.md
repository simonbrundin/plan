---
title: CNPG skapar initdb-ägarrollen innan postInitSQL körs - managed.roles kommer för sent
type: note
summary: postInitSQL i CloudNativePG körs före managed.roles, så GRANT-satser mot en managed role ger 'role does not exist'. Lösningen är bootstrap.initdb.owner.
tags: [cnpg, postgresql, bug, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/cloudnative-pg, notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor]
---

# CNPG skapar initdb-ägarrollen innan postInitSQL körs - managed.roles kommer för sent

Plan-appens databas `plan-db` (CNPG) gick aldrig initdb-klar. Loggen från
`plan-db-1-initdb` gav:

```
ERROR:  role "plan" does not exist (SQLSTATE 42704)
```

## Varför

Bootstrap-fasen i CloudNativePG har en bestämd ordning:

1. `bootstrap.initdb.owner` skapas som databasens ägare **innan** `postInitSQL` körs.
2. `postInitSQL` körs.
3. `managed.roles` skapas **efter** bootstrapen.

`postInitSQL` i Plan-repots `environments/base/db/postgres-cluster.yaml` gav
`GRANT ALL ON SCHEMA public TO plan` - men `plan` var en *managed role*, som
alltså inte existerade ännu. Initdb-jobbet föll, CNPG-klustret fastnade i
"Setting up primary" och Go API:t svarade `Failed to create user session`
(`dial tcp 10.108.196.67:5432: connect: connection refused`).

## Lösning

Sätt `bootstrap.initdb.owner: plan` och låt CNPG skapa rollen. Då existerar
`plan` redan när `postInitSQL` körs, och schemat + tabellerna ägs av
applikationsrollen:

```yaml
bootstrap:
  initdb:
    owner: plan
    postInitSQL:
      - "GRANT ALL ON SCHEMA public TO plan"
      - "ALTER DEFAULT PRIVILEGES FOR ROLE plan IN SCHEMA public GRANT ALL ON TABLES TO plan"
      - "ALTER DEFAULT PRIVILEGES FOR ROLE plan IN SCHEMA public GRANT ALL ON SEQUENCES TO plan"
```

`managed.roles` behövs inte alls då - CNPG:s egen initdb-ägare skapar rollen
och operatören sätter lösenordet utifrån `plan-db-user`.

## Fällan: `DO $$ ... $$`

Första försöket var att skapa rollen manuellt i `postInitSQL` med
`DO $$ ... $$`. Det misslyckades med:

```
ERROR:  syntax error at or near "$" (SQLSTATE 42601)
```

CNPG delar upp `postInitSQL` på semikolon och skickar varje del till
PostgreSQL separat - dollarquoting överlever inte den uppdelningen. Använd
`bootstrap.initdb.owner` i stället för `CREATE ROLE`/`DO`-satser.

## Kopplingar

- [[concepts/cloudnative-pg]]
- [[notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor]]
