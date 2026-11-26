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
pub async fn validate_api_key(
    pool: Option<&sqlx::PgPool>,
    api_key: &str,
) -> Result<ApiKeyIdentity, String> {
    if let Some(agent_key) = std::env::var("AGENT_API_KEY")
        .ok()
        .filter(|k| !k.is_empty())
    {
        if api_key == agent_key {
            return Ok(ApiKeyIdentity {
                user_id: 1,
                scope: SCOPE_AGENT.to_string(),
            });
        }
    }

    if let Some(bypass) = std::env::var("API_KEY_BYPASS")
        .ok()
        .filter(|k| !k.is_empty())
    {
        if api_key == bypass {
            return Ok(ApiKeyIdentity {
                user_id: 1,
                scope: SCOPE_FULL.to_string(),
            });
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
    .await
    .map_err(|err| format!("databasfel: {err}"))?;

    match row {
        Some(row) => Ok(ApiKeyIdentity {
            user_id: row.get::<i64, _>(0),
            scope: row.get::<String, _>(1),
        }),
        // Nyckeln hittades inte - det är inte ett databasfel.
        None => Err("API key not found".to_string()),
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
