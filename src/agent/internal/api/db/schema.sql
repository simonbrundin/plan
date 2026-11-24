-- Schema for LLM API

CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    title VARCHAR(255) DEFAULT 'New conversation',
    model VARCHAR(100) DEFAULT 'MiniMax-M2.7',
    provider VARCHAR(50) DEFAULT 'minimax'
);

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
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at);

-- API Keys for LLM providers
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL UNIQUE,
    api_key TEXT NOT NULL,
    model VARCHAR(100) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Goals table (CheckpointPlan)
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
);

-- Checkpoints table
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
);

-- Checkpoint Results table (GateResult)
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
);

-- Indexes for new tables
CREATE INDEX IF NOT EXISTS idx_checkpoints_goal_id ON checkpoints(goal_id);
CREATE INDEX IF NOT EXISTS idx_checkpoints_status ON checkpoints(status);
CREATE INDEX IF NOT EXISTS idx_checkpoint_results_checkpoint_id ON checkpoint_results(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_checkpoint_results_gate_type ON checkpoint_results(gate_type);

-- Token Usage tracking
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
);

-- Indexes for token usage
CREATE INDEX IF NOT EXISTS idx_token_usage_conversation_id ON token_usage(conversation_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_provider ON token_usage(provider);
CREATE INDEX IF NOT EXISTS idx_token_usage_timestamp ON token_usage(timestamp);
