/**
 * Färglösning för kalenderhändelser.
 *
 * Regeln är enkel: en händelse som har en egen färg vinner, annars ärver
 * den färgen från kalendern den ligger i, och händelser utan kalender får
 * standardfärgen.
 *
 * Logiken ligger här, utanför komponenterna, så att den kan testas och
 * användas på flera ställen (kort, prickar och formulär).
 */

/** Färgen som används för händelser utan kalender och utan egen färg. */
export const DEFAULT_EVENT_COLOR = "#6366f1";

export interface ColoredCalendar {
	id: number;
	color: string;
}

export interface EventColorSource {
	/** Händelsens egen färg, om den har en. */
	color?: string | null;
	/** Vilken kalender händelsen ligger i. */
	calendar_id?: number | null;
}

/**
 * resolveEventColor returnerar den färg en händelse ska visas med.
 */
export function resolveEventColor(
	event: EventColorSource,
	calendars: Map<number, ColoredCalendar> | ColoredCalendar[],
): string {
	if (event.color) return event.color;

	const calendarId = event.calendar_id;
	if (calendarId === undefined || calendarId === null) return DEFAULT_EVENT_COLOR;

	const calendar = Array.isArray(calendars)
		? calendars.find((candidate) => candidate.id === calendarId)
		: calendars.get(calendarId);

	return calendar?.color || DEFAULT_EVENT_COLOR;
}

/**
 * isEventHidden berättar om en händelse ska döljas, vilket den gör om
 * kalendern den ligger i är dold.
 *
 * Händelser utan kalender visas alltid - de tillhör ingen kalender som
 * kan stängas av.
 */
export function isEventHidden(
	event: Pick<EventColorSource, "calendar_id">,
	hiddenCalendarIds: Iterable<number>,
): boolean {
	const calendarId = event.calendar_id;
	if (calendarId === undefined || calendarId === null) return false;

	for (const hidden of hiddenCalendarIds) {
		if (hidden === calendarId) return true;
	}

	return false;
}
