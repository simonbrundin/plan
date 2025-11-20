export interface Goal {
  id: number
  title: string
  icon: string
  created: string
  finished: string | null
}

export interface ChildRelation {
  child_id: number
  order: number
}

export interface ParentRelation {
  goalByParentId: Goal
}

export interface GoalWithRelations extends Goal {
  childRelations: ChildRelation[]
  parentRelations: ParentRelation[]
}

export interface GetGoalResponse {
  goal: GoalWithRelations
  allGoals: Goal[]
}
