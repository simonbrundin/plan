//! Agentens egen databas.
//!
//! Agenten har en egen databas med samtal, meddelanden, API-nycklar för
//! leverantörerna, checkpointplaner och tokenanvändning. Den skiljer sig
//! från Plan API:ets databas - Plan äger målen, agenten äger sin egen
//! arbetsyta.

pub mod models;
pub mod repository;

use sqlx::PgPool;

/// Agentens anslutningspool.
#[derive(Clone)]
pub struct AgentDb {
    pool: PgPool,
}

impl AgentDb {
    /// Skapar anslutningen till agentens databas.
    ///
    /// Utan `DATABASE_URL` returneras `None` - då startar API:t utan
    /// databas, precis som Go-agenten gör när den inte kan ansluta.
    pub async fn connect(url: Option<&str>) -> Result<Option<Self>, sqlx::Error> {
        let Some(url) = url else {
            return Ok(None);
        };

        let pool = sqlx::postgres::PgPoolOptions::new()
            .max_connections(10)
            .min_connections(2)
            .connect(url)
            .await?;

        let db = Self { pool };
        db.migrera().await?;

        Ok(Some(db))
    }

    /// Poolen, för den som vill köra egna frågor.
    pub fn pool(&self) -> &PgPool {
        &self.pool
    }

    /// Skapar tabellerna om de saknas.
    ///
    /// Go-agenten gör samma sak - `CREATE TABLE IF NOT EXISTS` för varje
    /// tabell. Därför körs samma SQL mot samma databas.
    async fn migrera(&self) -> Result<(), sqlx::Error> {
        for sql in MIGRATIONER {
            sqlx::query(sql).execute(&self.pool).await?;
        }

        Ok(())
    }
}

/// Tabellerna agenten behöver.
///
/// SQL:en är densamma som Go-agentens `internal/api/db/schema.sql`, så
/// att båda kan köra mot samma databas.
const MIGRATIONER: &[&str] = &[
    r#"
    CREATE TABLE IF NOT EXISTS conversations (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW(),
        title VARCHAR(255) DEFAULT 'New conversation',
        model VARCHAR(100) DEFAULT 'MiniMax-M2.7',
        provider VARCHAR(50) DEFAULT 'minimax',
        goal_id BIGINT
    )
    "#,
    // Kolumnen läggs till separat - tabellen kan redan finnas i
    // databaser som kör agenten sedan tidigare.
    "ALTER TABLE conversations ADD COLUMN IF NOT EXISTS goal_id BIGINT",
    "CREATE INDEX IF NOT EXISTS idx_conversations_goal_id ON conversations(goal_id)",
    r#"
    CREATE TABLE IF NOT EXISTS messages (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
        role VARCHAR(20) NOT NULL,
        content TEXT NOT NULL,
        tool_name VARCHAR(100) DEFAULT '',
        tool_args JSONB DEFAULT '{}',
        tool_result TEXT DEFAULT '',
        is_error BOOLEAN DEFAULT FALSE,
        thinking TEXT DEFAULT '',
        tokens_used INT DEFAULT 0,
        latency_ms INT DEFAULT 0,
        created_at TIMESTAMPTZ DEFAULT NOW()
    )
    "#,
    "CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)",
    "CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at)",
    r#"
    CREATE TABLE IF NOT EXISTS api_keys (
        id SERIAL PRIMARY KEY,
        provider VARCHAR(50) NOT NULL UNIQUE,
        api_key TEXT NOT NULL,
        model VARCHAR(100) DEFAULT '',
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW()
    )
    "#,
    r#"
    CREATE TABLE IF NOT EXISTS goals (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        feature VARCHAR(500) NOT NULL,
        description TEXT DEFAULT '',
        status VARCHAR(50) DEFAULT 'pending',
        approved_by VARCHAR(255) DEFAULT '',
        approved_at TIMESTAMPTZ,
        rejected_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW()
    )
    "#,
    r#"
    CREATE TABLE IF NOT EXISTS checkpoints (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
        name VARCHAR(255) NOT NULL,
        description TEXT DEFAULT '',
        complexity VARCHAR(20) DEFAULT 'medium',
        gates JSONB DEFAULT '[]',
        tasks JSONB DEFAULT '[]',
        status VARCHAR(50) DEFAULT 'pending',
        position INT DEFAULT 0,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW(),
        completed_at TIMESTAMPTZ
    )
    "#,
    r#"
    CREATE TABLE IF NOT EXISTS checkpoint_results (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        checkpoint_id UUID NOT NULL REFERENCES checkpoints(id) ON DELETE CASCADE,
        gate_type VARCHAR(50) NOT NULL,
        status VARCHAR(50) DEFAULT 'pending',
        started_at TIMESTAMPTZ DEFAULT NOW(),
        ended_at TIMESTAMPTZ,
        duration VARCHAR(100) DEFAULT '',
        feedback TEXT DEFAULT '',
        issues JSONB DEFAULT '[]',
        summary TEXT DEFAULT '',
        created_at TIMESTAMPTZ DEFAULT NOW(),
        UNIQUE(checkpoint_id, gate_type)
    )
    "#,
    "CREATE INDEX IF NOT EXISTS idx_checkpoints_goal_id ON checkpoints(goal_id)",
    "CREATE INDEX IF NOT EXISTS idx_checkpoints_status ON checkpoints(status)",
    "CREATE INDEX IF NOT EXISTS idx_checkpoint_results_checkpoint_id ON checkpoint_results(checkpoint_id)",
    "CREATE INDEX IF NOT EXISTS idx_checkpoint_results_gate_type ON checkpoint_results(gate_type)",
    r#"
    CREATE TABLE IF NOT EXISTS token_usage (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
        provider VARCHAR(50) NOT NULL,
        model VARCHAR(100) NOT NULL,
        input_tokens INT NOT NULL DEFAULT 0,
        output_tokens INT NOT NULL DEFAULT 0,
        total_tokens INT GENERATED ALWAYS AS (input_tokens + output_tokens) STORED,
        cost DECIMAL(10, 6) DEFAULT 0,
        latency_ms INT DEFAULT 0,
        timestamp TIMESTAMPTZ DEFAULT NOW()
    )
    "#,
    "CREATE INDEX IF NOT EXISTS idx_token_usage_conversation_id ON token_usage(conversation_id)",
    "CREATE INDEX IF NOT EXISTS idx_token_usage_provider ON token_usage(provider)",
    "CREATE INDEX IF NOT EXISTS idx_token_usage_timestamp ON token_usage(timestamp)",
];
