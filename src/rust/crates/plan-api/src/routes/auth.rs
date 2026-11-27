//! Auth-endpoints.
//!
//! Port av `src/api/internal/handlers/auth.go`. Loginflödet är samma som i
//! Go: PKCE mot Zitadel, en egen sessionstoken i `plan_session`, och en
//! redirect tillbaka till gränssnittet. Sessionstokenen är byte-kompatibel
//! med Go-tjänstens, så en cookie fungerar i båda under migreringen.
//!
//! Go svarar med `{"error": "..."}` på auth-endpoints - inte `{"message":
//! ...}` som resten av API:t. Det är därför svaren byggs här direkt.

use axum::extract::{Query, State};
use axum::http::{header, HeaderMap, HeaderValue, StatusCode};
use axum::response::{IntoResponse, Response};
use axum::Json;
use base64::engine::general_purpose::URL_SAFE_NO_PAD;
use base64::Engine;
use serde::Deserialize;
use serde_json::json;
use sha2::{Digest, Sha256};

use crate::auth::session;
use crate::repository::user;
use crate::routes::state::AppState;

/// Namnet på cookien som håller OAuth-state under inloggningen.
const OAUTH_STATE_COOKIE: &str = "oauth_state";

/// Hur länge OAuth-state-cookien håller.
const OAUTH_STATE_MAX_AGE: i64 = 600;

/// Parametrarna Zitadel skickar tillbaka till callbacken.
#[derive(Debug, Deserialize)]
pub struct CallbackQuery {
    /// Auktorisationskoden.
    pub code: Option<String>,
    /// State som skickades med i authorize-anropet.
    pub state: Option<String>,
    /// Felkod, om Zitadel avbröt.
    pub error: Option<String>,
    /// Förklaring till felet.
    pub error_description: Option<String>,
}

/// Startar inloggningen.
///
/// I utvecklingsläget (`AUTH_DISABLED=true`) skickas användaren till appens
/// inloggningssida, där dev-login-knappen loggar in utan Zitadel. Annars
/// byggs en PKCE-url mot Zitadel.
#[utoipa::path(
    get,
    path = "/api/v1/auth/login",
    tag = "auth",
    responses(
        (status = 307, description = "Omdirigering till Zitadel"),
        (status = 400, description = "Zitadel är inte konfigurerat")
    )
)]
pub async fn login() -> Response {
    if env("AUTH_DISABLED").as_deref() == Some("true") {
        return redirect_temporary(&format!("{}/login", frontend_base_url()));
    }

    let domain = env("ZITADEL_DOMAIN").unwrap_or_default();
    let client_id = env("ZITADEL_CLIENT_ID").unwrap_or_default();
    let worktree = env("WORKTREE_NAME").unwrap_or_default();

    // Samma redirect_uri i authorize-anropet och i tokenbytet - annars
    // avvisar Zitadel bytet.
    let redirect_uri = oauth_redirect_uri();

    let mut cookies = Vec::new();

    // I worktree-läge markeras webbläsaren, så att dev-gatewayn kan
    // skicka callbacken till rätt worktree.
    if !worktree.is_empty() {
        cookies.push(format!(
            "wt={worktree}; Path=/; Max-Age={OAUTH_STATE_MAX_AGE}{}",
            secure_suffix()
        ));
    }

    if domain.is_empty() || client_id.is_empty() {
        return json_error(StatusCode::BAD_REQUEST, "Zitadel configuration missing");
    }

    // PKCE: verifieraren är hemlig, challengen skickas i klartext.
    let verifier = random_string(64);
    let challenge = URL_SAFE_NO_PAD.encode(Sha256::digest(verifier.as_bytes()));
    let state_value = random_string(32);

    // State sparas i cookien - det binder callbacken till den här
    // webbläsaren.
    let state_data = json!({
        "state": state_value,
        "verifier": verifier,
        "worktree": worktree,
    });
    cookies.push(format!(
        "{OAUTH_STATE_COOKIE}={}; Path=/; Max-Age={OAUTH_STATE_MAX_AGE}; HttpOnly{}",
        query_escape(&state_data.to_string()),
        secure_suffix()
    ));

    let auth_url = format!(
        "https://{domain}/oauth/v2/authorize?response_type=code&client_id={}&redirect_uri={}&scope=openid%20email%20profile&state={}&code_challenge={}&code_challenge_method=S256",
        query_escape(&client_id),
        query_escape(&redirect_uri),
        query_escape(&state_value),
        query_escape(&challenge),
    );

    redirect_temporary_med_cookies(&auth_url, cookies)
}

