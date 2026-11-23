// Agentkörningens faser - följer agentens livscykel.
export type AgentRunPhase =
  | 'define'
  | 'plan'
  | 'execute'
  | 'verify'
  | 'completed';

// Agentkörningens tillstånd. "blocked" betyder att agenten väntar på
// människan - till exempel ett godkännande.
export type AgentRunStatus =
  | 'queued'
  | 'running'
  | 'blocked'
  | 'done'
  | 'failed';

export interface AgentRun {
  id: number;
  goal_id: number;
  phase: AgentRunPhase;
  status: AgentRunStatus;
  // Sammanfattning av det senaste agentmeddelandet.
  summary: string;
  // Länk till resultatet - en PR eller ett bokningsutkast.
  artifact_url: string | null;
  // Kostnad för körningen hittills.
  cost_usd: number | null;
  // Koppling till en konversation i agent-tjänsten.
  conversation_id: string | null;
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
  goal_id: number;
  title: string;
  description: string;
  created_at: string;
  attempts: number;
  feedback?: string;
  status?: string;
  conversation_id?: string;
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
