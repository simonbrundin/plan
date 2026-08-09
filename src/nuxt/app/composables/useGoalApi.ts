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

export function useGoalApi() {
	const config = useRuntimeConfig();
	const { user } = useUserSession();
	const goApiUrl = config.public.goApiUrl || "http://localhost:8080";

	const authHeaders = () => ({
		Authorization: `Bearer user_${(user.value as any)?.id}`,
	});

	const fetchGoalData = async (
		goalId: number,
		forceRefresh = false,
	): Promise<GoalData> => {
		const url = forceRefresh
			? `${goApiUrl}/goals/${goalId}?_=${Date.now()}`
			: `${goApiUrl}/goals/${goalId}`;
		return await $fetch<GoalData>(url, { headers: authHeaders() });
	};

	const updateGoalTitle = async (goalId: number, title: string) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { title },
			headers: authHeaders(),
		});
	};

	const updateGoalIcon = async (goalId: number, icon: string) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { icon },
			headers: authHeaders(),
		});
	};

	const updateGoalStatus = async (goalId: number, statusId: number) => {
		await $fetch(`${goApiUrl}/goals/${goalId}/status`, {
			method: "PATCH",
			body: { status_id: statusId },
			headers: authHeaders(),
		});
	};

	const toggleGoalStarted = async (goalId: number, started: string | null) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { started },
			headers: authHeaders(),
		});
	};

	const toggleGoalFinished = async (
		goalId: number,
		finished: string | null,
	) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { finished },
			headers: authHeaders(),
		});
	};

	const deleteGoal = async (goalId: number) => {
		await $fetch(`${goApiUrl}/goals/${goalId}`, {
			method: "DELETE",
			headers: authHeaders(),
		});
	};

	const addParentRelation = async (childId: number, parentId: number) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "POST",
			body: { childId, parentId },
			headers: authHeaders(),
		});
	};

	const removeParentRelation = async (childId: number, parentId: number) => {
		await $fetch(`${goApiUrl}/goals/relations`, {
			method: "DELETE",
			body: { childId, parentId },
			headers: authHeaders(),
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
			headers: authHeaders(),
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
			headers: authHeaders(),
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
			headers: authHeaders(),
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
			headers: authHeaders(),
		});
	};

	const loadAllGoals = async (): Promise<Goal[]> => {
		return await $fetch<Goal[]>(`${goApiUrl}/goals`, {
			headers: authHeaders(),
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
			headers: authHeaders(),
		});
	};

	const addDependency = async (
		goalId: number,
		dependsOnId: number,
	): Promise<GoalDependency> => {
		return await $fetch<GoalDependency>(
			`${goApiUrl}/goals/dependencies`,
			{
				method: "POST",
				body: { goalId, dependsOnId },
				headers: authHeaders(),
			},
		);
	};

	const removeDependency = async (goalId: number, dependsOnId: number) => {
		await $fetch(`${goApiUrl}/goals/dependencies`, {
			method: "DELETE",
			body: { goalId, dependsOnId },
			headers: authHeaders(),
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
