import type { Goal } from "~/types/goal";

interface GoalWithWeight extends Goal {
	weight: number;
	order: number;
}

interface GoalData {
	goal: Goal;
	children: GoalWithWeight[];
	parents: Goal[];
	allGoals: Goal[];
}

export function useGoalApi() {
	const fetchGoalData = async (goalId: number): Promise<GoalData> => {
		return await $fetch<GoalData>(`/api/goals/${goalId}`);
	};

	const updateGoalTitle = async (goalId: number, title: string) => {
		await $fetch(`/api/goals/${goalId}`, {
			method: "PATCH",
			body: { title },
		});
	};

	const updateGoalIcon = async (goalId: number, icon: string) => {
		await $fetch(`/api/goals/${goalId}`, {
			method: "PATCH",
			body: { icon },
		});
	};

	const toggleGoalStarted = async (goalId: number, started: string | null) => {
		await $fetch(`/api/goals/${goalId}`, {
			method: "PATCH",
			body: { started },
		});
	};

	const toggleGoalFinished = async (
		goalId: number,
		finished: string | null,
	) => {
		await $fetch(`/api/goals/${goalId}`, {
			method: "PATCH",
			body: { finished },
		});
	};

	const deleteGoal = async (goalId: number) => {
		await $fetch(`/api/goals/${goalId}`, {
			method: "DELETE",
		});
	};

	const addParentRelation = async (childId: number, parentId: number) => {
		await $fetch("/api/goals/relations", {
			method: "POST",
			body: { childId, parentId },
		});
	};

	const removeParentRelation = async (childId: number, parentId: number) => {
		await $fetch("/api/goals/relations", {
			method: "DELETE",
			body: JSON.stringify({ childId, parentId }),
		});
	};

	const addChildRelation = async (
		childId: number,
		parentId: number,
		order: number,
	) => {
		await $fetch("/api/goals/relations", {
			method: "POST",
			body: { childId, parentId, order },
		});
	};

	const updateGoalOrder = async (
		parentId: number,
		childId: number,
		order: number,
	) => {
		await $fetch("/api/goals/relations", {
			method: "PATCH",
			body: { childId, parentId, order },
		});
	};

	const updateGoalWeight = async (
		parentId: number,
		childId: number,
		weight: number,
	) => {
		await $fetch("/api/goals/relations", {
			method: "PATCH",
			body: { childId, parentId, weight },
		});
	};

	// Sätt vikt via UPSERT — skapar relationen om den inte finns, uppdaterar
	// om den finns. Användbart när prioriteringssidan vill sätta vikt på ett
	// top-level-mål som kanske inte har någon parent-relation ännu.
	const setGoalWeight = async (
		childId: number,
		parentId: number,
		weight: number,
	) => {
		await $fetch("/api/goals/relations", {
			method: "POST",
			body: { childId, parentId, weight },
		});
	};

	const loadAllGoals = async (): Promise<Goal[]> => {
		return await $fetch<Goal[]>("/api/goals");
	};

	const createGoal = async (title: string): Promise<Goal> => {
		return await $fetch<Goal>("/api/goals", {
			method: "POST",
			body: { title },
		});
	};

	return {
		fetchGoalData,
		updateGoalTitle,
		updateGoalIcon,
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
	};
}
