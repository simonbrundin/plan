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

	const createGoal = async (
		title: string,
		statusId?: number,
	): Promise<Goal> => {
		const body: { title: string; status_id?: number } = { title };
		if (statusId) {
			body.status_id = statusId;
		}
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
		updateGoalStatus,
		toggleGoalStarted,
		toggleGoalFinished,
		deleteGoal,
		addParentRelation,
		removeParentRelation,
		addChildRelation,
		updateGoalOrder,
		updateGoalWeight,
		setGoalWeight,
		createGoal,
		loadAllGoals,
		addDependency,
		removeDependency,
	};
}
