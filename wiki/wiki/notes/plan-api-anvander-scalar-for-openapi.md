---
title: Plan API använder Scalar för OpenAPI
type: note
tags: [openapi, scalar, go, dokumentation, projekt:plan]
created: 2026-11-17
updated: 2026-11-17
status: active
---

Plan API bytte från gin-swagger till samma Scalar-upplägg som repo-template. Go-servern serverar Scalar på /docs/ och OpenAPI-specen på /docs/openapi.json via gin-openapi, med API_HOST-stöd och DOCS_PATH. Docker genererar swag-dokumentationen i ett separat byggsteg och kopierar specen till runtime. Kubernetes HTTPRoute pekar därför på /docs i stället för /swagger. swag kunde inte generera API-key-dokumentationen förrän ErrorResponse-annotationerna kvalificerades som models.ErrorResponse.
