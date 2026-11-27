import type { CalendarEvent } from "~/composables/useCalendarApi";

/**
 * Sökindex för kalenderhändelserna.
 *
 * Kommandopaletten behöver hitta händelser, men den får inte ladda data
 * varje gång den öppnas. Händelserna hämtas därför en gång och delas via
 * `useState`, så att både paletten och kalendern kan använda samma
 * uppsättning.
 */
export function useCalendarSearch() {
	const events = useState<CalendarEvent[]>("calendar:search:events", () => []);
	const isLoaded = useState("calendar:search:loaded", () => false);

	const { loadAllCalendarEvents } = useCalendarApi();

	/**
	 * Hämtar händelserna om de inte redan är inlästa.
	 *
	 * Ett fel ignoreras - paletten ska fungera även om kalendern inte
	 * svarar, då saknas bara händelserna i sökresultatet.
	 */
	async function load(force = false): Promise<void> {
		if (isLoaded.value && !force) return;

		try {
			events.value = await loadAllCalendarEvents();
			isLoaded.value = true;
		} catch (error) {
			console.warn("Kunde inte läsa kalenderhändelser för sökning:", error);
		}
	}

	return {
		events,
		isLoaded: readonly(isLoaded),
		load,
	};
}