/// Hanterar callbacken från Zitadel.
///
/// Koden byts mot en access token, användaren slås upp eller skapas, och en
/// egen sessionstoken sätts i `plan_session`.
#[utoipa::path(
    get,
    path = "/api/v1/auth/callback",
    tag = "auth",
    params(
        ("code" = Option<String>, Query, description = "Auktorisationskoden"),
        ("state" = Option<String>, Query, description = "State från authorize-anropet"),
        ("error" = Option<String>, Query, description = "Felkod från Zitadel"),
        (
            "error_description" = Option<String>,
            Query,
            description = "Förklaring till felet"
        )
    ),
    responses(
        (status = 307, description = "Omdirigering till gränssnittet"),
        (status = 400, description = "Kod, state eller OAuth-state saknas"),
        (status = 500, description = "Tokenbytet eller userinfo misslyckades")
    )
)]
pub async fn callback(
    State(state): State<AppState>,
    Query(params): Query<CallbackQuery>,
    headers: HeaderMap,
) -> Response {
    let code = params.code.unwrap_or_default();
    let state_value = params.state.unwrap_or_default();

    if let Some(error) = params.error.filter(|value| !value.is_empty()) {
        return json_error_with_description(
            StatusCode::BAD_REQUEST,
            &error,
            &params.error_description.unwrap_or_default(),
        );
    }

    if code.is_empty() || state_value.is_empty() {
        return json_error(StatusCode::BAD_REQUEST, "Missing code or state");
    }

    let Some(state_cookie) = cookie_value(&headers, OAUTH_STATE_COOKIE) else {
        return json_error(StatusCode::BAD_REQUEST, "Missing OAuth state");
    };

    // Gin skriver cookie-värden url-escapade och läser dem avkodade.
    let state_cookie = query_unescape(&state_cookie);

    let stored: serde_json::Value = match serde_json::from_str(&state_cookie) {
        Ok(value) => value,
        Err(_) => return json_error(StatusCode::BAD_REQUEST, "Invalid OAuth state"),
    };

    if stored.get("state").and_then(|value| value.as_str()) != Some(state_value.as_str()) {
        return json_error(StatusCode::BAD_REQUEST, "State mismatch");
    }

    // State-cookien rensas så att den inte kan återanvändas.
    let mut cookies = vec![format!(
        "{OAUTH_STATE_COOKIE}=; Path=/; Max-Age=0; HttpOnly{}",
        secure_suffix()
    )];

    let domain = env("ZITADEL_DOMAIN").unwrap_or_default();
    let client_id = env("ZITADEL_CLIENT_ID").unwrap_or_default();
    let client_secret = env("ZITADEL_CLIENT_SECRET").unwrap_or_default();
    let redirect_uri = oauth_redirect_uri();
    let verifier = stored
        .get("verifier")
        .and_then(|value| value.as_str())
        .unwrap_or_default()
        .to_string();

    // Koden byts mot en access token.
    let client = match reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(10))
        .build()
    {
        Ok(client) => client,
        Err(err) => {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                &format!("Token exchange failed: {err}"),
            )
        }
    };

    let mut request = client
        .post(format!("https://{domain}/oauth/v2/token"))
        .form(&[
            ("grant_type", "authorization_code".to_string()),
            ("client_id", client_id.clone()),
            ("code", code.clone()),
            ("redirect_uri", redirect_uri.clone()),
            ("code_verifier", verifier.clone()),
        ]);

    if !client_secret.is_empty() {
        request = request.basic_auth(&client_id, Some(&client_secret));
    }

    let token_response = match request.send().await {
        Ok(response) => response,
        Err(err) => {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                &format!("Token exchange failed: {err}"),
            )
        }
    };

    let body = match token_response.json::<serde_json::Value>().await {
        Ok(body) => body,
        Err(_) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, "Invalid token response"),
    };

    if let Some(error) = body.get("error").filter(|value| !value.is_null()) {
        let description = body
            .get("error_description")
            .and_then(|value| value.as_str())
            .unwrap_or_default();
        return json_error_with_description(
            StatusCode::BAD_REQUEST,
            error.as_str().unwrap_or(""),
            description,
        );
    }

    let Some(access_token) = body.get("access_token").and_then(|value| value.as_str()) else {
        return json_error(StatusCode::INTERNAL_SERVER_ERROR, "Invalid token response");
    };

    // Användarens uppgifter hämtas från userinfo-endpointen.
    let user_info = match client
        .get(format!("https://{domain}/oidc/v1/userinfo"))
        .bearer_auth(access_token)
        .send()
        .await
    {
        Ok(response) => response,
        Err(_) => {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                "User info request failed",
            )
        }
    };

    let user_info = match user_info.json::<serde_json::Value>().await {
        Ok(value) => value,
        Err(_) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, "Invalid user info"),
    };

    let sub = user_info
        .get("sub")
        .and_then(|value| value.as_str())
        .unwrap_or_default();
    let email = user_info
        .get("email")
        .and_then(|value| value.as_str())
        .unwrap_or_default();

    if sub.is_empty() || email.is_empty() {
        return json_error(StatusCode::INTERNAL_SERVER_ERROR, "Missing user data");
    }

    let Some(pool) = state.pool.as_ref() else {
        return json_error(
            StatusCode::INTERNAL_SERVER_ERROR,
            "Failed to create user session",
        );
    };

    let user_id = match user::lookup_or_create_user(pool, sub, email).await {
        Ok(user_id) => user_id,
        Err(_) => {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                "Failed to create user session",
            )
        }
    };

    let session_token = match session::create_session(
        pool,
        state.config.session_secret.as_bytes(),
        user_id,
        sub,
        email,
    )
    .await
    {
        Ok(token) => token,
        Err(_) => {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                "Failed to create session",
            )
        }
    };

    // Webbläsaren får sina cookies direkt från API:t - tokenen cirkulerar
    // aldrig i en URL.
    cookies.extend(session::session_cookies(
        &session_token,
        sub,
        email,
        is_secure(&headers),
    ));

    redirect_temporary_med_cookies(&frontend_root_url(), cookies)
}

