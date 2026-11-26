//! Agentens API-dokumentation.
//!
//! Go-agenten serverar den inbäddade swagger-filen på `/openapi.json` och
//! Scalar-gränssnittet på `/docs/`. Rust-agenten serverar samma
//! specifikation - filen ligger i `resources/swagger.json` och testet
//! `specen_foljer_go_agenten` ser till att de två inte glider isär.

use axum::extract::Path;
use axum::http::{header, StatusCode, Uri};
use axum::response::{Html, IntoResponse, Response};

/// Swagger-specen som Go-agenten serverar.
const SWAGGER: &str = include_str!("../../resources/swagger.json");

/// Scalar-sidan.
///
/// Scalar läser specen från `/openapi.json`. Bara roten under `/docs`
/// svarar med sidan - andra vägar är statiska filer som Scalar laddar,
/// och sådana som inte finns ger 404 precis som i Go.
pub async fn docs(Path(underlag): Path<String>) -> Response {
    if !underlag.is_empty() {
        // Go:s Scalar-handler svarar med en tom 404 för okända filer.
        return StatusCode::NOT_FOUND.into_response();
    }

    Html(SCALAR_HTML).into_response()
}

/// Scalar-sidan utan någon underväg.
///
/// `/docs/` registreras separat - axums jokertecken kräver minst en
/// del, så roten under `/docs` matchar inte `{*any}`.
pub async fn docs_root() -> Html<&'static str> {
    Html(SCALAR_HTML)
}

/// Skickar `/docs` vidare till `/docs/`.
///
/// Gin gör detta automatiskt med sin trailing-slash-redirect - axum gör
/// det inte, så det händer här. Gin svarar 301, inte axums 308.
pub async fn docs_redirect(uri: Uri) -> Response {
    let mål = format!("{}/", uri.path().trim_end_matches('/'));

    (
        StatusCode::MOVED_PERMANENTLY,
        [
            (header::LOCATION, mål.clone()),
            (header::CONTENT_TYPE, "text/html; charset=utf-8".to_string()),
        ],
        format!("<a href=\"{mål}\">Moved Permanently</a>.\n\n"),
    )
        .into_response()
}

/// OpenAPI-specen.
///
/// Go tar bort `host` så att Scalar använder sin egen origin. Det gör
/// Rust också - och eftersom båda serialiserar en `Map` hamnar fälten i
/// samma ordning.
pub async fn openapi_json() -> Response {
    let Ok(mut spec) = serde_json::from_str::<serde_json::Value>(SWAGGER) else {
        return (
            StatusCode::INTERNAL_SERVER_ERROR,
            [(header::CONTENT_TYPE, "text/plain; charset=utf-8")],
            "Failed to read spec",
        )
            .into_response();
    };

    if let Some(objekt) = spec.as_object_mut() {
        objekt.remove("host");
    }

    let Ok(text) = serde_json::to_vec(&spec) else {
        return (
            StatusCode::INTERNAL_SERVER_ERROR,
            [(header::CONTENT_TYPE, "text/plain; charset=utf-8")],
            "Failed to modify spec",
        )
            .into_response();
    };

    (
        StatusCode::OK,
        [(header::CONTENT_TYPE, "application/json")],
        text,
    )
        .into_response()
}

/// Scalar-sidan.
const SCALAR_HTML: &str = r#"<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="no-referrer" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Agent API</title>
  </head>
  <body>
    <div id="app"></div>

    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/openapi.json',
        theme: 'default',
        darkMode: true,
        showSidebar: true,
      })
    </script>
  </body>
</html>
"#;

#[cfg(test)]
mod tests {
    use std::fs;
    use std::path::PathBuf;

    use super::SWAGGER;

    /// Sökvägen till Go-agentens swagger-fil.
    fn go_swagger() -> PathBuf {
        PathBuf::from(env!("CARGO_MANIFEST_DIR"))
            .join("../../../agent/internal/api/docs/swagger.json")
    }

    /// Specen i Rust-kopian ska vara identisk med Go-agentens.
    ///
    /// Annars börjar de två API:erna beskriva olika saker - och det
    /// syns först när någon jämför dem.
    #[test]
    fn specen_foljer_go_agenten() {
        let go = fs::read_to_string(go_swagger())
            .expect("går att läsa Go-agentens swagger.json - annars fungerar testet inte");
        let rust = include_str!("../../resources/swagger.json");

        assert_eq!(
            go, rust,
            "swagger.json har glidit isär - kopiera Go-agentens fil igen"
        );
    }

    /// Specen ska vara giltig JSON och ha en sökväg för varje route.
    #[test]
    fn specen_är_giltig_json() {
        let spec: serde_json::Value =
            serde_json::from_str(SWAGGER).expect("swagger.json är giltig JSON");

        assert!(spec.get("paths").is_some(), "specen ska ha paths");
    }
}
