import type { Goal } from "~/types/goal";

type OrderableGoal = Pick<Goal, "order" | "created">;

/**
 * Jämför två grundmål efter deras sparade ordning.
 *
 * Målen med lägre order kommer först. Målen som delar samma order (t.ex.
 * innan någon har sorterat) hamnar efter varandra med nyast först - det är
 * samma ordning som API:t gav innan orderfältet fanns.
 */
export function compareGoalOrder(a: OrderableGoal, b: OrderableGoal): number {
	const orderDiff = (a.order ?? 0) - (b.order ?? 0);
	if (orderDiff !== 0) return orderDiff;
	return new Date(b.created).getTime() - new Date(a.created).getTime();
}

/**
 * Returnerar målen sorterade efter deras sparade ordning.
 * Ursprungliga arrayen påverkas inte.
 */
export function sortGoalsByOrder<T extends OrderableGoal>(goals: T[]): T[] {
	return [...goals].sort(compareGoalOrder);
}