/// Loggar in en utvecklingsanvändare utan Zitadel.
///
/// Endast tillgängligt när `DEV_AUTH_ENABLED=true`, vilket bara är satt i
/// utvecklingsmiljön.
#[utoipa::path(
    post,
    path = "/api/v1/auth/dev-login",
    tag = "auth",
    responses(
        (status = 307, description = "Omdirigering till gränssnittet"),
        (status = 404, description = "Dev-inloggning är inte aktiverad"),
        (status = 500, description = "Kunde inte skapa sessionen")
    )
)]
pub async fn dev_login(State(state): State<AppState>, headers: HeaderMap) -> Response {
    if env("DEV_AUTH_ENABLED").as_deref() != Some("true") {
        return json_error(StatusCode::NOT_FOUND, "Not found");
    }

    let sub = env("DEV_AUTH_SUB").unwrap_or_else(|| "dev-user".to_string());
    let email = env("DEV_AUTH_EMAIL").unwrap_or_else(|| "dev@plan.local".to_string());

    let Some(pool) = state.pool.as_ref() else {
        return json_error(StatusCode::INTERNAL_SERVER_ERROR, "Failed to create user");
    };

    let user_id = match user::lookup_or_create_user(pool, &sub, &email).await {
        Ok(user_id) => user_id,
        Err(_) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, "Failed to create user"),
    };

    let session_token = match session::create_session(
        pool,
        state.config.session_secret.as_bytes(),
        user_id,
        &sub,
        &email,
    )
    .await
    {
        Ok(token) => token,
        Err(_) => {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                "Failed to create session",
            )
        }
    };

    let cookies = session::session_cookies(&session_token, &sub, &email, is_secure(&headers));

    redirect_temporary_med_cookies(&frontend_root_url(), cookies)
}

/// Returnerar användaren bakom den aktuella sessionen.
#[utoipa::path(
    get,
    path = "/api/v1/auth/session",
    tag = "auth",
    responses(
        (status = 200, description = "Användaren bakom sessionen"),
        (status = 401, description = "Ingen giltig session")
    )
)]
pub async fn session(State(state): State<AppState>, headers: HeaderMap) -> Response {
    let Some(pool) = state.pool.as_ref() else {
        return json_error(StatusCode::UNAUTHORIZED, "Not authenticated");
    };

    let token = session::session_token_from_request(
        cookie_value(&headers, session::SESSION_COOKIE_NAME).as_deref(),
        header_value(&headers, header::AUTHORIZATION.as_str()).as_deref(),
    );

    let Some(token) = token else {
        return json_error(StatusCode::UNAUTHORIZED, "Not authenticated");
    };

    match session::validate_session(pool, state.config.session_secret.as_bytes(), &token).await {
        Ok(session) => (
            StatusCode::OK,
            Json(json!({
                "sub": session.sub,
                "email": session.email,
                "user_id": session.user_id,
            })),
        )
            .into_response(),
        Err(_) => json_error(StatusCode::UNAUTHORIZED, "Invalid or expired session"),
    }
}

