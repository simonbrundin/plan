//! Zitadel-token.
//!
//! Port av JWT-delen av `src/api/internal/middleware/auth.go`: JWKS hämtas
//! och cachas, JWT verifieras med RS256, och opaka tokener valideras via
//! introspektion.

use std::sync::{Mutex, OnceLock};
use std::time::{Duration, Instant};

use jsonwebtoken::{decode, decode_header, Algorithm, DecodingKey, Validation};
use serde::Deserialize;

/// Standardsdomänen för Zitadel.
pub const DEFAULT_ZITADEL_DOMAIN: &str = "auth.simonbrundin.com";
/// Hur länge JWKS cachas.
const JWKS_TTL: Duration = Duration::from_secs(3600);

/// Ett svar från Zitadels introspektionsendpoint.
#[derive(Debug, Clone, Deserialize)]
pub struct IntrospectionResponse {
    #[serde(default)]
    pub active: bool,
    #[serde(default)]
    pub sub: String,
    #[serde(default)]
    pub email: String,
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub username: String,
    #[serde(default)]
    pub client_id: String,
    #[serde(default)]
    pub scope: String,
    #[serde(default)]
    pub exp: i64,
    #[serde(default)]
    pub iat: i64,
}

/// Identiteten i ett validerat token.
#[derive(Debug, Clone, PartialEq)]
pub struct Identity {
    pub sub: String,
    pub email: String,
}

/// En public nyckel från JWKS.
#[derive(Debug, Clone, Deserialize)]
struct Jwk {
    #[serde(default)]
    kty: String,
    #[serde(default)]
    kid: String,
    #[serde(default)]
    n: String,
    #[serde(default)]
    e: String,
}

/// JWKS-svaret.
#[derive(Debug, Clone, Deserialize)]
struct JwkSet {
    #[serde(default)]
    keys: Vec<Jwk>,
}

/// Claims i ett Zitadel-JWT.
#[derive(Debug, Clone, Deserialize)]
struct Claims {
    sub: String,
    #[serde(default)]
    email: String,
}

/// Cachen för JWKS.
fn jwks_cache() -> &'static Mutex<Option<(Instant, JwkSet)>> {
    static CACHE: OnceLock<Mutex<Option<(Instant, JwkSet)>>> = OnceLock::new();
    CACHE.get_or_init(|| Mutex::new(None))
}

/// Hämtar JWKS, med cache.
async fn get_jwks(domain: &str) -> Result<JwkSet, String> {
    {
        let cache = jwks_cache().lock().expect("JWKS-cachen är inte låst");
        if let Some((hämtad, uppsättning)) = cache.as_ref() {
            if hämtad.elapsed() < JWKS_TTL {
                return Ok(uppsättning.clone());
            }
        }
    }

    let url = format!("https://{domain}/oauth/v2/keys");
    let response = reqwest::get(&url).await.map_err(|err| err.to_string())?;
    let uppsättning: JwkSet = response.json().await.map_err(|err| err.to_string())?;

    let mut cache = jwks_cache().lock().expect("JWKS-cachen är inte låst");
    *cache = Some((Instant::now(), uppsättning.clone()));

    Ok(uppsättning)
}

/// Validerar ett JWT mot Zitadels JWKS.
///
/// Returnerar `None` om tokenet inte är ett giltigt JWT - anroparen
/// faller då tillbaka på introspektionen.
pub async fn validate_jwt(token: &str, domain: &str) -> Option<Identity> {
    let header = decode_header(token).ok()?;
    let kid = header.kid.clone()?;

    let uppsättning = get_jwks(domain).await.ok()?;
    let nyckel = uppsättning
        .keys
        .iter()
        .find(|nyckel| nyckel.kid == kid && (nyckel.kty == "RSA" || nyckel.kty.is_empty()))?;

    let decoding = DecodingKey::from_rsa_components(&nyckel.n, &nyckel.e).ok()?;
    let mut validation = Validation::new(Algorithm::RS256);
    validation.validate_aud = false;
    // Go kräver inte att tokenet har en utgångstid - den kontrolleras bara
    // om den finns.
    validation.required_spec_claims.clear();

    let data = decode::<Claims>(token, &decoding, &validation).ok()?;
    let claims = data.claims;

    Some(Identity {
        sub: claims.sub,
        email: claims.email,
    })
}

/// Validerar ett opakt token via Zitadels introspektionsendpoint.
pub async fn introspect_token(token: &str, domain: &str) -> Result<IntrospectionResponse, String> {
    let url = format!("https://{domain}/oauth/v2/introspect");

    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(10))
        .build()
        .map_err(|err| err.to_string())?;

    let mut request = client.post(&url).form(&[("token", token)]);

    let client_id = std::env::var("ZITADEL_CLIENT_ID").unwrap_or_default();
    if !client_id.is_empty() {
        let client_secret = std::env::var("ZITADEL_CLIENT_SECRET").unwrap_or_default();
        request = request.basic_auth(client_id, Some(client_secret));
    }

    let response = request.send().await.map_err(|err| err.to_string())?;

    response
        .json::<IntrospectionResponse>()
        .await
        .map_err(|err| err.to_string())
}

/// Identiteten ur ett introspektionssvar.
///
/// Om e-post saknas används användarnamnet - samma fallback som Go.
pub fn identity_from_introspection(svar: &IntrospectionResponse) -> Identity {
    let email = if svar.email.is_empty() {
        svar.username.clone()
    } else {
        svar.email.clone()
    };

    Identity {
        sub: svar.sub.clone(),
        email,
    }
}

/// Domänen agenten autentiserar mot.
pub fn zitadel_domain() -> String {
    std::env::var("ZITADEL_DOMAIN")
        .ok()
        .filter(|domain| !domain.is_empty())
        .unwrap_or_else(|| DEFAULT_ZITADEL_DOMAIN.to_string())
}
