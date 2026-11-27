import type { Place } from "~/types/place";

// Ett mål som platsväljaren gäller. Det behöver bara id och titel.
export interface PlaceTarget {
	id: number;
	title: string;
}

// Målet platsväljaren gäller. Null = väljaren är stängd.
//
// State ligger utanför funktionen (som tagg- och datumväljaren) så att
// väljaren kan renderas i layouten och öppnas från alla sidor.
const target = ref<PlaceTarget | null>(null);
const draft = ref<Place[]>([]);

// Bumpas varje gång en platslista har sparats, så att sidan kan skriva
// tillbaka de nya platserna i sitt eget urval.
const savedRevision = ref(0);

// Senast sparade platslista - sidan använder den för att uppdatera
// sitt eget urval.
const lastSaved = ref<{ id: number; places: Place[] } | null>(null);

/**
 * Platsväljaren som sätter platser på ett mål.
 *
 * Öppnas med leader + p på alla sidor där ett mål är markerat. Den
 * renderas i layouten, så sidor behöver bara anropa openFor().
 */
export function useGoalPlacePicker() {
	const { places: available, fetchPlaces, fetchGoalPlaces, setGoalPlaces } =
		usePlaces();
	const goalsStore = useGoalsStore();

	const isOpen = computed(() => target.value !== null);

	/** Öppnar väljaren för ett mål och hämtar dess nuvarande platser. */
	async function openFor(goal: PlaceTarget): Promise<void> {
		try {
			// Platslistan kan vara stale - hämta den så att nya platser syns.
			if (available.value.length === 0) {
				await fetchPlaces().catch(() => {});
			}
			draft.value = await fetchGoalPlaces(goal.id);
			target.value = goal;
		} catch (error) {
			console.error("Failed to load places for goal:", error);
		}
	}

	function close(): void {
		target.value = null;
		draft.value = [];
	}

	/**
	 * Sparar platserna på målet.
	 * Returnerar true om sparandet lyckades.
	 */
	async function save(places: Place[]): Promise<boolean> {
		const goal = target.value;
		if (!goal) return false;

		try {
			const saved = await setGoalPlaces(
				goal.id,
				places.map((place) => place.id),
			);
			draft.value = saved;
			lastSaved.value = { id: goal.id, places: saved };
			// Håll grafen i synk så att platserna syns på andra sidor.
			goalsStore.updateGoal(goal.id, { places: saved });
			savedRevision.value++;
			return true;
		} catch (error) {
			console.error("Failed to save places for goal:", error);
			return false;
		}
	}

	return {
		target: readonly(target),
		// Utan readonly - sidorna skriver tillbaka listan på sitt eget
		// urval när platserna har sparats.
		draft,
		available,
		isOpen,
		savedRevision: readonly(savedRevision),
		lastSaved,
		openFor,
		close,
		save,
	};
}
