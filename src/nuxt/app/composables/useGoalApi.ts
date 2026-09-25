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

// Custom session data interface
interface SessionData {
	userId: string;
	email?: string;
	name?: string;
	sessionToken: string;
	loggedInAt: number;
}

// Get session from our custom cookie
async function getSession(): Promise<SessionData | null> {
	try {
		const response = await $fetch<SessionData>('/api/auth/session');
		return response;
	} catch {
		return null;
	}
}

export function useGoalApi() {
	const config = useRuntimeConfig();
	const goApiUrl = config.public.goApiUrl || "http://localhost:8080";
	
	let cachedSession: SessionData | null = null;
	
	const getAuthHeaders = async () => {
		// Get session from our custom endpoint
		if (!cachedSession) {
			cachedSession = await getSession();
		}
		
		if (!cachedSession?.sessionToken) {
			throw new AuthenticationError("Du måste vara inloggad för att utföra denna åtgärd");
		}
		
		return {
			Authorization: `Bearer ${cachedSession.sessionToken}`,
		};
	};

	const fetchGoalData = async (
		goalId: number,
		forceRefresh = false,
	): Promise<GoalData> => {
		const url = forceRefresh
			? `${goApiUrl}/goals/${goalId}?_=${Date.now()}`
			: `${goApiUrl}/goals/${goalId}`;
		return await $fetch<GoalData>(url, { headers: await getAuthHeaders() });
	};

	const updateGoalTitle = async (goalId: number, title: string) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { title },
			headers: await getAuthHeaders(),
		});
	};

	const updateGoalIcon = async (goalId: number, icon: string) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { icon },
			headers: await getAuthHeaders(),
		});
	};

	const updateGoalStatus = async (goalId: number, statusId: number) => {
		await $fetch(`${goApiUrl}/goals/${goalId}/status`, {
			method: "PATCH",
			body: { status_id: statusId },
			headers: await getAuthHeaders(),
		});
	};

	const toggleGoalStarted = async (goalId: number, started: string | null) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { started },
			headers: await getAuthHeaders(),
		});
	};

	const toggleGoalFinished = async (
		goalId: number,
		finished: string | null,
	) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { finished },
			headers: await getAuthHeaders(),
		});
	};

	const deleteGoal = async (goalId: number) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "DELETE",
			headers: await getAuthHeaders(),
		});
	};

	const addParentRelation = async (childId: number, parentId: number) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "POST",
			body: { childId, parentId },
			headers: await getAuthHeaders(),
		});
	};

	const removeParentRelation = async (childId: number, parentId: number) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "DELETE",
			body: { childId, parentId },
			headers: await getAuthHeaders(),
		});
	};

	const addChildRelation = async (
		childId: number,
		parentId: number,
		order: number,
	) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "POST",
			body: { childId, parentId, order },
			headers: await getAuthHeaders(),
		});
	};

	const updateGoalOrder = async (
		parentId: number,
		childId: number,
		order: number,
	) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, order },
			headers: await getAuthHeaders(),
		});
	};

	const updateGoalWeight = async (
		parentId: number,
		childId: number,
		weight: number,
	) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, weight },
			headers: await getAuthHeaders(),
		});
	};

	const setGoalWeight = async (
		childId: number,
		parentId: number,
		weight: number,
	) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "POST",
			body: { childId, parentId, weight },
			headers: await getAuthHeaders(),
		});
	};

	const loadAllGoals = async (): Promise<Goal[]> => {
		return await $fetch<Goal[]>(`${goApiUrl}/goals`, {
			headers: await getAuthHeaders(),
		});
	};

	const createGoal = async (
		title: string,
		statusId?: number,
	): Promise<Goal> => {
		const body: { title: string; status_id?: number } = { title };
		if (statusId) {
			body.status_id = statusId;
		}
		return await $fetch<Goal>(`${goApiUrl}/goals`, {
			method: "POST",
			body,
			headers: await getAuthHeaders(),
		});
	};

	const addDependency = async (
		goalId: number,
		dependsOnId: number,
	): Promise<GoalDependency> => {
		return await $fetch<GoalDependency>(`${goApiUrl}/goals/dependencies`, {
			method: "POST",
			body: { goalId, dependsOnId },
			headers: await getAuthHeaders(),
		});
	};

	const removeDependency = async (goalId: number, dependsOnId: number) => {
		await $fetch(`${goApiUrl}/goals/dependencies`, {
			method: "DELETE",
			body: { goalId, dependsOnId },
			headers: await getAuthHeaders(),
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
