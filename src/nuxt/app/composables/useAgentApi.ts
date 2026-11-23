import type { AgentRun } from '~/types/agent';
import type { Goal } from '~/types/goal';

// Alla anrop går via serverns proxy, som skickar med sessioncookien.
const apiProxy = async <T>(path: string, options: RequestInit = {}): Promise<T> => {
  const response = await $fetch<T>(`/api/proxy/${path}`, {
    ...options,
  });
  return response;
};

export function useAgentApi() {
  // Agentmålen - de mål som är markerade för en agent.
  const fetchAgentGoals = async (): Promise<Goal[]> => {
    const goals = await apiProxy<Goal[]>('goals?agent=true');
    return goals ?? [];
  };

  // Körningarna för ett mål, eller alla körningar om inget mål anges.
  const fetchAgentRuns = async (goalId?: number): Promise<AgentRun[]> => {
    const query = goalId === undefined ? '' : `?goal_id=${goalId}`;
    const runs = await apiProxy<AgentRun[]>(`agent-runs${query}`);
    return runs ?? [];
  };

  return {
    fetchAgentGoals,
    fetchAgentRuns,
  };
}
