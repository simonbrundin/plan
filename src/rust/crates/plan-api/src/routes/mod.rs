//! Router.
//!
//! Routern bygger samma vägar som Go-tjänsten (`src/api/cmd/api/main.go`).
//! Paritetstestet i `table.rs` håller de två i synk.

pub mod goals;
pub mod handlers;
pub mod state;
pub mod statuses;
pub mod table;
pub mod tags;

use axum::extract::State;
use axum::http::{header, HeaderValue, Method, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::routing::get;
use axum::{Json, Router};
use tower_http::trace::TraceLayer;

use crate::config::Config;
use crate::docs::scalar_ui;
use crate::openapi::openapi;
use crate::routes::state::AppState;

/// Bygger API:ts router med tillstånd.
pub fn router(pool: Option<sqlx::postgres::PgPool>, config: &Config) -> Router {
    let state = AppState {
        pool,
        config: config.clone(),
    };

    Router::new()
        .route("/docs", get(scalar_ui))
        .route("/docs/{*any}", get(scalar_ui))
        .route("/docs/openapi.json", get(openapi_json))
        .merge(table::build())
        .layer(axum::middleware::from_fn(cors))
        .layer(axum::middleware::from_fn_with_state(
            state.clone(),
            crate::auth::authenticate,
        ))
        .layer(TraceLayer::new_for_http())
        .with_state(state)
}

/// OpenAPI-specen, genererad ur koden.
async fn openapi_json(State(state): State<AppState>) -> Json<utoipa::openapi::OpenApi> {
    Json(openapi(state.config.api_host.as_deref()))
}

/// CORS - samma headers som Go-tjänsten sätter.
async fn cors(request: axum::extract::Request, next: axum::middleware::Next) -> Response {
    let origin = std::env::var("CORS_ALLOWED_ORIGINS").unwrap_or_else(|_| "*".to_string());

    if request.method() == Method::OPTIONS {
        let mut response = StatusCode::NO_CONTENT.into_response();
        let headers = response.headers_mut();
        headers.insert(
            header::ACCESS_CONTROL_ALLOW_ORIGIN,
            HeaderValue::from_str(&origin).unwrap_or(HeaderValue::from_static("*")),
        );
        headers.insert(
            header::ACCESS_CONTROL_ALLOW_CREDENTIALS,
            HeaderValue::from_static("true"),
        );
        headers.insert(
            header::ACCESS_CONTROL_ALLOW_HEADERS,
            HeaderValue::from_static(
                "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
            ),
        );
        headers.insert(
            header::ACCESS_CONTROL_ALLOW_METHODS,
            HeaderValue::from_static("POST, OPTIONS, GET, PUT, PATCH, DELETE"),
        );
        return response;
    }

    let mut response = next.run(request).await;
    let headers = response.headers_mut();
    headers.insert(
        header::ACCESS_CONTROL_ALLOW_ORIGIN,
        HeaderValue::from_str(&origin).unwrap_or(HeaderValue::from_static("*")),
    );
    headers.insert(
        header::ACCESS_CONTROL_ALLOW_CREDENTIALS,
        HeaderValue::from_static("true"),
    );
    headers.insert(
        header::ACCESS_CONTROL_ALLOW_HEADERS,
        HeaderValue::from_static(
            "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
        ),
    );
    headers.insert(
        header::ACCESS_CONTROL_ALLOW_METHODS,
        HeaderValue::from_static("POST, OPTIONS, GET, PUT, PATCH, DELETE"),
    );

    response
}
