export type WorkflowPlanStatus = 'pending' | 'approved' | 'rejected' | 'completed'
export type CheckpointStatus = 'pending' | 'in_progress' | 'passed' | 'failed' | 'skipped'

export interface GateResult {
  type: string
  status: string
  feedback?: string
  summary?: string
  issues?: Array<{ severity: string; title: string; description: string }>
}

export interface Checkpoint {
  id: string
  name: string
  description: string
  complexity: string
  gates: string[]
  tasks: string[]
  status: CheckpointStatus
  gateResults?: Record<string, GateResult>
}

export interface WorkflowPlan {
  id: string
  feature: string
  description: string
  checkpoints: Checkpoint[]
  status: WorkflowPlanStatus
  approvedBy?: string
  createdAt: string
  updatedAt: string
}

export function checkpointStatusLabel(status: CheckpointStatus): string {
  switch (status) {
    case 'in_progress': return 'Pågår'
    case 'passed': return 'Klar'
    case 'failed': return 'Misslyckad'
    case 'skipped': return 'Hoppad över'
    default: return 'Väntar'
  }
}

export function workflowStatusLabel(status: WorkflowPlanStatus): string {
  switch (status) {
    case 'approved': return 'Godkänd'
    case 'rejected': return 'Avvisad'
    case 'completed': return 'Klar'
    default: return 'Väntar på granskning'
  }
}
