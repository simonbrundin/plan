//! API-nycklar.
//!
//! Port av `src/api/internal/handlers/apikeys.go`.
//!
//! Go-tjänsten svarar här med `{"error": ...}` i stället för
//! `{"message": ...}` - det är ett avtalet gränssnittet redan vänder sig
//! till, så formatet hålls här också.

use axum::body::Bytes;
use axum::extract::{Query, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::{Extension, Json};
use serde::Deserialize;
use serde_json::json;

use crate::auth::Caller;
use crate::repository::api_key;
use crate::routes::state::AppState;

/// Ett fel i API-nyckelhanteringen.
///
/// Go-tjänsten använder `gin.H{"error": ...}` för de här anropen - därför
/// har det egna felformatet.
#[derive(Debug)]
pub struct ApiKeyError {
    /// Statuskoden.
    pub status: StatusCode,
    /// Meddelandet.
    pub message: String,
}

impl ApiKeyError {
    /// Ett fel med given status.
    pub fn nytt(status: StatusCode, message: impl Into<String>) -> Self {
        ApiKeyError {
            status,
            message: message.into(),
        }
    }
}

impl IntoResponse for ApiKeyError {
    fn into_response(self) -> Response {
        (self.status, Json(json!({ "error": self.message }))).into_response()
    }
}

/// Användarens id ur anropet.
fn user_id(caller: &Caller) -> Result<i64, ApiKeyError> {
    if caller.user_id == 0 {
        return Err(ApiKeyError::nytt(StatusCode::UNAUTHORIZED, "Unauthorized"));
    }
    Ok(caller.user_id)
}

/// Databasen, eller ett fel som säger att den saknas.
fn pool(state: &AppState) -> Result<&sqlx::PgPool, ApiKeyError> {
    state.pool.as_ref().ok_or_else(|| {
        ApiKeyError::nytt(StatusCode::INTERNAL_SERVER_ERROR, "Database not connected")
    })
}

/// En API-nyckel i svaret.
#[derive(Debug, serde::Serialize, Deserialize)]
pub struct ApiKeyResponse {
    /// Nyckelns id.
    pub id: i64,
    /// Namnet.
    pub name: String,
    /// Scopet.
    pub scope: String,
    /// När nyckeln skapades.
    pub created: chrono::DateTime<chrono::Utc>,
    /// När nyckeln sist användes.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_used: Option<chrono::DateTime<chrono::Utc>>,
    /// När nyckeln går ut.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub expires: Option<chrono::DateTime<chrono::Utc>>,
}

/// Svaret när en nyckel skapas - innehåller den råa nyckeln en enda gång.
#[derive(Debug, serde::Serialize)]
pub struct CreatedApiKey {
    /// Nyckelns id.
    pub id: i64,
    /// Namnet.
    pub name: String,
    /// Scopet.
    pub scope: String,
    /// Den råa nyckeln - syns bara här.
    pub key: String,
}

/// Request-body för att skapa en nyckel.
#[derive(Debug, Default, Deserialize, utoipa::ToSchema)]
pub struct CreateApiKeyRequest {
    /// Namnet - krävs.
    #[serde(default)]
    pub name: String,
    /// Scopet - "full" eller "agent".
    #[serde(default)]
    pub scope: String,
    /// Utgångsdatum i RFC3339.
    #[serde(default)]
    pub expires: String,
}

/// Listar nycklarna.
#[utoipa::path(
    get,
    path = "/api/v1/api-keys",
    tag = "api-keys",
    responses(
        (status = 200, description = "Nycklarna"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn list(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
) -> Result<Json<Vec<ApiKeyResponse>>, ApiKeyError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let nycklar = api_key::list(pool, user_id)
        .await
        .map_err(|err| ApiKeyError::nytt(StatusCode::INTERNAL_SERVER_ERROR, err.to_string()))?;

    let svar: Vec<ApiKeyResponse> = nycklar
        .into_iter()
        .map(|nyckel| ApiKeyResponse {
            id: nyckel.id,
            name: nyckel.name,
            scope: nyckel.scope,
            created: nyckel.created,
            last_used: nyckel.last_used,
            expires: nyckel.expires,
        })
        .collect();

    Ok(Json(svar))
}

/// Skapar en nyckel.
#[utoipa::path(
    post,
    path = "/api/v1/api-keys",
    tag = "api-keys",
    request_body = CreateApiKeyRequest,
    responses(
        (status = 201, description = "Den nya nyckeln"),
        (status = 400, description = "Namnet saknas eller scopet är okänt"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn create(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    body: Bytes,
) -> Result<impl IntoResponse, ApiKeyError> {
    let req: CreateApiKeyRequest = crate::routes::parse_body(&body, "name is required")
        .map_err(|err| ApiKeyError::nytt(err.status(), err.to_string()))?;

    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    if req.name.trim().is_empty() {
        return Err(ApiKeyError::nytt(
            StatusCode::BAD_REQUEST,
            "name is required",
        ));
    }

    // Scopet är en teknisk gräns - en okänd scope ger inga rättigheter.
    let scope = if req.scope.is_empty() {
        crate::auth::apikey::SCOPE_FULL.to_string()
    } else {
        req.scope.clone()
    };

    if !crate::auth::apikey::giltig_scope(&scope) {
        return Err(ApiKeyError::nytt(
            StatusCode::BAD_REQUEST,
            format!("okänt scope: {}", req.scope),
        ));
    }

    let (id, rå) = api_key::create(pool, user_id, &req.name, &scope)
        .await
        .map_err(|_| {
            ApiKeyError::nytt(
                StatusCode::INTERNAL_SERVER_ERROR,
                "Failed to create API key",
            )
        })?;

    Ok((
        StatusCode::CREATED,
        Json(CreatedApiKey {
            id,
            name: req.name,
            scope,
            key: rå,
        }),
    ))
}

/// Query-parametern för borttagning.
#[derive(Debug, Default, Deserialize)]
pub struct DeleteParams {
    /// Nyckelns id.
    #[serde(default)]
    pub id: Option<String>,
}

/// Tar bort en nyckel.
///
/// Go-tjänsten registrerar vägen utan `:id`, så dess anrop alltid svarar
/// med "Invalid ID". Samma fel ges här - att ändra det skulle bryta
/// klienter som redan hanterar felet.
#[utoipa::path(
    delete,
    path = "/api/v1/api-keys",
    tag = "api-keys",
    responses(
        (status = 200, description = "Nyckeln togs bort"),
        (status = 400, description = "Ogiltigt id"),
        (status = 401, description = "Ingen giltig session"),
        (status = 404, description = "Nyckeln hittades inte")
    )
)]
pub async fn delete(
    State(state): State<AppState>,
    Extension(caller): Extension<Caller>,
    Query(params): Query<DeleteParams>,
) -> Result<Json<serde_json::Value>, ApiKeyError> {
    let user_id = user_id(&caller)?;
    let pool = pool(&state)?;

    let id = params
        .id
        .unwrap_or_default()
        .parse::<i64>()
        .map_err(|_| ApiKeyError::nytt(StatusCode::BAD_REQUEST, "Invalid ID"))?;

    let borttagen = api_key::delete(pool, user_id, id)
        .await
        .map_err(|err| ApiKeyError::nytt(StatusCode::INTERNAL_SERVER_ERROR, err.to_string()))?;

    if !borttagen {
        return Err(ApiKeyError::nytt(
            StatusCode::NOT_FOUND,
            "API key not found",
        ));
    }

    Ok(Json(json!({ "deleted": true })))
}
