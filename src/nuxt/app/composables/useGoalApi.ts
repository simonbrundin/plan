import type { Goal, GoalDependency } from "~/types/goal";

interface GoalWithWeight extends Goal {
	weight: number;
	order: number;
}

interface GoalData {
	goal: Goal;
	children: GoalWithWeight[];
	parents: Goal[];
	allGoals: Goal[];
	dependencies: GoalDependency[];
	dependsOn: Goal[];
	blocking: Goal[];
	statusHistory?: StatusUpdate[];
}

interface StatusUpdate {
	id: number;
	goal_id: number;
	from_status_id?: number;
	to_status_id: number;
	changed_at: string;
	from_status_name?: string;
	to_status_name: string;
}

// Custom error for authentication failures
export class AuthenticationError extends Error {
	constructor(message = "Authentication required") {
		super(message);
		this.name = "AuthenticationError";
	}
}

// Proxy API calls through our server route which attaches session cookie
const apiProxy = async <T>(path: string, options: RequestInit = {}): Promise<T> => {
	try {
		const response = await $fetch<T>(`/api/proxy/${path}`, {
			...options,
		});
		return response;
	} catch (error: any) {
		if (error?.statusCode === 401) {
			throw new AuthenticationError("Sessionen har gått ut. Vänligen logga in igen.");
		}
		throw error;
	}
};

export function useGoalApi() {
	const fetchGoalData = async (
		goalId: number,
		forceRefresh = false,
	): Promise<GoalData> => {
		const query = forceRefresh ? `?_=${Date.now()}` : '';
		return await apiProxy<GoalData>(`goals/${goalId}${query}`);
	};

	const updateGoalTitle = async (goalId: number, title: string) => {
		await apiProxy(`goals/${goalId}`, {
			method: "PATCH",
			body: { title },
		});
	};

	const updateGoalIcon = async (goalId: number, icon: string) => {
		await apiProxy(`goals/${goalId}`, {
			method: "PATCH",
			body: { icon },
		});
	};

	// Sätter målets beskrivning (lång fritext). Tom sträng rensar den.
	const updateGoalDescription = async (
		goalId: number,
		description: string,
	) => {
		await apiProxy(`goals/${goalId}`, {
			method: "PATCH",
			body: { description },
		});
	};

	const updateGoalStatus = async (goalId: number, statusId: number) => {
		await apiProxy(`goals/${goalId}/status`, {
			method: "PATCH",
			body: { status_id: statusId },
		});
	};

	const toggleGoalStarted = async (goalId: number, started: string | null) => {
		await apiProxy(`goals/${goalId}`, {
			method: "PATCH",
			body: { started },
		});
	};

	const toggleGoalFinished = async (
		goalId: number,
		finished: string | null,
	) => {
		await apiProxy(`goals/${goalId}`, {
			method: "PATCH",
			body: { finished },
		});
	};

	// Markerar ett mål som evigt (eller upphäver evigt-flaggan).
	// Ett evigt mål kan inte slutföras - API:t avvisar det.
	const setGoalEternal = async (goalId: number, eternal: boolean) => {
		await apiProxy(`goals/${goalId}`, {
			method: "PATCH",
			body: { eternal },
		});
	};

	const deleteGoal = async (goalId: number) => {
		await apiProxy(`goals/${goalId}`, {
			method: "DELETE",
		});
	};

	const addParentRelation = async (childId: number, parentId: number) => {
		await apiProxy(`goals/relations`, {
			method: "POST",
			body: { childId, parentId },
		});
	};

	const removeParentRelation = async (childId: number, parentId: number) => {
		await apiProxy(`goals/relations`, {
			method: "DELETE",
			body: { childId, parentId },
		});
	};

	const addChildRelation = async (
		childId: number,
		parentId: number,
		order: number,
	) => {
		await apiProxy(`goals/relations`, {
			method: "POST",
			body: { childId, parentId, order },
		});
	};

	const updateGoalOrder = async (
		parentId: number,
		childId: number,
		order: number,
	) => {
		await apiProxy(`goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, order },
		});
	};

	// Sätter en ny ordning på alla undermål i ett svep (en request, en transaktion)
	const reorderChildren = async (
		parentId: number,
		childIds: number[],
	) => {
		await apiProxy(`goals/relations/reorder`, {
			method: "PATCH",
			body: { parentId, childIds },
		});
	};

	// Sätter en ny ordning på grundmålen (mål utan förälder) i ett svep
	const reorderGoals = async (goalIds: number[]) => {
		await apiProxy(`goals/reorder`, {
			method: "PATCH",
			body: { goalIds },
		});
	};

	const updateGoalWeight = async (
		parentId: number,
		childId: number,
		weight: number,
	) => {
		await apiProxy(`goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, weight },
		});
	};

	const setGoalWeight = async (
		childId: number,
		parentId: number,
		weight: number,
	) => {
		await apiProxy(`goals/relations`, {
			method: "POST",
			body: { childId, parentId, weight },
		});
	};

	const loadAllGoals = async (): Promise<Goal[]> => {
		return await apiProxy<Goal[]>(`goals`);
	};

	const findIconForGoalTitle = async (title: string): Promise<string | undefined> => {
		const query = title.trim();
		if (query.length < 2) return undefined;

		try {
			const response = await $fetch<{ results?: Array<{ full?: string }> }>("/api/icons", {
				query: { q: query, translate: "true" },
			});
			return response.results?.[0]?.full || undefined;
		} catch (error) {
			// Ikonvalet får inte hindra att målet skapas om översättning eller Iconify är otillgängligt.
			console.warn("Kunde inte hitta en ikon för målet:", error);
			return undefined;
		}
	};

	const createGoal = async (
		title: string,
		statusId?: number,
		icon?: string,
	): Promise<Goal> => {
		const body: { title: string; status_id?: number; icon?: string } = { title };
		if (statusId) {
			body.status_id = statusId;
		}
		body.icon = icon ?? await findIconForGoalTitle(title);

		return await apiProxy<Goal>(`goals`, {
			method: "POST",
			body,
		});
	};

	const addDependency = async (
		goalId: number,
		dependsOnId: number,
	): Promise<GoalDependency> => {
		return await apiProxy<GoalDependency>(`goals/dependencies`, {
			method: "POST",
			body: { goalId, dependsOnId },
		});
	};

	const removeDependency = async (goalId: number, dependsOnId: number) => {
		await apiProxy(`goals/dependencies`, {
			method: "DELETE",
			body: { goalId, dependsOnId },
		});
	};

	return {
		fetchGoalData,
		updateGoalTitle,
		updateGoalIcon,
		updateGoalDescription,
		updateGoalStatus,
		toggleGoalStarted,
		toggleGoalFinished,
		setGoalEternal,
		deleteGoal,
		addParentRelation,
		removeParentRelation,
		addChildRelation,
		updateGoalOrder,
		reorderChildren,
		reorderGoals,
		updateGoalWeight,
		setGoalWeight,
		createGoal,
		loadAllGoals,
		addDependency,
		removeDependency,
	};
}
