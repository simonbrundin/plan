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
use axum::http::{header, HeaderMap, HeaderValue, Method, StatusCode};
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
        .route("/docs/{*any}", get(scalar_ui))
        .route("/docs/openapi.json", get(openapi_json))
        .merge(api)
        .fallback(gins_svar)
        .method_not_allowed_fallback(gins_svar)
        .layer(axum::middleware::from_fn(cors))
        .layer(TraceLayer::new_for_http())
        .with_state(state)
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

/// CORS - samma regler som Go-tjänsten.
///
/// Webbläsaren anropar API:t direkt (Nuxt har ingen proxy), så API:t måste
/// svara med CORS-headrar när anropet kommer från en annan origin - i dev
/// kör appen på :3100 och API:t på :4100.
///
/// `CORS_ALLOWED_ORIGINS` är en kommaseparerad lista. Origin echas tillbaka
/// (aldrig som wildcard) - ett credentialed anrop med sessioncookien
/// blockeras av webbläsaren om `Access-Control-Allow-Origin` är `*`.
async fn cors(request: axum::extract::Request, next: axum::middleware::Next) -> Response {
    let origin = request
        .headers()
        .get(header::ORIGIN)
        .and_then(|value| value.to_str().ok())
        .map(|value| value.to_string());

    let allowed = origin
        .as_deref()
        .map(|origin| origin_allowed(origin))
        .unwrap_or(false);

    if request.method() == Method::OPTIONS {
        let mut response = StatusCode::NO_CONTENT.into_response();
        apply_cors_headers(response.headers_mut(), origin.as_deref(), allowed);
        return response;
    }

    let mut response = next.run(request).await;
    apply_cors_headers(response.headers_mut(), origin.as_deref(), allowed);

    response
}

/// Sätter CORS-headrarna på svaret.
fn apply_cors_headers(headers: &mut HeaderMap, origin: Option<&str>, allowed: bool) {
    if allowed {
        if let Some(origin) = origin {
            if let Ok(value) = HeaderValue::from_str(origin) {
                headers.insert(header::ACCESS_CONTROL_ALLOW_ORIGIN, value);
            }
        }
        headers.insert(
            header::ACCESS_CONTROL_ALLOW_CREDENTIALS,
            HeaderValue::from_static("true"),
        );
        headers.append(header::VARY, HeaderValue::from_static("Origin"));
    }
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
}

/// Returnerar om en origin får anropa API:t från webbläsaren. Origin jämförs
/// exakt - scheme, värd och port.
fn origin_allowed(origin: &str) -> bool {
    let allowed = std::env::var("CORS_ALLOWED_ORIGINS").unwrap_or_default();

    allowed.split(',').any(|entry| {
        let entry = entry.trim();
        !entry.is_empty() && (entry == "*" || entry == origin)
    })
}
