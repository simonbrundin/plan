//! Sessioner.
//!
//! Port av `src/api/internal/middleware/sessions.go`. Tokenformaten är
//! byte-kompatibel med Go-tjänsten, så en cookie som skapats av Go-tjänsten
//! fungerar i Rust-tjänsten och vice versa - det behövs när båda körs
//! parallellt under migreringen.

use base64::engine::general_purpose::URL_SAFE_NO_PAD;
use base64::Engine;
use hmac::{Hmac, Mac};
use serde::{Deserialize, Serialize};
use sha2::Sha256;

type HmacSha256 = Hmac<Sha256>;

/// En användars session.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Session {
    pub user_id: i64,
    pub sub: String,
    pub email: String,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub expires_at: chrono::DateTime<chrono::Utc>,
}

/// Nyckeln i httpOnly-cookien.
pub const SESSION_COOKIE_NAME: &str = "plan_session";
/// Läsbar hint-cookie för klienten (sub + email).
///
/// Den är inte auktoritativ - servern validerar sessionen själv.
pub const USER_COOKIE_NAME: &str = "plan_user";

/// Hur länge en session håller.
pub const SESSION_LIFETIME_HOURS: i64 = 24;
/// Hur länge cookien håller i webbläsaren.
pub const COOKIE_MAX_AGE_SECONDS: i64 = 60 * 60 * 24 * 7;

/// Payloaden i en sessionstoken.
///
/// Fälten ligger i alfabetisk ordning - Go marshalerar en `map`, som
/// sorterar nycklarna. Samma ordning ger samma bytes och därmed samma
/// signatur.
#[derive(Debug, Serialize, Deserialize)]
struct SessionPayload {
    created_at: String,
    email: String,
    expires_at: String,
    sub: String,
    user_id: i64,
}

impl From<&Session> for SessionPayload {
    fn from(session: &Session) -> Self {
        Self {
            created_at: session
                .created_at
                .to_rfc3339_opts(chrono::SecondsFormat::Secs, true),
            email: session.email.clone(),
            expires_at: session
                .expires_at
                .to_rfc3339_opts(chrono::SecondsFormat::Secs, true),
            sub: session.sub.clone(),
            user_id: session.user_id,
        }
    }
}

/// Skapar en signerad sessionstoken.
///
/// Tokenen innehåller sessionen (base64url) och en HMAC-signatur över
/// payloaden, separerad med en punkt.
pub fn sign_session(session: &Session, secret: &[u8]) -> String {
    let payload = SessionPayload::from(session);
    let json = serde_json::to_vec(&payload).expect("sessionen går att serialisera");
    let encoded = URL_SAFE_NO_PAD.encode(&json);

    let mut mac = HmacSha256::new_from_slice(secret).expect("HMAC tar alla nycklar");
    mac.update(&json);
    let signature = URL_SAFE_NO_PAD.encode(mac.finalize().into_bytes());

    format!("{encoded}.{signature}")
}

/// Läser och verifierar en sessionstoken.
///
/// Signaturen kontrolleras i konstant tid - en angripare ska inte kunna
/// gissa signaturen byte för byte.
pub fn parse_session(token: &str, secret: &[u8]) -> Result<Session, String> {
    let (encoded, signature) = split_token(token).ok_or("invalid token format")?;

    let json = URL_SAFE_NO_PAD
        .decode(encoded)
        .map_err(|_| "invalid token encoding")?;

    let signature = URL_SAFE_NO_PAD
        .decode(signature)
        .map_err(|_| "invalid token encoding")?;

    let mut mac = HmacSha256::new_from_slice(secret).expect("HMAC tar alla nycklar");
    mac.update(&json);
    // Verifieringen sker i konstant tid.
    mac.verify_slice(&signature)
        .map_err(|_| "invalid session token")?;

    let payload: SessionPayload =
        serde_json::from_slice(&json).map_err(|_| "invalid session data")?;

    Ok(Session {
        user_id: payload.user_id,
        sub: payload.sub,
        email: payload.email,
        created_at: parse_time(&payload.created_at)?,
        expires_at: parse_time(&payload.expires_at)?,
    })
}

/// Delar en token vid sista punkten.
fn split_token(token: &str) -> Option<(&str, &str)> {
    let index = token.rfind('.')?;
    Some((&token[..index], &token[index + 1..]))
}

