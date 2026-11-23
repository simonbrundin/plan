import type { Tag } from "~/types/tag";

// Ett mål som taggväljaren gäller. Det behöver bara id, titel och de
// taggar som redan sitter på målet.
export interface TagTarget {
	id: number;
	title: string;
	tags?: string[];
}

// Målet taggväljaren gäller. Null = väljaren är stängd.
//
// State ligger utanför funktionen (som datum-popupen) så att väljaren kan
// renderas i layouten och öppnas från alla sidor.
const target = ref<TagTarget | null>(null);
const draft = ref<string[]>([]);

// Bumpas varje gång en tagglista har sparats, så att sidan kan hämta
// de nya taggarna till sitt eget urval.
const savedRevision = ref(0);

// Senast sparade tagglista - sidan använder den för att uppdatera
// sitt eget urval.
const lastSaved = ref<{ id: number; tags: string[] } | null>(null);

/**
 * Taggväljaren som sätter taggar på ett mål.
 *
 * Öppnas med leader + t på alla sidor där ett mål är markerat. Den
 * renderas i layouten, så sidor behöver bara anropa openFor().
 */
export function useGoalTagPicker() {
	const { tags: available, fetchTags, fetchGoalTags, setGoalTags } = useTags();
	const goalsStore = useGoalsStore();

	const isOpen = computed(() => target.value !== null);

	/** Öppnar väljaren för ett mål och hämtar dess nuvarande taggar. */
	async function openFor(goal: TagTarget): Promise<void> {
		try {
			draft.value = await fetchGoalTags(goal.id);
			target.value = goal;
			// Tagglistan kan vara stale - hämta den så att nya taggar syns.
			await fetchTags().catch(() => {});
		} catch (error) {
			console.error("Failed to load tags for goal:", error);
		}
	}

	function close(): void {
		target.value = null;
		draft.value = [];
	}

	/**
	 * Sparar tagglistan på målet.
	 * Returnerar true om sparandet lyckades.
	 */
	async function save(names: string[]): Promise<boolean> {
		const goal = target.value;
		if (!goal) return false;

		try {
			const saved = await setGoalTags(goal.id, names);
			draft.value = saved;
			lastSaved.value = { id: goal.id, tags: saved };
			// Håll grafen i synk så att taggar syns på andra sidor.
			goalsStore.updateGoal(goal.id, { tags: saved });
			// En ny tagg kan ha skapats - hämta listan igen.
			await fetchTags().catch(() => {});
			savedRevision.value++;
			return true;
		} catch (error) {
			console.error("Failed to save tags for goal:", error);
			return false;
		}
	}

	return {
		target: readonly(target),
		draft: readonly(draft),
		available,
		isOpen,
		savedRevision: readonly(savedRevision),
		lastSaved: readonly(lastSaved),
		openFor,
		close,
		save,
	};
}