/// Loggar ut - tar bort sessionen och rensar cookierna.
#[utoipa::path(
    post,
    path = "/api/v1/auth/logout",
    tag = "auth",
    responses(
        (status = 200, description = "Sessionen är borttagen")
    )
)]
pub async fn logout(State(state): State<AppState>, headers: HeaderMap) -> Response {
    let token = session::session_token_from_request(
        cookie_value(&headers, session::SESSION_COOKIE_NAME).as_deref(),
        header_value(&headers, header::AUTHORIZATION.as_str()).as_deref(),
    );

    if let (Some(pool), Some(token)) = (state.pool.as_ref(), token) {
        session::delete_session(pool, &token).await;
    }

    let cookies = session::clear_session_cookies(is_secure(&headers));

    let mut response = (StatusCode::OK, Json(json!({ "success": true }))).into_response();
    for cookie in cookies {
        if let Ok(value) = HeaderValue::from_str(&cookie) {
            response.headers_mut().append(header::SET_COOKIE, value);
        }
    }

    response
}

/// Bygger ett JSON-svar med `error`, precis som Go gör.
fn json_error(status: StatusCode, message: &str) -> Response {
    (status, Json(json!({ "error": message }))).into_response()
}

/// Bygger ett JSON-svar med `error` och `description`.
fn json_error_with_description(status: StatusCode, error: &str, description: &str) -> Response {
    (
        status,
        Json(json!({ "error": error, "description": description })),
    )
        .into_response()
}

/// En temporär redirect - samma status och body som Go:s `c.Redirect`.
fn redirect_temporary(url: &str) -> Response {
    let mut response = Response::new(axum::body::Body::empty());
    *response.status_mut() = StatusCode::TEMPORARY_REDIRECT;

    let headers = response.headers_mut();
    if let Ok(location) = HeaderValue::from_str(url) {
        headers.insert(header::LOCATION, location);
    }
    headers.insert(
        header::CONTENT_TYPE,
        HeaderValue::from_static("text/html; charset=utf-8"),
    );

    response
}

/// En temporär redirect med cookies - som Go när både cookie och redirect
/// skickas i samma svar.
fn redirect_temporary_med_cookies(url: &str, cookies: Vec<String>) -> Response {
    let mut response = redirect_temporary(url);

    for cookie in cookies {
        if let Ok(value) = HeaderValue::from_str(&cookie) {
            response.headers_mut().append(header::SET_COOKIE, value);
        }
    }

    response
}

/// Läser en miljövariabel, tom sträng räknas som satt.
fn env(namn: &str) -> Option<String> {
    std::env::var(namn).ok().filter(|value| !value.is_empty())
}

/// Pekar på gränssnittet användaren ska landa på efter inloggning.
fn frontend_base_url() -> String {
    if let Some(base) = env("FRONTEND_URL") {
        return base.trim_end_matches('/').to_string();
    }

    let domain = env("APP_DOMAIN").unwrap_or_else(|| "plan.simonbrundin.com".to_string());

    if env("WORKTREE_NAME").is_some() {
        return format!("http://{domain}");
    }

    format!("https://{domain}")
}

/// Gränssnittets rotadress - med worktree-prefix i worktree-läge.
fn frontend_root_url() -> String {
    let base = frontend_base_url();

    if let Some(worktree) = env("WORKTREE_NAME") {
        if env("FRONTEND_URL").is_none() {
            return format!("{base}/{worktree}/");
        }
    }

    format!("{base}/")
}

/// Den redirect_uri som skickas till Zitadel.
///
/// Samma värde måste användas i authorize-anropet och i tokenbytet.
fn oauth_redirect_uri() -> String {
    if let Some(uri) = env("ZITADEL_REDIRECT_URI") {
        return uri;
    }

    let domain = env("APP_DOMAIN").unwrap_or_else(|| "plan.simonbrundin.com".to_string());

    if env("WORKTREE_NAME").is_some() {
        // Dev-gatewayn äger callbacken och skickar vidare till rätt
        // worktree via `wt`-cookien.
        return format!("http://{domain}/oauth/callback");
    }

    format!("https://{domain}/api/v1/auth/callback")
}

/// Läser ett cookie-värde ur requestens Cookie-header.
fn cookie_value(headers: &HeaderMap, name: &str) -> Option<String> {
    let raw = header_value(headers, header::COOKIE.as_str())?;

    for part in raw.split(';') {
        let part = part.trim();
        if let Some((key, value)) = part.split_once('=') {
            if key.trim() == name {
                return Some(value.trim().to_string());
            }
        }
    }

    None
}

