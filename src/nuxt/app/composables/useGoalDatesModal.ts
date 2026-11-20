import type { Goal } from "~/types/goal";

export interface GoalDatesPayload {
	/** Startdatumet ("YYYY-MM-DD") eller null. */
	startDate: string | null;
	/** Due-datumet ("YYYY-MM-DD") eller null. */
	dueDate: string | null;
}

// Målet popupen gäller. Null = popupen är stängd.
//
// State ligger utanför funktionen (som leader-läget) så att popupen kan
// renderas i layouten och öppnas från alla sidor.
const target = ref<Goal | null>(null);
const savedRevision = ref(0);

/** Stänger popupen direkt, även när den anropas från popup-komponenten själv. */
export function closeGoalDatesModal() {
	target.value = null;
}

/**
 * Popupen som sätter både startdatum och due-datum för ett mål.
 *
 * Användaren öppnar den med leader + d. Den sparar bara de fält som
 * faktiskt ändrats - annars skulle ett tomt fält nollställa det andra.
 *
 * Varje ändring läggs på undo-stapeln, så att Z ångra tillbaka den.
 */
export function useGoalDatesModal() {
	const { setGoalDue, setGoalStartDate } = useGoalApi();
	const { undoDates } = useGoalUndo();
	const goalsStore = useGoalsStore();

	const isOpen = computed(() => target.value !== null);

	function openFor(goal: Goal) {
		target.value = goal;
	}

	function close() {
		closeGoalDatesModal();
	}

	/**
	 * Sparar datumen. Returnerar true om något faktiskt ändrades, så att
	 * anropande sida kan ladda om sin lista.
	 */
	async function save(payload: GoalDatesPayload): Promise<boolean> {
		const goal = target.value;
		if (!goal) return false;

		const previousDue = goal.due ?? null;
		const previousStartDate = goal.start_date ?? null;

		const dueChanged = payload.dueDate !== previousDue;
		const startDateChanged = payload.startDate !== previousStartDate;

		if (!dueChanged && !startDateChanged) {
			close();
			return false;
		}

		try {
			if (dueChanged) {
				await setGoalDue(goal.id, payload.dueDate);
			}
			if (startDateChanged) {
				await setGoalStartDate(goal.id, payload.startDate);
			}
			undoDates(goal, previousStartDate, previousDue);
			goalsStore.updateGoal(goal.id, {
				due: payload.dueDate,
				start_date: payload.startDate,
			});
		} catch (error) {
			console.error("Failed to save goal dates:", error);
			return false;
		}

		close();
		savedRevision.value++;
		return true;
	}

	return { target, isOpen, savedRevision: readonly(savedRevision), openFor, close, save };
}
