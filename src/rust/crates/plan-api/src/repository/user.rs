//! Användare.
//!
//! Port av `LookupOrCreateUser` i `src/api/internal/middleware/auth.go`.
//! Samma ordning på sökningarna: först `sub`, sedan `email` (då uppdateras
//! `sub`), och till sist skapas användaren.

use sqlx::PgPool;

/// Slår upp en användare, eller skapar den.
///
/// `users.id` är INTEGER i databasen, så id:t läses som i32 och konverteras
/// till i64 - samma modell som i resten av API:t.
pub async fn lookup_or_create_user(
    pool: &PgPool,
    sub: &str,
    email: &str,
) -> Result<i64, sqlx::Error> {
    // Finns användaren på sub?
    match sqlx::query_scalar::<_, i32>("SELECT id FROM users WHERE sub = $1")
        .bind(sub)
        .fetch_optional(pool)
        .await
    {
        Ok(Some(id)) => return Ok(id as i64),
        Ok(None) => {}
        // Go behandlar ett databasfel som "hittades inte" - samma sökväg
        // fortsätter, men felet loggas så att det inte försvinner.
        Err(err) => tracing::warn!("kunde inte slå upp användaren på sub: {err}"),
    }

    // Finns användaren på email? Då uppdateras sub.
    match sqlx::query_scalar::<_, i32>("SELECT id FROM users WHERE email = $1")
        .bind(email)
        .fetch_optional(pool)
        .await
    {
        Ok(Some(id)) => {
            sqlx::query("UPDATE users SET sub = $1 WHERE id = $2")
                .bind(sub)
                .bind(id)
                .execute(pool)
                .await?;
            return Ok(id as i64);
        }
        Ok(None) => {}
        Err(err) => tracing::warn!("kunde inte slå upp användaren på email: {err}"),
    }

    // Användaren finns inte - skapa den.
    let id: i32 = sqlx::query_scalar("INSERT INTO users (sub, email) VALUES ($1, $2) RETURNING id")
        .bind(sub)
        .bind(email)
        .fetch_one(pool)
        .await?;

    Ok(id as i64)
}
