export interface Goal {
  id: number
  title: string
  icon: string
  created: string
  finished: string | null
  inbox: number
}

export interface ChildRelation {
  child_id: number
  order: number
  weight: number
}

export interface ParentRelation {
  goalByParentId: Goal
}

export interface GoalWithRelations extends Goal {
  childRelations: ChildRelation[]
  parentRelations: ParentRelation[]
}

export interface GoalWithWeight extends Goal {
  weight: number
}

export interface GetGoalResponse {
  goal: GoalWithRelations
  allGoals: Goal[]
}
