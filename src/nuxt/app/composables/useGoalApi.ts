import type { FetchOptions } from "ofetch";

import type { Goal, GoalDependency, GoalGraphResponse, GoalStatus } from "~/types/goal";
import { apiRequestWithAuth } from "~/utils/api";
import type { RecurrenceMode } from "~/utils/recurrence";

const FALLBACK_GOAL_ICON = "lucide:check-square";
const ICON_LOOKUP_TIMEOUT_MS = 5000;

// Alla anrop går direkt mot API:t, som skickar med sessioncookien.
const apiRequest = async <T>(path: string, options: FetchOptions = {}): Promise<T> => {
	return await apiRequestWithAuth<T>(path, options);
};

export function useGoalApi() {
	const updateGoalTitle = async (goalId: number, title: string) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { title },
		});
	};

	const updateGoalIcon = async (goalId: number, icon: string) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { icon },
		});
	};

	// Sätter målets beskrivning (lång fritext). Tom sträng rensar den.
	const updateGoalDescription = async (
		goalId: number,
		description: string,
	) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { description },
		});
	};

	const updateGoalStatus = async (goalId: number, statusId: number) => {
		await apiRequest(`goals/${goalId}/status`, {
			method: "PATCH",
			body: { status_id: statusId },
		});
	};

	const loadGoalStatuses = async (): Promise<GoalStatus[]> => {
		return await apiRequest<GoalStatus[]>("statuses");
	};

	const toggleGoalStarted = async (goalId: number, started: string | null) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { started },
		});
	};

	const toggleGoalFinished = async (
		goalId: number,
		finished: string | null,
	) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { finished },
		});
	};

	// Sätter målets due-datum ("YYYY-MM-DD"). Null rensar datumet.
	const setGoalDue = async (goalId: number, due: string | null) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { due },
		});
	};

	// Sätter målets startdatum ("YYYY-MM-DD") - det datum då målet tidigast
	// får dyka upp i prioriteringslistan. Null rensar datumet.
	const setGoalStartDate = async (goalId: number, startDate: string | null) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { start_date: startDate },
		});
	};

	// Sätter målets uppskattade tidsåtgång i minuter. Null rensar den.
	const setGoalEstimate = async (goalId: number, minutes: number | null) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { estimated_minutes: minutes },
		});
	};

	// Markerar ett mål som evigt (eller upphäver evigt-flaggan).
	// Ett evigt mål kan inte slutföras - API:t avvisar det.
	const setGoalEternal = async (goalId: number, eternal: boolean) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { eternal },
		});
	};

	// Markerar ett mål som ett mål för en agent (eller upphäver
	// agent-flaggan).
	const setGoalAgent = async (goalId: number, agent: boolean) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { agent },
		});
	};

	// Markerar ett mål som ett mål som ska göras nu (eller upphäver
	// nu-flaggan). Sådana mål syns på "Gör nu"-sidan.
	const setGoalNow = async (goalId: number, now: boolean) => {
		await apiRequest(`goals/${goalId}`, {
			method: "PATCH",
			body: { now },
		});
	};

	// Sätter målets upprepningsregel och schemaläggningsstrategi.
	// En nullad regel tar bort upprepningen så att målet blir vanligt igen.
	//
	// startDate är det datum serien startar på - det är INTE målets
	// startdatum (det sätts med setGoalStartDate).
	const setGoalRecurrence = async (
		goalId: number,
		rule: string | null,
		mode?: RecurrenceMode | null,
		startDate?: string,
	): Promise<Goal> => {
		return await apiRequest<Goal>(`goals/${goalId}`, {
			method: "PATCH",
			body: {
				recurrence_rule: rule,
				recurrence_mode: mode,
				recurrence_start_date: startDate,
			},
		});
	};

	// Slutför den aktuella förekomsten av ett återkommande mål.
	//
	// occurrenceDate är den förekomst som slutförs. Skickas den samma
	// förekomst igen ändras ingenting - anropet är en dubblett.
	const completeGoalOccurrence = async (
		goalId: number,
		occurrenceDate?: string,
		completedOn?: string,
	): Promise<Goal> => {
		return await apiRequest<Goal>(`goals/${goalId}/completions`, {
			method: "POST",
			body: {
				occurrence_date: occurrenceDate,
				completed_on: completedOn,
			},
		});
	};

	// Avbockar den senast slutförda förekomsten av ett återkommande mål.
	const undoGoalOccurrence = async (goalId: number): Promise<Goal> => {
		return await apiRequest<Goal>(`goals/${goalId}/completions`, {
			method: "DELETE",
		});
	};

	const deleteGoal = async (goalId: number) => {
		await apiRequest(`goals/${goalId}`, {
			method: "DELETE",
		});
	};

	const addParentRelation = async (childId: number, parentId: number) => {
		await apiRequest(`goals/relations`, {
			method: "POST",
			body: { childId, parentId },
		});
	};

	const removeParentRelation = async (childId: number, parentId: number) => {
		await apiRequest(`goals/relations`, {
			method: "DELETE",
			body: { childId, parentId },
		});
	};

	const addChildRelation = async (
		childId: number,
		parentId: number,
		order: number,
	) => {
		await apiRequest(`goals/relations`, {
			method: "POST",
			body: { childId, parentId, order },
		});
	};

	const updateGoalOrder = async (
		parentId: number,
		childId: number,
		order: number,
	) => {
		await apiRequest(`goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, order },
		});
	};

	// Sätter en ny ordning på alla undermål i ett svep (en request, en transaktion)
	const reorderChildren = async (
		parentId: number,
		childIds: number[],
	) => {
		await apiRequest(`goals/relations/reorder`, {
			method: "PATCH",
			body: { parentId, childIds },
		});
	};

	// Sätter en ny ordning på grundmålen (mål utan förälder) i ett svep
	const reorderGoals = async (goalIds: number[]) => {
		await apiRequest(`goals/reorder`, {
			method: "PATCH",
			body: { goalIds },
		});
	};

	const reorderPriorityGoals = async (goalIds: number[]) => {
		await apiRequest(`goals/priority/reorder`, {
			method: "PATCH",
			body: { goalIds },
		});
	};

	const updateGoalWeight = async (
		parentId: number,
		childId: number,
		weight: number,
	) => {
		await apiRequest(`goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, weight },
		});
	};

	const updateGoalWeights = async (
		parentId: number,
		weights: Array<{ childId: number; weight: number }>,
	) => {
		await apiRequest(`goals/relations/weights`, {
			method: "PATCH",
			body: { parentId, weights },
		});
	};

	const setGoalWeight = async (
		childId: number,
		parentId: number,
		weight: number,
	) => {
		await apiRequest(`goals/relations`, {
			method: "POST",
			body: { childId, parentId, weight },
		});
	};

	const loadAllGoals = async (): Promise<Goal[]> => {
		return await apiRequest<Goal[]>(`goals`);
	};

	// Hämta hela målgraphen: målen, relationerna och beroendena.
	//
	// Grafen räcker för att bygga varje målsida - barn, föräldrar och
	// beroenden räknas ut lokalt. Därför behövs bara en fråga när man
	// navigerar mellan mål.
	const loadGoalGraph = async (): Promise<GoalGraphResponse> => {
		return await apiRequest<GoalGraphResponse>(`goals/graph`);
	};

	const findIconForGoalTitle = async (title: string): Promise<string | undefined> => {
		const query = title.trim();
		if (query.length < 2) return undefined;

		try {
			const response = await $fetch<{ results?: Array<{ full?: string }> }>("/api/icons", {
				query: { q: query, translate: "true" },
				timeout: ICON_LOOKUP_TIMEOUT_MS,
			});
			return response.results?.[0]?.full || undefined;
		} catch (error) {
			// Ikonvalet får inte hindra att målet skapas om översättning, MiniMax
			// eller Iconify är otillgängligt. En checkbox visar samtidigt att
			// målet saknar en semantiskt vald ikon.
			console.warn("Kunde inte hitta en ikon för målet:", error);
			return FALLBACK_GOAL_ICON;
		}
	};

	const createGoal = async (
		title: string,
		statusId?: number,
		icon?: string,
		recurrenceRule?: string,
	): Promise<Goal> => {
		const body: {
			title: string;
			status_id?: number;
			icon?: string;
			recurrence_rule?: string;
		} = { title };
		if (statusId) {
			body.status_id = statusId;
		}
		body.icon = icon ?? await findIconForGoalTitle(title);
		if (recurrenceRule) {
			body.recurrence_rule = recurrenceRule;
		}

		return await apiRequest<Goal>(`goals`, {
			method: "POST",
			body,
		});
	};

	const addDependency = async (
		goalId: number,
		dependsOnId: number,
	): Promise<GoalDependency> => {
		return await apiRequest<GoalDependency>(`goals/dependencies`, {
			method: "POST",
			body: { goalId, dependsOnId },
		});
	};

	const removeDependency = async (goalId: number, dependsOnId: number) => {
		await apiRequest(`goals/dependencies`, {
			method: "DELETE",
			body: { goalId, dependsOnId },
		});
	};

	return {
		updateGoalTitle,
		updateGoalIcon,
		updateGoalDescription,
		updateGoalStatus,
		loadGoalStatuses,
		toggleGoalStarted,
		toggleGoalFinished,
		setGoalDue,
		setGoalStartDate,
		setGoalEstimate,
		setGoalEternal,
		setGoalAgent,
		setGoalNow,
		setGoalRecurrence,
		completeGoalOccurrence,
		undoGoalOccurrence,
		deleteGoal,
		addParentRelation,
		removeParentRelation,
		addChildRelation,
		updateGoalOrder,
		reorderChildren,
		reorderGoals,
		reorderPriorityGoals,
		updateGoalWeight,
		updateGoalWeights,
		setGoalWeight,
		createGoal,
		loadAllGoals,
		loadGoalGraph,
		addDependency,
		removeDependency,
	};
}
