/**
 * Sortering av "Gör nu"-listan med deadline först.
 *
 * Ordningen är ett val i gränssnittet, inte en regel i API:t - därför
 * ligger den här och inte i frågan. Toggle:en "Deadline först" i /now
 * slår den på och av.
 */

/** Det som behövs för att sortera på deadline. */
export interface DueSortable {
	/** Deadlinen som ISO-datum (YYYY-MM-DD), eller null om målet saknar den. */
	due?: string | null;
}

/**
 * Sorterar målen så att de med deadline hamnar först, den närmaste fristen
 * först. Mål utan deadline behåller sin inbördes ordning.
 *
 * Sorteringen är stabil - mål med samma deadline ligger kvar i den ordning
 * de hade, så den manuella ordningen finns kvar inom gruppen.
 *
 * Deadlinedatum jämförs som text: de är ISO-datum, så lexikografisk
 * ordning är samma sak som kronologisk.
 */
export function sortByDueFirst<T extends DueSortable>(
	goals: readonly T[],
): T[] {
	return [...goals].sort((a, b) => {
		const aDue = a.due ?? null;
		const bDue = b.due ?? null;

		if (aDue && bDue) {
			if (aDue === bDue) return 0;
			return aDue < bDue ? -1 : 1;
		}
		if (aDue) return -1;
		if (bDue) return 1;
		return 0;
	});
}
