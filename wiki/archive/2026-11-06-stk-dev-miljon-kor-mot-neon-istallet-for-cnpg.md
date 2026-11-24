---
title: stk: dev-miljön kör mot Neon istället för CNPG
type: note
tags: [infra, neon, database, dev, projekt:stk]
---

I repo-template/stk byttes dev-databasen ut från ett CNPG-kluster i klustret till Neon (Lakebase Postgres, ep-odd-bar-b9q2tfon.eu-central-1). CNPG-klustret repo-template-db är borttaget ur environments/dev-overlayn.

Varför: samma databas som resten av miljöerna, ingen databas i kluster som ska backupas, och Postgres-versionen är alltid aktuell (Neon kör 18, klustret körde 18.6).

Hur: environments/dev/db/neon.env (gitignoras) innehåller DATABASE_URL (poolad), DATABASE_URL_UNPOOLED och uri (direkt, utan channel_binding - Atlas använder lib/pq som inte stöder parametern). Kustomize genererar Secreten repo-template-db-neon ur filen. Atlas Operator (v0.8.0) applicerar environments/base/db/schema.sql mot Neon via urlFrom/uri och deployar en egen dev-databas-pod i klustret - ingen extra databas behövs på Neon. devURLFrom i CRD:n ignoreras av operator 0.8.0, så det ska inte sättas.

Fällor: kustomize v5.8.3 stöder bara EN valuesFile per helmChart (valuesFile-lista fungerar inte), så pgAdmin-värden genereras av environments/dev/db/sync-neon-env.sh ur neon.env (basfiler + serverdefinition med lösenord). Dessutom: kubectl apply av rå kustomize-output nollställer image-referenserna till go-api:latest - Tilt patchar dem till localhost:5000/<image>:tilt-<hash>, så man måste antingen köra tilt up eller kubectl set image efteråt.