/// Tolkar en RFC3339-tid.
fn parse_time(value: &str) -> Result<chrono::DateTime<chrono::Utc>, String> {
    chrono::DateTime::parse_from_rfc3339(value)
        .map(|tid| tid.with_timezone(&chrono::Utc))
        .map_err(|_| "invalid session data".to_string())
}

/// Skapar en session och sparar den i databasen.
pub async fn create_session(
    pool: &sqlx::PgPool,
    secret: &[u8],
    user_id: i64,
    sub: &str,
    email: &str,
) -> Result<String, sqlx::Error> {
    let now = chrono::Utc::now();
    let session = Session {
        user_id,
        sub: sub.to_string(),
        email: email.to_string(),
        created_at: now,
        expires_at: now + chrono::Duration::hours(SESSION_LIFETIME_HOURS),
    };

    let token = sign_session(&session, secret);

    sqlx::raw_sql(
        r#"
        CREATE TABLE IF NOT EXISTS sessions (
            token      TEXT PRIMARY KEY,
            user_id   BIGINT NOT NULL,
            sub       TEXT NOT NULL,
            email     TEXT NOT NULL,
            created_at TIMESTAMPTZ NOT NULL,
            expires_at TIMESTAMPTZ NOT NULL
        )
        "#,
    )
    .execute(pool)
    .await?;

    sqlx::query(
        r#"
        INSERT INTO sessions (token, user_id, sub, email, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (token) DO UPDATE SET expires_at = $6
        "#,
    )
    .bind(&token)
    .bind(session.user_id)
    .bind(&session.sub)
    .bind(&session.email)
    .bind(session.created_at)
    .bind(session.expires_at)
    .execute(pool)
    .await?;

    Ok(token)
}

/// Validerar en sessionstoken.
///
/// Tokenen signaturkontrolleras, ses till att den inte gått ut och att den
/// finns kvar i databasen.
pub async fn validate_session(
    pool: &sqlx::PgPool,
    secret: &[u8],
    token: &str,
) -> Result<Session, String> {
    let session = parse_session(token, secret)?;

    if chrono::Utc::now() > session.expires_at {
        // Städa bort den utgångna sessionen.
        let _ = sqlx::query("DELETE FROM sessions WHERE token = $1")
            .bind(token)
            .execute(pool)
            .await;
        return Err("session expired".to_string());
    }

    let found: bool = sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM sessions WHERE token = $1 AND expires_at > NOW())",
    )
    .bind(token)
    .fetch_one(pool)
    .await
    .map_err(|err| format!("session not found: {err}"))?;

    if !found {
        return Err("session not found".to_string());
    }

    Ok(session)
}

/// Tar bort en session.
pub async fn delete_session(pool: &sqlx::PgPool, token: &str) {
    let _ = sqlx::query("DELETE FROM sessions WHERE token = $1")
        .bind(token)
        .execute(pool)
        .await;
}

/// Tar bort utgångna sessioner.
pub async fn cleanup_expired_sessions(pool: &sqlx::PgPool) {
    let _ = sqlx::query("DELETE FROM sessions WHERE expires_at < NOW()")
        .execute(pool)
        .await;
}

/// Läser sessionstoken ur requesten.
///
/// Först från sessioncookien (webbläsaren), sedan från
/// `Authorization: Bearer` (API-klienter).
pub fn session_token_from_request(
    cookie: Option<&str>,
    authorization: Option<&str>,
) -> Option<String> {
    if let Some(token) = cookie {
        if !token.is_empty() {
            return Some(token.to_string());
        }
    }

    let header = authorization?;
    let (scheme, token) = header.split_once(' ')?;
    if !scheme.eq_ignore_ascii_case("bearer") {
        return None;
    }
    let token = token.trim();
    if token.is_empty() {
        None
    } else {
        Some(token.to_string())
    }
}

