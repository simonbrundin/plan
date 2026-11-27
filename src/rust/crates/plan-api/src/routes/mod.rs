//! Router.
//!
//! Routern bygger samma vägar som Go-tjänsten (`src/api/cmd/api/main.go`).
//! Paritetstestet i `table.rs` håller de två i synk.

pub mod agent_approvals;
pub mod agent_runs;
pub mod agent_usage;
pub mod api_keys;
pub mod auth;
pub mod calendar_events;
pub mod calendars;
pub mod goals;
pub mod handlers;
pub mod places;
pub mod relations;
pub mod state;
pub mod statuses;
pub mod table;
pub mod tags;

use axum::body::Bytes;
use axum::extract::State;
use axum::http::{header, HeaderValue, Method, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::routing::get;
use axum::{Json, Router};
use tower_http::trace::TraceLayer;

use crate::config::Config;
use crate::docs::scalar_ui;
use crate::error::AppError;
use crate::openapi::openapi;
use crate::routes::state::AppState;

/// Läser en request-body och ger Go-kompatibla felmeddelanden.
///
/// Gin svarar med ett eget meddelande när bindningen misslyckas, till
/// exempel "childId and parentId are required". Axums inbyggda JSON-läsare
/// ger i stället en teknisk beskrivning och status 422 - därför läses
/// body:n manuellt här, så varje endpoint kan ange sitt eget meddelande.
pub fn parse_body<T: serde::de::DeserializeOwned>(
    body: &Bytes,
    meddelande: &str,
) -> Result<T, AppError> {
    serde_json::from_slice(body).map_err(|_| AppError::BadRequest(meddelande.to_string()))
}

/// Bygger API:ts router med tillstånd.
pub fn router(pool: Option<sqlx::postgres::PgPool>, config: &Config) -> Router {
    let state = AppState {
        pool,
        config: config.clone(),
    };

    // Autentiseringen ligger på API:ts routes - precis som i Go, där
    // middleware-kedjan sitter på gruppen `/api/v1`. Okända vägar ska
    // därför ge 404 som Gin gör, inte 401.
    let api = table::build().layer(axum::middleware::from_fn_with_state(
        state.clone(),
        crate::auth::authenticate,
    ));

    Router::new()
        .route("/docs", get(docs_redirect))
        // "/docs/" måste ha en egen route - axums jokertecken matchar
        // inte en tom segfault. Gin serverar sidan på "/docs/".
        .route("/docs/", get(scalar_ui))
        .route("/docs/{*any}", get(scalar_ui))
        .route("/docs/openapi.json", get(openapi_json))
        .merge(api)
        .fallback(gins_svar)
        .method_not_allowed_fallback(gins_svar)
        .layer(axum::middleware::from_fn(cors))
        .layer(axum::middleware::from_fn(panikfångare))
        .layer(TraceLayer::new_for_http())
        .with_state(state)
}

/// Fångar paniker i handlerna och svarar med 500.
///
/// Go läser rader med `lib/pq`, som klarar både INT4 och INT8 i samma
/// fält. Rust läser kolumner med exakt typ - en kolumn av fel typ ger
/// då en panic, och anropet hänger bara. Med den här lagren blir
/// svaret i stället ett 500, som i Go.
async fn panikfångare(request: axum::extract::Request, next: axum::middleware::Next) -> Response {
    use futures_util::FutureExt;

    let svar = std::panic::AssertUnwindSafe(next.run(request))
        .catch_unwind()
        .await;

    match svar {
        Ok(svar) => svar,
        Err(_) => {
            tracing::error!("Panik i en handler - svarar med 500");
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({ "message": "internal server error" })),
            )
                .into_response()
        }
    }
}

/// Gin skickar vidare "/docs" till "/docs/" - samma sak gör Rust.
///
/// Go använder `http.Redirect`, som svarar med en liten HTML-länk.
async fn docs_redirect() -> Response {
    (
        StatusCode::MOVED_PERMANENTLY,
        [
            (header::LOCATION, "/docs/"),
            (header::CONTENT_TYPE, "text/html; charset=utf-8"),
        ],
        "<a href=\"/docs/\">Moved Permanently</a>.\n\n",
    )
        .into_response()
}

/// Gin's svar för okända vägar och för metoder som inte stöds.
///
/// Gin har ingen egen hantering av metodfel - båda fallen ger
/// "404 page not found" i ren text.
async fn gins_svar() -> Response {
    (
        StatusCode::NOT_FOUND,
        [(header::CONTENT_TYPE, "text/plain")],
        "404 page not found",
    )
        .into_response()
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
