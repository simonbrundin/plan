//! Autentisering.
//!
//! Port av `src/api/internal/middleware`:
//!
//! * `apikey.rs` - API-nycklar (X-API-Key) med scope,
//! * `session.rs` - egna sessioner (HMAC-signerade cookies),
//! * `zitadel.rs` - Zitadel-JWT och introspektion,
//! * `scope.rs` - agent-scopets allowlist.
//!
//! Kedjan är densamma som i Go: API-nyckeln prövas först, sedan sessionen,
//! och till sist begränsas agent-nycklar till sitt scope.

pub mod apikey;
pub mod scope;
pub mod session;
pub mod zitadel;

use axum::extract::State;
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde_json::json;
use sqlx::Row;

use crate::routes::state::AppState;

/// Vem som anropar API:t.
#[derive(Debug, Clone)]
pub struct Caller {
    /// Användarens id i databasen.
    pub user_id: i64,
    /// Användarens `sub` från identitetsleverantören.
    pub sub: String,
    /// Användarens e-post.
    pub email: String,
    /// Nyckelns scope - "full" eller "agent".
    pub scope: String,
    /// Hur anropet autentiserades.
    pub method: AuthMethod,
}

/// Sättet anropet autentiserades.
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum AuthMethod {
    /// Via API-nyckel.
    ApiKey,
    /// Via session (cookie eller Bearer).
    Session,
    /// Via Zitadel-token.
    Zitadel,
    /// Via utvecklingsläget (AUTH_DISABLED).
    Disabled,
}

/// Sätter in identiteten i requesten, så att handlerna kan läsa den.
pub fn caller_from_request(request: &axum::extract::Request) -> Option<Caller> {
    request.extensions().get::<Caller>().cloned()
}

/// Kedjan av autentiseringsmiddleware.
///
/// Samma ordning som i Go:
///
/// 1. API-nyckel - om en nyckel skickas måste den vara giltig.
/// 2. Session - hoppas över om API-nyckeln redan satt identiteten.
/// 3. Agent-scope - en agent-nyckel får bara göra sina egna anrop.
pub async fn authenticate(
    State(state): State<AppState>,
    mut request: axum::extract::Request,
    next: axum::middleware::Next,
) -> Response {
    let path = request.uri().path().to_string();
    let method = request.method().as_str().to_string();

    // Autentiseringsendpoints är publika - de är själva vägen in.
    if !path.starts_with("/api/v1") || path.starts_with("/api/v1/auth/") {
        return next.run(request).await;
    }

    // Utvecklingsläget hoppar över allt.
    if std::env::var("AUTH_DISABLED").ok().as_deref() == Some("true") {
        let caller = Caller {
            user_id: 4,
            sub: "378032824856347117".to_string(),
            email: "simonbrundin@gmail.com".to_string(),
            scope: apikey::SCOPE_FULL.to_string(),
            method: AuthMethod::Disabled,
        };
        request.extensions_mut().insert(caller);
        return next.run(request).await;
    }

    // 1. API-nyckel.
    let api_key = request
        .headers()
        .get(apikey::API_KEY_HEADER)
        .and_then(|value| value.to_str().ok())
        .map(str::to_string)
        .or_else(|| {
            request.uri().query().and_then(|query| {
                query.split('&').find_map(|par| {
                    let (nyckel, värde) = par.split_once('=')?;
                    (nyckel == apikey::API_KEY_QUERY_PARAM).then(|| värde.to_string())
                })
            })
        });

    if let Some(api_key) = api_key.filter(|key| !key.is_empty()) {
        match apikey::validate_api_key(state.pool.as_ref(), &api_key).await {
            Ok(Some(identitet)) => {
                let caller = Caller {
                    user_id: identitet.user_id,
                    sub: String::new(),
                    email: String::new(),
                    scope: identitet.scope,
                    method: AuthMethod::ApiKey,
                };

                // 3. Agent-scope.
                if caller.scope == apikey::SCOPE_AGENT && !scope::agent_scope_allows(&method, &path)
                {
                    return forbidden();
                }

                // 4. Läs-scope - får bara läsa.
                if caller.scope == apikey::SCOPE_READONLY && method != "GET" {
                    return forbidden();
                }

                request.extensions_mut().insert(caller);
                return next.run(request).await;
            }
            // Nyckeln finns inte - Go svarar då utan extra meddelande.
            Ok(None) => {
                return (
                    StatusCode::UNAUTHORIZED,
                    Json(json!({
                        "error": "Invalid API key",
                    })),
                )
                    .into_response();
            }
            Err(err) => {
                return (
                    StatusCode::UNAUTHORIZED,
                    Json(json!({
                        "error": "Invalid API key",
                        "message": err,
                    })),
                )
                    .into_response();
            }
        }
    }

    // 2. Session.
    let token = session::session_token_from_request(
        request
            .headers()
            .get(axum::http::header::COOKIE)
            .and_then(|value| value.to_str().ok())
            .and_then(cookie_value)
            .as_deref(),
        request
            .headers()
            .get(axum::http::header::AUTHORIZATION)
            .and_then(|value| value.to_str().ok()),
    );

    let Some(token) = token else {
        return (
            StatusCode::UNAUTHORIZED,
            Json(json!({ "message": "Authorization header required" })),
        )
            .into_response();
    };

    // Egen session.
    if let Some(pool) = state.pool.as_ref() {
        if let Ok(session) =
            session::validate_session(pool, state.config.session_secret.as_bytes(), &token).await
        {
            let caller = Caller {
                user_id: session.user_id,
                sub: session.sub,
                email: session.email,
                scope: apikey::SCOPE_FULL.to_string(),
                method: AuthMethod::Session,
            };
            request.extensions_mut().insert(caller);
            return next.run(request).await;
        }
    }

    // Utvecklingstoken: "user_<id>".
    if let Some(user_id) = token.strip_prefix("user_") {
        if !user_id.is_empty() && user_id.chars().all(|tecken| tecken.is_ascii_digit()) {
            let user_id: i64 = user_id.parse().unwrap_or(0);
            let caller = Caller {
                user_id,
                sub: format!("user_{user_id}"),
                email: String::new(),
                scope: apikey::SCOPE_FULL.to_string(),
                method: AuthMethod::Session,
            };
            request.extensions_mut().insert(caller);
            return next.run(request).await;
        }

        return (
            StatusCode::UNAUTHORIZED,
            Json(json!({ "message": "Invalid user token" })),
        )
            .into_response();
    }

    // Zitadel-JWT.
    let domain = zitadel::zitadel_domain();
    if let Some(identitet) = zitadel::validate_jwt(&token, &domain).await {
        let user_id = match lookup_or_create_user(
            state.pool.as_ref(),
            &identitet.sub,
            &identitet.email,
        )
        .await
        {
            Ok(user_id) => user_id,
            Err(_) => {
                return (
                    StatusCode::UNAUTHORIZED,
                    Json(json!({ "message": "Failed to create user session" })),
                )
                    .into_response();
            }
        };

        let caller = Caller {
            user_id,
            sub: identitet.sub,
            email: identitet.email,
            scope: apikey::SCOPE_FULL.to_string(),
            method: AuthMethod::Zitadel,
        };
        request.extensions_mut().insert(caller);
        return next.run(request).await;
    }

    // Opaque token - introspektion.
    match zitadel::introspect_token(&token, &domain).await {
        Ok(svar) if svar.active => {
            let identitet = zitadel::identity_from_introspection(&svar);
            let user_id =
                match lookup_or_create_user(state.pool.as_ref(), &identitet.sub, &identitet.email)
                    .await
                {
                    Ok(user_id) => user_id,
                    Err(_) => {
                        return (
                            StatusCode::UNAUTHORIZED,
                            Json(json!({ "message": "Failed to create user session" })),
                        )
                            .into_response();
                    }
                };

            let caller = Caller {
                user_id,
                sub: identitet.sub,
                email: identitet.email,
                scope: apikey::SCOPE_FULL.to_string(),
                method: AuthMethod::Zitadel,
            };
            request.extensions_mut().insert(caller);
            next.run(request).await
        }
        Ok(_) => (
            StatusCode::UNAUTHORIZED,
            Json(json!({ "message": "Token expired or revoked" })),
        )
            .into_response(),
        Err(_) => (
            StatusCode::UNAUTHORIZED,
            Json(json!({ "message": "Invalid token" })),
        )
            .into_response(),
    }
}

