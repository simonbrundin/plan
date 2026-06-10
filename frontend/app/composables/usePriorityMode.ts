import type { Goal } from "~/types/goal";

interface PrioritizedGoal extends Goal {
	weight: number;
	parentTitle: string | null;
}

const { isSearchOpen } = useSearchState();

function isSearchFieldFocused(): boolean {
	if (isSearchOpen.value) return true;
	const active = document.activeElement;
	if (!active) return false;
	const inModal =
		active.closest('[class*="bg-elevated"]') ||
		active.closest('[class*="fixed"]');
	if (inModal && active.tagName === "INPUT") return true;
	return false;
}
const heldKeys = ref<Set<string>>(new Set());
const isPriorityMode = ref(false);
const selectedGoalId = ref<number | null>(null);
const prioritizedGoals = ref<PrioritizedGoal[]>([]);
const isLoading = ref(false);

let lastUpdateTime = 0;
const UPDATE_DEBOUNCE_MS = 150;

export function usePriorityMode() {
	const config = useRuntimeConfig();

	async function loadPrioritizedGoals() {
		if (isLoading.value) return;
		isLoading.value = true;

		try {
			const response = await fetch(`${config.public.GQL_HOST}`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					"x-hasura-admin-secret": config.public.hasuraAdminSecret,
				},
				body: JSON.stringify({
					query: `
            query GetPrioritizedGoals {
              goals(limit: 1000, order_by: { created: desc }) {
                id
                title
                icon
                created
                finished
                inbox
              }
              goal_relations(limit: 1000) {
                parent_id
                child_id
                weight
              }
            }
          `,
				}),
			});

			const result = await response.json();

			if (result.errors) {
				throw new Error(result.errors[0].message);
			}

			const goals = result.data.goals;
			const relations = result.data.goal_relations;

			const goalsWithWeight: PrioritizedGoal[] = goals
				.filter((g: Goal) => !g.finished)
				.map((g: Goal) => {
					const relation = relations.find((r: any) => r.child_id === g.id);

					const weight = relation?.weight || 10;
					let parentTitle: string | null = null;

					if (relation) {
						const parentGoal = goals.find(
							(pg: Goal) => pg.id === relation.parent_id,
						);
						parentTitle = parentGoal?.title || null;
					}

					return {
						...g,
						weight,
						parentTitle,
					};
				})
				.sort((a: PrioritizedGoal, b: PrioritizedGoal) => b.weight - a.weight);

			prioritizedGoals.value = goalsWithWeight;

			if (selectedGoalId.value !== null) {
				const stillExists = goalsWithWeight.some(
					(g) => g.id === selectedGoalId.value,
				);
				if (!stillExists) {
					selectedGoalId.value = goalsWithWeight[0]?.id ?? null;
				}
			} else if (goalsWithWeight.length > 0) {
				selectedGoalId.value = goalsWithWeight[0]?.id ?? null;
			}
		} catch (err) {
			console.error("Failed to load prioritized goals:", err);
		} finally {
			isLoading.value = false;
		}
	}

	async function updateWeight(goalId: number, delta: number) {
		const now = Date.now();
		if (now - lastUpdateTime < UPDATE_DEBOUNCE_MS) {
			return;
		}
		lastUpdateTime = now;

		const goal = prioritizedGoals.value.find((g) => g.id === goalId);
		if (!goal) return;

		const newWeight = Math.max(1, Math.min(100, goal.weight + delta));

		try {
			await fetch(`${config.public.GQL_HOST}`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					"x-hasura-admin-secret": config.public.hasuraAdminSecret,
				},
				body: JSON.stringify({
					query: `
            mutation UpdateGoalWeight($parent_id: Int!, $child_id: Int!, $weight: Int!) {
              update_goal_relations(
                where: { child_id: { _eq: $child_id }, parent_id: { _eq: $parent_id } }
                _set: { weight: $weight }
              ) {
                affected_rows
              }
            }
          `,
					variables: {
						parent_id: goal.parentTitle
							? prioritizedGoals.value.find((g) => g.title === goal.parentTitle)
									?.id || 1
							: 1,
						child_id: goalId,
						weight: newWeight,
					},
				}),
			});

			goal.weight = newWeight;
			prioritizedGoals.value = [...prioritizedGoals.value].sort(
				(a, b) => b.weight - a.weight,
			);
		} catch (err) {
			console.error("Failed to update weight:", err);
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		const key = event.key.toLowerCase();

		heldKeys.value.add(key);
		isPriorityMode.value = heldKeys.value.has("p");

		const target = event.target as HTMLElement;
		if (
			target.tagName === "INPUT" ||
			target.tagName === "TEXTAREA" ||
			target.isContentEditable ||
			isSearchFieldFocused()
		) {
			return;
		}

		if (isPriorityMode.value) {
			if (key === "k" && selectedGoalId.value !== null) {
				event.preventDefault();
				updateWeight(selectedGoalId.value, 1);
			} else if (key === "j" && selectedGoalId.value !== null) {
				event.preventDefault();
				updateWeight(selectedGoalId.value, -1);
			} else if (key === "p") {
				isPriorityMode.value = false;
			}
		} else {
			if (key === "l" && selectedGoalId.value !== null) {
				event.preventDefault();
				navigateTo(`/goal/${selectedGoalId.value}`);
			} else if (key === "j") {
				event.preventDefault();
				const currentIndex = prioritizedGoals.value.findIndex(
					(g) => g.id === selectedGoalId.value,
				);
				if (currentIndex < prioritizedGoals.value.length - 1) {
					const nextGoal = prioritizedGoals.value[currentIndex + 1];
					if (nextGoal) {
						selectedGoalId.value = nextGoal.id;
					}
				}
			} else if (key === "k") {
				event.preventDefault();
				const currentIndex = prioritizedGoals.value.findIndex(
					(g) => g.id === selectedGoalId.value,
				);
				if (currentIndex > 0) {
					const prevGoal = prioritizedGoals.value[currentIndex - 1];
					if (prevGoal) {
						selectedGoalId.value = prevGoal.id;
					}
				}
			}
		}
	}

	function handleKeyup(event: KeyboardEvent) {
		const key = event.key.toLowerCase();
		heldKeys.value.delete(key);

		if (key === "p") {
			isPriorityMode.value = false;
		}
	}

	function selectGoal(goalId: number) {
		selectedGoalId.value = goalId;
	}

	function getSelectedIndex(): number {
		return prioritizedGoals.value.findIndex(
			(g) => g.id === selectedGoalId.value,
		);
	}

	return {
		isPriorityMode: readonly(isPriorityMode),
		selectedGoalId: readonly(selectedGoalId),
		prioritizedGoals: readonly(prioritizedGoals),
		isLoading: readonly(isLoading),
		loadPrioritizedGoals,
		updateWeight,
		handleKeydown,
		handleKeyup,
		selectGoal,
		getSelectedIndex,
	};
}
