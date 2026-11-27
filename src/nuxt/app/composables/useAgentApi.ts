import type { AgentApproval, AgentRun, AgentRunEvent } from '~/types/agent';
import type { Goal } from '~/types/goal';
import { apiRequest } from '~/utils/api';

export function useAgentApi() {
  // Agentmålen - de mål som är markerade för en agent.
  const fetchAgentGoals = async (): Promise<Goal[]> => {
    const goals = await apiRequest<Goal[]>('goals?agent=true');
    return goals ?? [];
  };

  // Körningarna för ett mål, eller alla körningar om inget mål anges.
  const fetchAgentRuns = async (goalId?: number): Promise<AgentRun[]> => {
    const query = goalId === undefined ? '' : `?goal_id=${goalId}`;
    const runs = await apiRequest<AgentRun[]>(`agent-runs${query}`);
    return runs ?? [];
  };

  const fetchAgentRunEvents = async (runId: number): Promise<AgentRunEvent[]> => {
    const events = await apiRequest<AgentRunEvent[]>(`agent-runs/${runId}/events`);
    return events ?? [];
  };

  // Godkännandena för en körning - de som väntar på en människa.
  const fetchAgentApprovals = async (runId?: number): Promise<AgentApproval[]> => {
    if (runId === undefined) return [];
    const approvals = await apiRequest<AgentApproval[]>(`agent-runs/${runId}/approvals`);
    return approvals ?? [];
  };

  return {
    fetchAgentGoals,
    fetchAgentRuns,
    fetchAgentRunEvents,
    fetchAgentApprovals,
  };
}