/// Läser ett header-värde.
fn header_value(headers: &HeaderMap, name: &str) -> Option<String> {
    headers
        .get(name)
        .and_then(|value| value.to_str().ok())
        .map(str::to_string)
}

/// Om cookierna ska skickas med `Secure`.
///
/// Go sätter flaggan när requesten kom över TLS eller när
/// `COOKIE_SECURE=true`.
fn is_secure(headers: &HeaderMap) -> bool {
    if std::env::var("COOKIE_SECURE").ok().as_deref() == Some("true") {
        return true;
    }

    // Axum sätter inte "Forwarded" automatiskt, men proxyn gör det.
    header_value(headers, "x-forwarded-proto")
        .map(|value| value.eq_ignore_ascii_case("https"))
        .unwrap_or(false)
}

/// Suffixet som läggs till i cookie-headern när `Secure` ska sättas.
fn secure_suffix() -> &'static str {
    if std::env::var("COOKIE_SECURE").ok().as_deref() == Some("true") {
        "; Secure"
    } else {
        ""
    }
}

/// Escape av ett värde i en query-sträng - samma som Go:s `url.QueryEscape`.
fn query_escape(value: &str) -> String {
    let mut ut = String::with_capacity(value.len());

    for byte in value.bytes() {
        match byte {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => {
                ut.push(byte as char)
            }
            b' ' => ut.push('+'),
            other => ut.push_str(&format!("%{other:02X}")),
        }
    }

    ut
}

/// Avkodar ett url-escapat värde - som Go gör när cookie läses.
///
/// Går det inte att avkoda blir värdet tomt, precis som i Go.
fn query_unescape(value: &str) -> String {
    let bytes = value.as_bytes();
    let mut ut: Vec<u8> = Vec::with_capacity(bytes.len());
    let mut index = 0;

    while index < bytes.len() {
        match bytes[index] {
            b'+' => {
                ut.push(b' ');
                index += 1;
            }
            b'%' if index + 2 < bytes.len() => {
                let hex = &value[index + 1..index + 3];
                match u8::from_str_radix(hex, 16) {
                    Ok(byte) => {
                        ut.push(byte);
                        index += 3;
                    }
                    Err(_) => return String::new(),
                }
            }
            b'%' => return String::new(),
            other => {
                ut.push(other);
                index += 1;
            }
        }
    }

    String::from_utf8_lossy(&ut).into_owned()
}

/// En slumpad sträng i URL-säkert base64-format.
///
/// Samma längd som Go använder: 64 tecken för verifieraren och 32 för
/// state.
fn random_string(length: usize) -> String {
    let mut bytes = Vec::with_capacity(length);

    while bytes.len() < length {
        bytes.extend_from_slice(uuid::Uuid::new_v4().as_bytes());
    }

    URL_SAFE_NO_PAD.encode(&bytes)[..length].to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn query_escape_matchar_go() {
        assert_eq!(
            query_escape("https://plan.simonbrundin.com/api/v1/auth/callback"),
            "https%3A%2F%2Fplan.simonbrundin.com%2Fapi%2Fv1%2Fauth%2Fcallback"
        );
        assert_eq!(query_escape("abc123"), "abc123");
        assert_eq!(query_escape("a b"), "a+b");
        assert_eq!(query_escape("a-b_c.d~e"), "a-b_c.d~e");
    }

    #[test]
    fn cookie_varden_escapas_som_gin() {
        // Gin skriver cookie-värden url-escapade.
        let json = r#"{"state":"a","verifier":"b","worktree":""}"#;
        let escaped = query_escape(json);

        assert_eq!(
            escaped,
            "%7B%22state%22%3A%22a%22%2C%22verifier%22%3A%22b%22%2C%22worktree%22%3A%22%22%7D"
        );
        assert_eq!(query_unescape(&escaped), json);
    }

    #[test]
    fn slumpade_strangar_ha_ratt_langd() {
        assert_eq!(random_string(64).len(), 64);
        assert_eq!(random_string(32).len(), 32);
        assert_ne!(random_string(32), random_string(32));
    }

    #[test]
    fn cookie_lases_ur_headern() {
        let mut headers = HeaderMap::new();
        headers.insert(
            header::COOKIE,
            HeaderValue::from_static("plan_session=abc; plan_user=def"),
        );

        assert_eq!(
            cookie_value(&headers, "plan_session"),
            Some("abc".to_string())
        );
        assert_eq!(cookie_value(&headers, "plan_user"), Some("def".to_string()));
        assert_eq!(cookie_value(&headers, "saknas"), None);
    }
}
