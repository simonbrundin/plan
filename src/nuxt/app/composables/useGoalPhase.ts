import type { Goal } from "~/types/goal";
import { GOAL_PHASES, nextPhase, type GoalPhase } from "~/utils/goalPhase";

/**
 * Byte av målfas (status) från tangentbordet.
 *
 * Samma tangent fungerar för ett enskilt markerat mål och för flera
 * markerade mål samtidigt - den som anropar composablen bestämmer vilka
 * mål som är aktuella.
 *
 * Eviga mål hoppas över: de har ingen fas (fasen visas inte för dem).
 */
export function useGoalPhase() {
	const goalsStore = useGoalsStore();
	const { updateGoalStatus } = useGoalApi();

	// Faserna behövs för att översätta fasnamn till status-id. De laddas
	// en gång och cachas i store.
	async function ensureStatuses() {
		try {
			await goalsStore.loadStatuses();
		} catch (error) {
			console.error("Kunde inte hämta målfaserna:", error);
		}
	}

	// Målets nuvarande fas. En okänd status räknas som första fasen.
	function phaseOf(goal: Goal): GoalPhase {
		const status = goalsStore.statuses.find((entry) => entry.id === goal.status_id);
		if (status && (GOAL_PHASES as readonly string[]).includes(status.name)) {
			return status.name as GoalPhase;
		}
		return GOAL_PHASES[0];
	}

	function statusIdFor(phase: GoalPhase): number | undefined {
		return goalsStore.statuses.find((entry) => entry.name === phase)?.id;
	}

	// Sätter en fas på ett mål och sparar den.
	async function setPhase(goal: Goal, phase: GoalPhase): Promise<boolean> {
		const statusId = statusIdFor(phase);
		if (statusId === undefined) return false;

		await updateGoalStatus(goal.id, statusId);
		goal.status_id = statusId;
		goalsStore.updateGoal(goal.id, { status_id: statusId });
		return true;
	}

	/**
	 * Flyttar ett eller flera mål en fas framåt (eller bakåt).
	 *
	 * Returnerar antalet mål som faktiskt bytte fas.
	 */
	async function cyclePhases(
		goals: Goal[],
		direction: 1 | -1 = 1,
	): Promise<number> {
		if (goals.length === 0) return 0;
		await ensureStatuses();

		const targets = goals.filter((goal) => !goal.eternal);
		let changed = 0;

		await Promise.all(
			targets.map(async (goal) => {
				const next = nextPhase(phaseOf(goal), direction);
				if (await setPhase(goal, next)) {
					changed += 1;
				}
			}),
		);

		return changed;
	}

	return { ensureStatuses, phaseOf, setPhase, cyclePhases };
}
