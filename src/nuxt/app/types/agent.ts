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
