//! Dokumentationsroutes.
//!
//! Go-tjänsten serverar Swagger UI på "/docs/" och skickar "/docs" vidare
//! dit med 301. Rust gör samma sak - Scalar läser specen från
//! "/docs/openapi.json".

use axum::body::Body;
use axum::http::{Request, StatusCode};
use tower::ServiceExt;

use plan_api::config::Config;
use plan_api::routes::router;

/// Bygger en router utan databas.
fn app() -> axum::Router {
    router(None, &Config::from_env())
}

/// Skickar ett anrop och ger tillbaka statuskoden.
async fn status(app: axum::Router, path: &str) -> StatusCode {
    let svar = app
        .oneshot(Request::builder().uri(path).body(Body::empty()).unwrap())
        .await
        .unwrap();

    svar.status()
}

#[tokio::test]
async fn docs_vidarevisas_till_slash() {
    assert_eq!(status(app(), "/docs").await, StatusCode::MOVED_PERMANENTLY);
}

#[tokio::test]
async fn docs_sidan_serveras() {
    // "/docs/" måste svara 200 - det är adressen Scalar annonserar.
    assert_eq!(status(app(), "/docs/").await, StatusCode::OK);
    assert_eq!(status(app(), "/docs/index.html").await, StatusCode::OK);
}

#[tokio::test]
async fn openapi_specen_serveras() {
    assert_eq!(status(app(), "/docs/openapi.json").await, StatusCode::OK);
}
