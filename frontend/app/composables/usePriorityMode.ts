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
	async function loadPrioritizedGoals() {
		if (isLoading.value) return;
		isLoading.value = true;

		try {
			const goals = await $fetch<Goal[]>("/api/goals");

			const goalsWithWeight: PrioritizedGoal[] = goals
				.filter((g) => !g.finished)
				.map((g) => ({
					...g,
					weight: 10,
					parentTitle: null,
				}))
				.sort((a, b) => b.weight - a.weight);

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
		goal.weight = newWeight;
		prioritizedGoals.value = [...prioritizedGoals.value].sort(
			(a, b) => b.weight - a.weight,
		);
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
