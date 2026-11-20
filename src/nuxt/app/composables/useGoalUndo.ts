import type { Goal } from "~/types/goal";
import type { RecurrenceMode } from "~/utils/recurrence";

/**
 * Registrerar målshandlingar på undo-stapeln.
 *
 * Varje hjälpfunktion sparar det gamla värdet så att ångrandet kan
 * skicka tillbaka exakt det som var innan handlingen.
 */
export function useGoalUndo() {
	const { push } = useUndoStack();
	const {
		toggleGoalFinished,
		toggleGoalStarted,
		setGoalEternal,
		setGoalDue,
		setGoalEstimate,
		setGoalRecurrence,
		undoGoalOccurrence,
	} = useGoalApi();
	const goalsStore = useGoalsStore();

	// Målen i sidomenyn och på målsidan ligger i store - den måste
	// hämta om när något har ångras.
	async function refreshStore() {
		try {
			await goalsStore.loadGraph(true);
		} catch (error) {
			console.error("Failed to refresh goals after undo:", error);
		}
	}

	/** Målet markerades som klart (eller inte). */
	function undoFinished(goal: Goal, previousFinished: string | null) {
		push({
			label: previousFinished
				? `Markerade "${goal.title}" som klar`
				: `Markerade "${goal.title}" som inte klar`,
			undo: async () => {
				await toggleGoalFinished(goal.id, previousFinished);
				await refreshStore();
			},
		});
	}

	/** En förekomst av ett återkommande mål slutfördes. */
	function undoOccurrence(goal: Goal) {
		push({
			label: `Slutförde en förekomst av "${goal.title}"`,
			undo: async () => {
				// API:t tar bort den senast slutförda förekomsten. Det
				// stämmer så länge ingen annan förekomst slutfördes efteråt.
				await undoGoalOccurrence(goal.id);
				await refreshStore();
			},
		});
	}

	/** Målet markerades som pågående (eller inte). */
	function undoStarted(goal: Goal, previousStarted: string | null) {
		push({
			label: previousStarted
				? `Markerade "${goal.title}" som pågående`
				: `Markerade "${goal.title}" som inte pågående`,
			undo: async () => {
				await toggleGoalStarted(goal.id, previousStarted);
				await refreshStore();
			},
		});
	}

	/** Evigt-märket sattes på eller togs bort. */
	function undoEternal(goal: Goal, previousEternal: boolean) {
		push({
			label: previousEternal
				? `Markerade "${goal.title}" som evigt`
				: `Tog bort evigt-markeringen på "${goal.title}"`,
			undo: async () => {
				await setGoalEternal(goal.id, previousEternal);
				await refreshStore();
			},
		});
	}

	/** Due-datumet sattes eller rensades. */
	function undoDue(goal: Goal, previousDue: string | null) {
		push({
			label: previousDue
				? `Satte due-datum ${previousDue} på "${goal.title}"`
				: `Tog bort due-datumet på "${goal.title}"`,
			undo: async () => {
				await setGoalDue(goal.id, previousDue);
				await refreshStore();
			},
		});
	}

	/** Den uppskattade tidsåtgången sattes eller rensades. */
	function undoEstimate(goal: Goal, previousMinutes: number | null) {
		push({
			label: previousMinutes
				? `Satte tidsåtgång ${previousMinutes} min på "${goal.title}"`
				: `Tog bort tidsåtgången på "${goal.title}"`,
			undo: async () => {
				await setGoalEstimate(goal.id, previousMinutes);
				await refreshStore();
			},
		});
	}

	/** Upprepningen sattes eller togs bort. */
	function undoRecurrence(
		goal: Goal,
		previousRule: string | null | undefined,
		previousMode: RecurrenceMode | null | undefined,
	) {
		push({
			label: previousRule
				? `Satte upprepning på "${goal.title}"`
				: `Tog bort upprepningen på "${goal.title}"`,
			undo: async () => {
				await setGoalRecurrence(
					goal.id,
					previousRule ?? null,
					previousMode ?? null,
				);
				await refreshStore();
			},
		});
	}

	return {
		undoFinished,
		undoOccurrence,
		undoStarted,
		undoEternal,
		undoDue,
		undoEstimate,
		undoRecurrence,
	};
}
