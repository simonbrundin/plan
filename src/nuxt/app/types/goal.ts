export interface Goal {
	id: number;
	title: string;
	// En längre fritext om målet. Titeln är kort - beskrivningen
	// får plats för allt som inte passar i titeln.
	description: string;
	icon: string;
	created: string;
	started: string | null;
	finished: string | null;
	// Due är det datum ("YYYY-MM-DD") målet ska vara klart.
	// Null = inget datum satt.
	due: string | null;
	inbox: number;
	// Evigt mål - pågår hela livet och kan inte slutföras.
	eternal: boolean;
	// Sant om målet har en förälder. Endast grundmål (has_parent = false)
	// listas under /goals.
	has_parent?: boolean;
	// Målets placering bland grundmål. Grundmål sorteras på /goals efter
	// detta värde (stigande).
	order?: number;
	priorityOrder?: number;
}

export interface ChildRelation {
	child_id: number;
	order: number;
	weight: number;
}

export interface ParentRelation {
	goalByParentId: Goal;
}

export interface GoalWithRelations extends Goal {
	childRelations: ChildRelation[];
	parentRelations: ParentRelation[];
}

export interface GoalWithWeight extends Goal {
	weight: number;
	order: number;
}

export interface GoalDependency {
	id: number;
	goal_id: number;
	depends_on_id: number;
	created: string;
}

export interface GetGoalResponse {
	goal: GoalWithRelations;
	children: GoalWithWeight[];
	allGoals: Goal[];
	dependencies: GoalDependency[];
	dependsOn: Goal[];
	blocking: Goal[];
}
