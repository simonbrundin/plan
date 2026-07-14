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
