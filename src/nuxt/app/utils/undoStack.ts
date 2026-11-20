/**
 * Undo-stapeln - logiken bakom "ångra senaste handlingen".
 *
 * Stapeln är en ring: när gränsen nås faller den äldsta posten bort.
 * Logiken är ren (inga Vue-referenser) så att den går att enhetstesta.
 */

export interface UndoEntry {
	/** Beskrivning av handlingen, t.ex. 'Markerade "Handla" som klar' */
	label: string;
	/** Återställer det som handlingen ändrade */
	undo: () => Promise<void> | void;
}

// Hur många handlingar som går att ångra.
export const UNDO_STACK_LIMIT = 25;

/** Lägger en handling överst i stapeln. */
export function pushUndoEntry(
	entries: UndoEntry[],
	entry: UndoEntry,
): UndoEntry[] {
	const next = [...entries, entry];
	return next.length > UNDO_STACK_LIMIT
		? next.slice(next.length - UNDO_STACK_LIMIT)
		: next;
}

/**
 * Tar bort den senaste handlingen ur stapeln.
 * Posten returneras så att den kan läggas tillbaka om ångrandet misslyckas.
 */
export function popUndoEntry(entries: UndoEntry[]): {
	entries: UndoEntry[];
	entry: UndoEntry | null;
} {
	if (entries.length === 0) {
		return { entries: [], entry: null };
	}

	return {
		entries: entries.slice(0, -1),
		entry: entries[entries.length - 1] ?? null,
	};
}
