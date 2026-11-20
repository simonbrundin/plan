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
	// StartDate är det datum ("YYYY-MM-DD") målet tidigast får dyka upp
	// i prioriteringslistan. Null = målet får synas direkt.
	start_date?: string | null;
	// Uppskattad tidsåtgång i minuter. Null = ingen uppskattning satt.
	estimated_minutes?: number | null;
	// Platser där målet kan slutföras. Flera platser har OR-semantik:
	// målet är görbart om man är på någon av dem.
	places?: Place[];
	inbox: number;
	// Evigt mål - pågår hela livet och kan inte slutföras.
	eternal: boolean;
	// Sant om målet är avsett att arbeta på av en agent. Används för att
	// se vilka mål som ligger för en agent att ta tag i.
	agent?: boolean;
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

export interface GoalWithRelations extends Goal {
	childRelations: ChildRelation[];
	// Föräldrarna skickas som vanliga mål - samma form som i "parents".
	parentRelations: Goal[];
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

// Relationen mellan ett barn och en förälder. Ordning och vikt hör till
// relationen, inte till målet.
export interface GoalRelation {
	parent_id: number;
	child_id: number;
	order: number;
	weight: number;
}

// Hela målgraphen: målen, relationerna mellan dem och beroendena.
// Klienten räknar ut barn, föräldrar och beroenden lokalt ur grafen.
export interface GoalGraphResponse {
	goals: Goal[];
	relations: GoalRelation[];
	dependencies: GoalDependency[];
}

// Ett undermål med sin ordning, vikt och beroenden.
export interface ChildWithDeps extends GoalWithWeight {
	dependsOn?: Goal[];
	blocking?: Goal[];
}

export interface GetGoalResponse {
	goal: GoalWithRelations;
	children: ChildWithDeps[];
	parents: Goal[];
	allGoals: Goal[];
	dependencies: GoalDependency[];
	dependsOn: Goal[];
	blocking: Goal[];
}
