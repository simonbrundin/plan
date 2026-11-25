import type { WorkflowPlan } from '~/types/workflow'

interface WorkflowPlansResponse {
  plans: WorkflowPlan[]
}

export function useWorkflowApi() {
  async function listPlans(): Promise<WorkflowPlan[]> {
    const response = await $fetch<WorkflowPlansResponse>('/api/agent/workflow/plans')
    return response.plans ?? []
  }

  async function createPlan(feature: string, description: string): Promise<WorkflowPlan> {
    const response = await $fetch<{ plan: WorkflowPlan }>('/api/agent/workflow', {
      method: 'POST',
      body: { feature, description },
    })
    return response.plan
  }

  async function approvePlan(id: string): Promise<void> {
    await $fetch(`/api/agent/workflow/${encodeURIComponent(id)}/approve`, {
      method: 'POST',
      body: { approvedBy: 'Plan UI' },
    })
  }

  async function rejectPlan(id: string): Promise<void> {
    await $fetch(`/api/agent/workflow/${encodeURIComponent(id)}/reject`, {
      method: 'POST',
    })
  }

  return { listPlans, createPlan, approvePlan, rejectPlan }
}
