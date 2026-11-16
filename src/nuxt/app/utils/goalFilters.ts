import type { Goal } from "~/types/goal";

export type StartedFilter = "all" | "started" | "not_started";

/**
 * Ett måls starttidstämpel. Fältet kan saknas helt i ett API-svar,
 * vilket betyder samma sak som null - målet är inte påbörjat.
 */
type StartedField = { started?: string | null };

/**
 * Avgör om ett mål ska visas givet det aktuella startfiltret.
 *
 * - "all"          → alla mål visas
 * - "started"      → endast mål med en satt started-tidsstämpel
 * - "not_started"  → endast mål där started är null
 */
export function matchesStartedFilter(
	goal: StartedField,
	filter: StartedFilter,
): boolean {
	// Saknas fältet (undefined) ska räknas som "inte påbörjat", inte som
	// "påbörjat" - annars matchar mål utan startdatum filtret "Påbörjade".
	const started = goal.started ?? null;
	if (filter === "all") return true;
	if (filter === "started") return started !== null;
	return started === null;
}

/**
 * Avgör om ett mål är ett grundmål, dvs. inte har någon förälder.
 *
 * Endast grundmål listas under /goals - undermål visas på sitt
 * föräldermål. Saknas has_parent (t.ex. från ett svar som inte
 * innehåller relationer) räknas målet som grundmål.
 */
export function isRootGoal(goal: Pick<Goal, "has_parent">): boolean {
	return !goal.has_parent;
}

/**
 * Avgör om ett mål fortfarande är öppet, dvs. inte är slutfört.
 *
 * Endast öppna mål listas under /goals - slutförda mål döljs.
 * Målet utan satt finished-tidsstämpel räknas som öppet.
 */
export function isUnfinishedGoal(goal: Pick<Goal, "finished">): boolean {
	return !goal.finished;
}
