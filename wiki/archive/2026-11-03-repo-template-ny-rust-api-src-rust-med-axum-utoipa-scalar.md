---
title: repo-template: ny Rust-API (src/rust) med Axum + utoipa + Scalar
type: note
tags: [rust, axum, utoipa, openapi, scalar, repo-template, feature, projekt:repo-template]
---

2026-11-03: Lade till src/rust - en komplett Rust-REST-API som alternativ till Go-API:t i src/go.

Stack: Axum 0.8 + Tokio + tower-http (CORS/tracing) + utoipa 6 med utoipa-axum 0.3 + Scalar UI. Detta är det mest etablerade Rust-REST-stacket; utoipa-axum ger code-first OpenAPI (specen genereras ur #[utoipa::path] och kan inte avvika från koden), till skillnad från Go-sidans swaggo + genererad swagger.json.

Varför code-first: ingen genererad fil att committa eller glömma regenerera. Paritet med Go upprätthålls: samma paths (/health, /hello), samma env-vars (PORT, APP_VERSION, API_HOST, CORS_ORIGINS), samma docs-URL:er (/docs = Scalar, /docs/openapi.json = spec).

Fällor som kostade tid:
- routes!(a, b) i utoipa-axum lägger båda handlers i samma method-router -> 'Overlapping method route' om de har olika paths. En .routes(routes!(handler)) per path.
- utoipa 6: derive-genererad openapi() returnerar impl OpenApi; vill man ha den konkreta openapi::OpenApi maskinen använder man <ApiDoc as utoipa::OpenApi>::openapi().
- tower-http 0.6 allow_origin tar Vec<HeaderValue>, inte Vec<String>.
- cargo fetch i Docker kräver att src/ finns (annars 'no targets specified in manifest') - lösning: placeholder-main.rs/lib.rs i ett lager, sedan riktig src.
- distroless saknar curl/wget, så HEALTHCHECK kör 'api healthcheck' - samma binär gör en TCP-probe (noll extra beroenden).

Verifierat: cargo test (5 tester gröna), clippy -D warnings rent, docker build + container kört (healthy), /health, /docs/openapi.json och /docs svarar korrekt.

Ej gjort: ingen k8s-integration (deployment/service, Tilt docker_build) och ingen databaskoppling - Go-API:t har inte det heller ännu.
