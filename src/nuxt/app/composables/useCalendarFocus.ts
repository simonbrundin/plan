import type { CalendarEvent } from "~/composables/useCalendarApi";

/**
 * En begäran om att kalendern ska fokusera en händelse.
 *
 * Kommandopaletten kan hitta en händelse, men kalendern ligger i en egen
 * komponent. Den här stateen är handskottet mellan dem: paletten lämnar
 * över händelsen, och kalendern zoomar dit och öppnar händelsen.
 *
 * Varje begäran har ett id, så att en kalender som monteras senare kan
 * hantera samma begäran utan att den hanteras två gånger i samma
 * kalender.
 */
export interface CalendarFocusRequest {
	id: number;
	event: CalendarEvent;
}

let nextRequestId = 0;

export function useCalendarFocus() {
	const request = useState<CalendarFocusRequest | null>(
		"calendar:focus-request",
		() => null,
	);

	/** Ber kalendern att fokusera en händelse. */
	function focusEvent(event: CalendarEvent): void {
		nextRequestId += 1;
		request.value = { id: nextRequestId, event };
	}

	/** Tar bort begäran, så att den inte hanteras igen senare. */
	function clearFocus(): void {
		request.value = null;
	}

	return {
		request,
		focusEvent,
		clearFocus,
	};
}
