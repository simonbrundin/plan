import type { Place } from './place';

export type RecurrenceMode = 'schedule' | 'push';

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
	// Platser där målet kan slutföras. Flera platser har OR-semantik:
	// målet är görbart om man är på någon av dem.
	places?: Place[];
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
	// Upprepningsregeln för ett återkommande mål (RRULE).
	// Null = målet är inte återkommande.
	recurrence_rule?: string | null;
	// Schemaläggningsstrategi: "schedule" (håll schemat) eller
	// "push" (skjuta fram nästa förekomst).
	recurrence_mode?: RecurrenceMode | null;
	// Datumet ("YYYY-MM-DD") för nästa förekomst som ska slutföras.
	recurrence_next?: string | null;
	// Antalet slutförda förekomster.
	recurrence_count?: number;
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