/// Bygger sessioncookierna.
///
/// `plan_session` är httpOnly och bär tokenen, `plan_user` är läsbar för
/// klienten och innehåller bara sub och email.
pub fn session_cookies(token: &str, sub: &str, email: &str, secure: bool) -> Vec<String> {
    let user_data = serde_json::json!({ "sub": sub, "email": email });
    let encoded_user = base64::engine::general_purpose::STANDARD.encode(user_data.to_string());

    // Go skriver attributen i ordningen Path, Max-Age, HttpOnly, Secure,
    // SameSite - samma ordning ger samma header.
    vec![
        format!(
            "{}={}; Path=/; Max-Age={}; HttpOnly{}; SameSite=Lax",
            SESSION_COOKIE_NAME,
            token,
            COOKIE_MAX_AGE_SECONDS,
            if secure { "; Secure" } else { "" }
        ),
        format!(
            "{}={}; Path=/; Max-Age={}{}; SameSite=Lax",
            USER_COOKIE_NAME,
            encoded_user,
            COOKIE_MAX_AGE_SECONDS,
            if secure { "; Secure" } else { "" }
        ),
    ]
}

/// Bygger cookie-headers som rensar sessionen.
pub fn clear_session_cookies(secure: bool) -> Vec<String> {
    // Go skriver "Max-Age=0" när max age är negativ - samma som här.
    vec![
        format!(
            "{}=; Path=/; Max-Age=0; HttpOnly{}; SameSite=Lax",
            SESSION_COOKIE_NAME,
            if secure { "; Secure" } else { "" }
        ),
        format!(
            "{}=; Path=/; Max-Age=0{}; SameSite=Lax",
            USER_COOKIE_NAME,
            if secure { "; Secure" } else { "" }
        ),
    ]
}

#[cfg(test)]
mod tests {
    use super::*;

    fn session() -> Session {
        Session {
            user_id: 42,
            sub: "378032824856347117".to_string(),
            email: "simonbrundin@gmail.com".to_string(),
            created_at: chrono::DateTime::parse_from_rfc3339("2026-11-26T10:00:00Z")
                .unwrap()
                .with_timezone(&chrono::Utc),
            expires_at: chrono::DateTime::parse_from_rfc3339("2026-11-27T10:00:00Z")
                .unwrap()
                .with_timezone(&chrono::Utc),
        }
    }

    #[test]
    fn token_round_trip() {
        let secret = b"hemligt";
        let token = sign_session(&session(), secret);

        let parsed = parse_session(&token, secret).expect("tokenen ska gå att läsa");

        assert_eq!(parsed.user_id, 42);
        assert_eq!(parsed.sub, "378032824856347117");
        assert_eq!(parsed.email, "simonbrundin@gmail.com");
        assert_eq!(parsed.created_at, session().created_at);
        assert_eq!(parsed.expires_at, session().expires_at);
    }

    #[test]
    fn manipulerad_token_avvisas() {
        let secret = b"hemligt";
        let token = sign_session(&session(), secret);

        let (data, _) = split_token(&token).unwrap();
        let mut delar: Vec<&str> = token.split('.').collect();
        delar[1] = " fingerad-signatur";
        assert!(parse_session(&delar.join("."), secret).is_err());
        assert!(parse_session(data, secret).is_err());
    }

    #[test]
    fn fel_hemlighet_avvisas() {
        let token = sign_session(&session(), b"hemligt");

        assert!(parse_session(&token, b"annat").is_err());
    }

    #[test]
    fn token_ur_requesten() {
        assert_eq!(
            session_token_from_request(Some("cookie-token"), None),
            Some("cookie-token".to_string())
        );
        assert_eq!(
            session_token_from_request(None, Some("Bearer abc")),
            Some("abc".to_string())
        );
        assert_eq!(
            session_token_from_request(None, Some("bearer abc")),
            Some("abc".to_string())
        );
        assert_eq!(session_token_from_request(None, Some("Basic abc")), None);
        assert_eq!(session_token_from_request(None, None), None);
    }

    /// Tokenen måste vara identisk med Go-tjänstens - annars fungerar inte
    /// en cookie som skapats av den andra tjänsten.
    #[test]
    fn tokenen_ar_byte_kompatibel_med_go() {
        let session = session();
        let token = sign_session(&session, b"hemligt");

        let (encoded, _) = split_token(&token).unwrap();
        let json = URL_SAFE_NO_PAD.decode(encoded).unwrap();
        let text = String::from_utf8(json).unwrap();

        // Go marshalerar en map - nycklarna sorteras alfabetiskt.
        assert_eq!(
            text,
            r#"{"created_at":"2026-11-26T10:00:00Z","email":"simonbrundin@gmail.com","expires_at":"2026-11-27T10:00:00Z","sub":"378032824856347117","user_id":42}"#
        );
    }
}
