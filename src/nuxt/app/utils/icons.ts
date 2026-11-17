/**
 * Hantering av ikoner för kalenderhändelser.
 *
 * En ikon kan vara antingen en emoji (t.ex. "🏃") eller ett Iconify-namn
 * (t.ex. "heroicons:star"). Båda lagras i samma fält, så vi behöver kunna
 * skilja på dem vid rendering.
 *
 * Iconify-namn har alltid formen "kollektion:namn", medan emoji aldrig
 * innehåller ett kolon. Det är den skillnaden vi utnyttjar.
 */

/** Maximal längd på en ikon, i tecken. Matchar databasens VARCHAR(100). */
export const ICON_MAX_LENGTH = 100;

/** Sant om värdet är ett Iconify-namn och inte en emoji. */
export function isIconifyIcon(value?: string | null): boolean {
	return !!value && value.includes(":");
}

/** Returnerar en avkortad ikon om den är längre än tillåtet. */
export function normalizeIcon(value?: string | null): string {
	return (value ?? "").trim().slice(0, ICON_MAX_LENGTH);
}