/// Svarar 403 - agent-nyckeln får inte göra anropet.
fn forbidden() -> Response {
    (
        StatusCode::FORBIDDEN,
        Json(json!({
            "error": "Forbidden",
            "message": "Agent API-nyckeln får inte göra det här anropet",
        })),
    )
        .into_response()
}

/// Läser ett cookie-värde ur en cookie-header.
fn cookie_value(header: &str) -> Option<String> {
    header.split(';').find_map(|par| {
        let (namn, varde) = par.split_once('=')?;
        (namn.trim() == session::SESSION_COOKIE_NAME).then(|| varde.trim().to_string())
    })
}

/// Slår upp en användare på `sub`, och skapar den om den saknas.
///
/// Användaren hittas först på `sub`, sedan på e-post - och i det fallet
/// uppdateras `sub`.
pub async fn lookup_or_create_user(
    pool: Option<&sqlx::PgPool>,
    sub: &str,
    email: &str,
) -> Result<i64, sqlx::Error> {
    let Some(pool) = pool else {
        return Err(sqlx::Error::Configuration("database not connected".into()));
    };

    if let Some(row) = sqlx::query("SELECT id FROM users WHERE sub = $1")
        .bind(sub)
        .fetch_optional(pool)
        .await?
    {
        return Ok(row.get::<i32, _>(0) as i64);
    }

    if !email.is_empty() {
        if let Some(row) = sqlx::query("SELECT id FROM users WHERE email = $1")
            .bind(email)
            .fetch_optional(pool)
            .await?
        {
            let user_id = row.get::<i32, _>(0) as i64;
            sqlx::query("UPDATE users SET sub = $1 WHERE id = $2")
                .bind(sub)
                .bind(user_id as i32)
                .execute(pool)
                .await?;
            return Ok(user_id);
        }
    }

    let row = sqlx::query("INSERT INTO users (sub, email) VALUES ($1, $2) RETURNING id")
        .bind(sub)
        .bind(email)
        .fetch_one(pool)
        .await?;

    Ok(row.get::<i32, _>(0) as i64)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn cookie_varde_lases_ur_headern() {
        assert_eq!(
            cookie_value("plan_session=abc; other=1").as_deref(),
            Some("abc")
        );
        assert_eq!(cookie_value("other=1"), None);
        assert_eq!(cookie_value(""), None);
    }
}
