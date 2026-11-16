import type { Goal } from "~/types/goal";

export type StartedFilter = "all" | "started" | "not_started";

/**
 * Avgör om ett mål ska visas givet det aktuella startfiltret.
 *
 * - "all"          → alla mål visas
 * - "started"      → endast mål med en satt started-tidsstämpel
 * - "not_started"  → endast mål där started är null
 */
export function matchesStartedFilter(
	goal: Pick<Goal, "started">,
	filter: StartedFilter,
): boolean {
	if (filter === "all") return true;
	if (filter === "started") return goal.started !== null;
	return goal.started === null;
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
