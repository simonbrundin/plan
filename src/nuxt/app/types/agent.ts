// Agentkörningens faser - följer agentens livscykel.
export type AgentRunPhase =
  | 'define'
  | 'plan'
  | 'execute'
  | 'verify'
  | 'completed';

// Agentkörningens tillstånd. "waiting_approval" betyder att agenten väntar på
// människan - till exempel ett godkännande.
//
// Namnen är de som Plan API:t använder. De definieras i
// src/go/internal/models/agent_run.go och läses från
// GET /api/v1/agent-run-lifecycle.
export type AgentRunStatus =
  | 'queued'
  | 'running'
  | 'waiting_approval'
  | 'succeeded'
  | 'failed'
  | 'cancelled';

// Vad som startade en körning: användare, schema, webhook eller agent.
export type AgentRunTrigger = 'user' | 'schedule' | 'webhook' | 'agent';

// Vilken sorts agent som utförde körningen.
export type AgentProfile = 'coding' | 'life';

export interface AgentRunEvent {
  id: number;
  agent_run_id: number;
  type: string;
  phase?: AgentRunPhase;
  payload?: Record<string, unknown>;
  created_at: string;
}

export interface AgentRun {
  id: number;
  goal_id: number;
  phase: AgentRunPhase;
  status: AgentRunStatus;
  // Vad som startade körningen.
  trigger: AgentRunTrigger;
  // Vem eller vad som initierade körningen, t.ex. "user:1" eller
  // "worker:agent-worker-goal-502".
  initiator?: string | null;
  // Vilken sorts agent som körde.
  agent_profile: AgentProfile;
  // Sammanfattning av det senaste agentmeddelandet.
  summary: string;
  // Vad körningen ledde till.
  result?: string | null;
  // Länk till resultatet - en PR eller ett bokningsutkast.
  artifact_url: string | null;
  // Kostnad för körningen hittills.
  cost_usd: number | null;
  // Koppling till en konversation i agent-tjänsten.
  conversation_id: string | null;
  // När körningen startade och slutade.
  started_at: string | null;
  finished_at: string | null;
  created_at: string;
  updated_at: string;
}

// Ett meddelande i chatten med agenten.
export interface AgentMessage {
  id: string;
  conversation_id?: string;
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
  tool_name?: string;
  tool_args?: string;
  tool_result?: string;
  is_error?: boolean;
  thinking?: string;
  created_at?: string;
}

// Ett förslag från agenten som väntar på ett beslut i Plan.
export interface AgentApproval {
  // Godkännandets id i Plan - det är det beslutet tas på.
  id: number;
  agent_run_id: number;
  goal_id: number;
  title: string;
  description: string;
  created_at: string;
  feedback?: string;
  status?: string;
  conversation_id?: string;
  // Vad som föreslogs - exakt de ändringar och det test som ska köras.
  payload?: {
    changes?: Array<{ path: string; content: string }>;
    test_command?: string;
    pre_verification?: { passed: boolean; output?: string };
  };
}

// En konversation hos agent-API:t.
export interface AgentConversation {
  id: string;
  title: string;
  model: string;
  provider: string;
  created_at?: string;
  updated_at?: string;
}

// Kostnaden för en agentkörning. Plan räknar fram den ur en
// versionsstyrd prislista.
export interface AgentRunUsage {
  id: number;
  agent_run_id: number;
  goal_id: number;
  agent_id?: string;
  provider?: string;
  model?: string;
  model_version?: string;
  input_tokens: number;
  output_tokens: number;
  cache_read_tokens: number;
  tool_calls: number;
  tool_names?: string[];
  cost_usd: number;
  currency: string;
  // Vilken prislista som räknades fram kostnaden med.
  price_list_version: string;
  // "price_list" eller "fallback" - fallback betyder schablonpris.
  price_source: string;
  // "reported" eller "incomplete".
  status: string;
  created_at: string;
  started_at?: string;
  finished_at?: string;
}

// En sammanfattad kostnad.
export interface AgentUsageSummary {
  runs: number;
  total_cost_usd: number;
  input_tokens: number;
  output_tokens: number;
  cache_read_tokens: number;
  tool_calls: number;
  currency: string;
  by_model?: Array<{
    provider: string;
    model: string;
    runs: number;
    cost_usd: number;
  }>;
}

// En budgetgräns för agentkostnad.
export interface AgentBudget {
  id: number;
  name: string;
  period: string;
  hard_limit_usd: number;
  warning_limit_usd?: number;
  reservation_per_run_usd: number;
  agent_id?: string;
  active: boolean;
  created_at: string;
}

// En budgets aktuella läge.
export interface BudgetStatus {
  budget_id: number;
  name: string;
  period: string;
  hard_limit_usd: number;
  spent_usd: number;
  reserved_usd: number;
  exceeded: boolean;
  warning: boolean;
}
