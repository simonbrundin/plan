//! API-nycklar.
//!
//! Port av `src/api/internal/handlers/apikeys.go` - nycklarna lagras
//! hashed, så att en nyckel inte kan läsas ur databasen.
//!
//! Go-tjänsten svarar här med `{"error": ...}` i stället för
//! `{"message": ...}` - det är ett avtalet gränssnittet redan vänder sig
//! till, så formatet hålls.

use sha2::{Digest, Sha256};
use sqlx::postgres::PgRow;
use sqlx::{PgPool, Row};

use crate::error::AppError;

/// En API-nyckel, utan själva nyckeln.
#[derive(Debug, Clone)]
pub struct ApiKey {
    /// Nyckelns id.
    pub id: i64,
    /// Namnet.
    pub name: String,
    /// Scopet - "full" eller "agent".
    pub scope: String,
    /// När nyckeln skapades.
    pub created: chrono::DateTime<chrono::Utc>,
    /// När nyckeln sist användes.
    pub last_used: Option<chrono::DateTime<chrono::Utc>>,
    /// När nyckeln går ut.
    pub expires: Option<chrono::DateTime<chrono::Utc>>,
}

/// Läser en nyckelrad.
fn läs_nyckel(rad: &PgRow) -> ApiKey {
    ApiKey {
        id: rad.get::<i32, _>(0) as i64,
        name: rad.get::<String, _>(1),
        scope: rad.get::<String, _>(2),
        created: rad.get::<chrono::DateTime<chrono::Utc>, _>(3),
        last_used: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(4),
        expires: rad.get::<Option<chrono::DateTime<chrono::Utc>>, _>(5),
    }
}

/// Användarens nycklar, nyast först.
pub async fn list(pool: &PgPool, user_id: i64) -> Result<Vec<ApiKey>, AppError> {
    let rader = sqlx::query(
        "SELECT id, name, scope, created, last_used, expires \
         FROM api_keys WHERE user_id = $1 ORDER BY created DESC",
    )
    .bind(user_id)
    .fetch_all(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to list API keys: {err}")))?;

    Ok(rader.iter().map(läs_nyckel).collect())
}

/// Skapar en nyckel och returnerar den råa nyckeln.
///
/// Den råa nyckeln syns bara här - i databasen ligger bara hashen.
pub async fn create(
    pool: &PgPool,
    user_id: i64,
    name: &str,
    scope: &str,
) -> Result<(i64, String), AppError> {
    let rå = slumpnyckel(32);
    let hash = hash_nyckel(&rå);

    let rad = sqlx::query(
        "INSERT INTO api_keys (name, key_hash, user_id, scope) VALUES ($1, $2, $3, $4) \
         RETURNING id",
    )
    .bind(name)
    .bind(hash)
    .bind(user_id)
    .bind(scope)
    .fetch_one(pool)
    .await
    .map_err(|err| AppError::Internal(format!("failed to create API key: {err}")))?;

    Ok((rad.get::<i32, _>(0) as i64, rå))
}

/// Tar bort en nyckel.
pub async fn delete(pool: &PgPool, user_id: i64, key_id: i64) -> Result<bool, AppError> {
    let resultat = sqlx::query("DELETE FROM api_keys WHERE id = $1 AND user_id = $2")
        .bind(key_id)
        .bind(user_id)
        .execute(pool)
        .await
        .map_err(|err| AppError::Internal(format!("failed to delete API key: {err}")))?;

    Ok(resultat.rows_affected() > 0)
}

/// Hashar en nyckel med SHA-256.
///
/// Samma hash som Go-tjänsten använder - därför fungerar en nyckel som
/// skapats av den andra tjänsten.
pub fn hash_nyckel(nyckel: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(nyckel.as_bytes());
    hex::encode(hasher.finalize())
}

/// Slumpar fram en nyckel.
///
/// Teckenuppsättningen är densamma som i Go-tjänsten - alfanumeriska
/// tecken, så att nyckeln kan skrivas i en URL utan escaping.
fn slumpnyckel(längd: usize) -> String {
    const TECKEN: &[u8] = b"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";

    let slump: Vec<u8> = (0..längd)
        .map(|_| {
            // `rand` är inte beroendet - UUID:n ger slumpen här.
            let byte = uuid::Uuid::new_v4().as_bytes()[0];
            TECKEN[usize::from(byte) % TECKEN.len()]
        })
        .collect();

    String::from_utf8(slump).unwrap_or_default()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn hashen_ar_sha256_av_nyckeln() {
        // Go: sha256("abc") = ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
        assert_eq!(
            hash_nyckel("abc"),
            "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
        );
    }

    #[test]
    fn den_slumpade_nyckeln_har_riktig_langd_och_giltiga_tecken() {
        let nyckel = slumpnyckel(32);

        assert_eq!(nyckel.chars().count(), 32);
        assert!(nyckel.chars().all(|c| c.is_ascii_alphanumeric()));
    }
}
