//! Scalar UI (API Reference) på `/docs`.
//!
//! Samma uppsättning som Go-tjänsten: Scalar läser OpenAPI-specen från
//! `/docs/openapi.json` och visas i mörkt tema.

use axum::response::Html;

/// Scalar-sidan.
pub async fn scalar_ui() -> Html<&'static str> {
    Html(SCALAR_HTML)
}

const SCALAR_HTML: &str = r#"<!doctype html>
<html>
  <head>
    <title>Goals Management API</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <div id="app"></div>

    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/docs/openapi.json',
        theme: 'default',
        darkMode: true,
        showSidebar: true,
      })
    </script>
  </body>
</html>
"#;
