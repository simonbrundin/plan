//! API-nycklar.
//!
//! Port av `src/api/internal/middleware/apikey.go`. Nyckeln skickas i
//! headern `X-API-Key` (eller som query-parameter `api_key`), hashas med
//! SHA-256 och slås upp i databasen.

use sha2::{Digest, Sha256};
use sqlx::Row;

/// Headern API-nyckeln skickas i.
pub const API_KEY_HEADER: &str = "x-api-key";
/// Query-parametern API-nyckeln kan skickas i.
pub const API_KEY_QUERY_PARAM: &str = "api_key";

/// Nyckelns scope: full behörighet.
pub const SCOPE_FULL: &str = "full";
/// Nyckelns scope: bara agentens egna anrop.
pub const SCOPE_AGENT: &str = "agent";

/// Kontrollerar att ett scope är känt.
///
/// En okänd scope ger inga rättigheter - därför är listan sluten.
pub fn giltig_scope(scope: &str) -> bool {
    scope == SCOPE_FULL || scope == SCOPE_AGENT
}

/// Resultatet av en nyckelvalidering.
#[derive(Debug, Clone, PartialEq)]
pub struct ApiKeyIdentity {
    /// Användarens id.
    pub user_id: i64,
    /// Nyckelns scope.
    pub scope: String,
}

/// Hashar en API-nyckel med SHA-256.
pub fn hash_api_key(api_key: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(api_key.as_bytes());
    hex::encode(hasher.finalize())
}

/// Validerar en API-nyckel och ger användarens id och scope.
///
/// Agentens nyckel sätts i `AGENT_API_KEY` och har agent-scope - gränsen
/// lever i API:t, inte i en prompt. `API_KEY_BYPASS` är en utvecklingsnyckel
/// med full behörighet.
///
/// `Ok(None)` betyder att nyckeln inte finns - Go svarar då med ett 401
/// utan extra meddelande. `Err` betyder att databasen inte är uppkopplad,
/// vilket är det enda fall där Go skickar med ett meddelande.
pub async fn validate_api_key(
    pool: Option<&sqlx::PgPool>,
    api_key: &str,
) -> Result<Option<ApiKeyIdentity>, String> {
    if let Some(agent_key) = std::env::var("AGENT_API_KEY")
        .ok()
        .filter(|k| !k.is_empty())
    {
        if api_key == agent_key {
            return Ok(Some(ApiKeyIdentity {
                user_id: 1,
                scope: SCOPE_AGENT.to_string(),
            }));
        }
    }

    if let Some(bypass) = std::env::var("API_KEY_BYPASS")
        .ok()
        .filter(|k| !k.is_empty())
    {
        if api_key == bypass {
            return Ok(Some(ApiKeyIdentity {
                user_id: 1,
                scope: SCOPE_FULL.to_string(),
            }));
        }
    }

    let Some(pool) = pool else {
        return Err("database not connected".to_string());
    };

    let key_hash = hash_api_key(api_key);

    let row = sqlx::query(
        r#"
        SELECT COALESCE(user_id, 1), COALESCE(scope, 'full')
        FROM api_keys
        WHERE key_hash = $1
        AND (expires IS NULL OR expires > NOW())
        "#,
    )
    .bind(&key_hash)
    .fetch_optional(pool)
    .await;

    // Go behandlar ett databasfel som "nyckeln hittades inte" - samma
    // svar, men felet loggas så att det inte försvinner.
    let row = match row {
        Ok(row) => row,
        Err(err) => {
            tracing::warn!("kunde inte slå upp API-nyckeln: {err}");
            return Ok(None);
        }
    };

    match row {
        Some(row) => {
            // `user_id` är INTEGER i databasen, så den läses som i32.
            // `try_get` används i stället för `get` - en typmismatch ska
            // ge ett fel, inte en panik.
            let user_id: i32 = row.try_get(0).map_err(|err| format!("databasfel: {err}"))?;
            let scope: String = row.try_get(1).map_err(|err| format!("databasfel: {err}"))?;
            Ok(Some(ApiKeyIdentity {
                user_id: user_id as i64,
                scope,
            }))
        }
        // Nyckeln hittades inte - det är inte ett databasfel.
        None => Ok(None),
    }
}

/// Genererar en ny API-nyckel och sparar hashen.
///
/// Den råa nyckeln syns bara här - i klienten.
pub async fn generate_api_key(
    pool: &sqlx::PgPool,
    name: &str,
    user_id: Option<i64>,
    scope: &str,
) -> Result<String, sqlx::Error> {
    let raw = format!("plan_{}", uuid::Uuid::new_v4().simple());
    let key_hash = hash_api_key(&raw);

    sqlx::query(
        r#"
        INSERT INTO api_keys (name, key_hash, user_id, scope)
        VALUES ($1, $2, $3, $4)
        "#,
    )
    .bind(name)
    .bind(&key_hash)
    .bind(user_id)
    .bind(scope)
    .execute(pool)
    .await?;

    Ok(raw)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn hashen_ar_sha256_hex() {
        // Samma hash som Go-tjänsten producerar.
        assert_eq!(
            hash_api_key("hemligt"),
            "033e887dd2004f157d1913b2e3e60d84eaa23678e605b63c2550efe7c372b2b8"
        );
    }

    #[test]
    fn hashen_ar_deterministisk() {
        assert_eq!(hash_api_key("a"), hash_api_key("a"));
        assert_ne!(hash_api_key("a"), hash_api_key("b"));
    }
}
