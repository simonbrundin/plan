//! Regressionstest för beroenden mellan mål.
//!
//! `goal_dependencies.id` är INTEGER i schemat, men API-modellen använder
//! i64. Om kolumnen läses som i64 ger sqlx en `ColumnDecode`-fel - och
//! med `unwrap` blir det en panic, så att anropet bara hänger.
//!
//! Testet kräver en databas och hoppas över när `DATABASE_URL` inte är
//! satt - precis som övriga databasnära tester i repot.

use plan_api::repository::goal::get_all_dependencies;
use sqlx::PgPool;

/// Returnerar en pool om `DATABASE_URL` är satt, annars hoppas testet över.
async fn pool_eller_skip() -> Option<PgPool> {
    let url = match std::env::var("DATABASE_URL") {
        Ok(url) if !url.is_empty() => url,
        _ => {
            eprintln!("DATABASE_URL saknas - testet hoppas över");
            return None;
        }
    };

    let pool = PgPool::connect(&url)
        .await
        .expect("kunde inte ansluta till databasen");
    Some(pool)
}

/// Lägger in ett mål med ett beroende och läser beroendet tillbaka.
#[tokio::test]
async fn beroende_läses_med_integer_id() {
    let Some(pool) = pool_eller_skip().await else {
        return;
    };

    let user_id: i32 = sqlx::query_scalar(
        "INSERT INTO users (sub, email) VALUES ('rust-beroende', 'rust-beroende@example.com') \
         RETURNING id",
    )
    .fetch_one(&pool)
    .await
    .expect("kunde inte skapa användare");

    let mål_a: i32 = sqlx::query_scalar(
        "INSERT INTO goals (title, status_id, created) \
         VALUES ('[test] beroende a', 1, NOW()) RETURNING id",
    )
    .fetch_one(&pool)
    .await
    .expect("kunde inte skapa mål");

    let mål_b: i32 = sqlx::query_scalar(
        "INSERT INTO goals (title, status_id, created) \
         VALUES ('[test] beroende b', 1, NOW()) RETURNING id",
    )
    .fetch_one(&pool)
    .await
    .expect("kunde inte skapa mål");

    sqlx::query("INSERT INTO user_goals (user_id, goal_id) VALUES ($1, $2), ($1, $3)")
        .bind(user_id)
        .bind(mål_a)
        .bind(mål_b)
        .execute(&pool)
        .await
        .expect("kunde inte koppla mål till användare");

    sqlx::query("INSERT INTO goal_dependencies (goal_id, depends_on_id) VALUES ($1, $2)")
        .bind(mål_a)
        .bind(mål_b)
        .execute(&pool)
        .await
        .expect("kunde inte skapa beroende");

    // Läser tillbaka - får inte panika, och id:t ska vara rätt.
    let beroenden = get_all_dependencies(&pool, user_id as i64)
        .await
        .expect("kunde inte läsa beroenden");

    assert_eq!(beroenden.len(), 1, "ett beroende ska läsas tillbaka");
    assert_eq!(beroenden[0].goal_id, mål_a as i64);
    assert_eq!(beroenden[0].depends_on_id, mål_b as i64);
    assert!(beroenden[0].id > 0, "beroendets id ska läsas som ett tal");

    // Städar upp testdatan.
    sqlx::query("DELETE FROM goal_dependencies WHERE goal_id = $1")
        .bind(mål_a)
        .execute(&pool)
        .await
        .ok();
    sqlx::query("DELETE FROM user_goals WHERE goal_id = $1 OR goal_id = $2")
        .bind(mål_a)
        .bind(mål_b)
        .execute(&pool)
        .await
        .ok();
    sqlx::query("DELETE FROM goals WHERE id = $1 OR id = $2")
        .bind(mål_a)
        .bind(mål_b)
        .execute(&pool)
        .await
        .ok();
    sqlx::query("DELETE FROM users WHERE id = $1")
        .bind(user_id)
        .execute(&pool)
        .await
        .ok();
}
