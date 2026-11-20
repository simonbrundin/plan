import type { Goal } from "~/types/goal";
import { startGoalDrag } from "~/utils/calendarScheduling";

/**
 * Gör att ett mål kan dras in i kalendern.
 *
 * Målet packas i dragdata, så att kalendern kan skapa en händelse med
 * rätt titel, ikon och koppling. Drag-bilden blir hela målraden, så att
 * det går att se vilket mål som dras.
 */
export function useGoalDrag() {
	const startDrag = (event: DragEvent, goal: Goal) => {
		startGoalDrag(event, {
			id: goal.id,
			title: goal.title,
			icon: goal.icon,
			estimated_minutes: goal.estimated_minutes ?? null,
		});

		// Hela raden visas som drag-bild. Utan det skulle bara ikonen
		// följa pekaren.
		const target = event.target as HTMLElement | null;
		const row = target?.closest("[data-goal-row]") as HTMLElement | null;

		if (row && event.dataTransfer) {
			event.dataTransfer.setDragImage(row, 20, 20);
		}
	};

	return { startDrag };
}
