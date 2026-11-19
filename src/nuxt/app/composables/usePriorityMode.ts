import type { Goal } from "~/types/goal";
import { matchesStartedFilter, type StartedFilter } from "~/utils/goalFilters";
import { isDoableAtPlace } from "~/utils/places";
import { moveChildId } from "~/utils/childOrder";

interface PrioritizedGoal extends Goal {
	weight: number;
	parentTitle: string | null;
	parentId: number | null;
}

const ROOT_GOAL_ID = 1;

function isSearchFieldFocused(isSearchOpen: Readonly<Ref<boolean>>): boolean {
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
// Default: visa endast mål som är påbörjade.
const DEFAULT_STARTED_FILTER: StartedFilter = "started";

const startedFilter = ref<StartedFilter>(DEFAULT_STARTED_FILTER);

// Filter "görbar här": visa bara mål som har den plats användaren
// befinner sig på. Flera platser på ett mål har OR-semantik - det
// räcker att vara på någon av dem.
const filterByCurrentPlace = useState<boolean>(
	"priority-filter-current-place",
	() => false,
);

let lastUpdateTime = 0;
const UPDATE_DEBOUNCE_MS = 150;

function matchesStartedFilterForGoals(goal: PrioritizedGoal): boolean {
	return matchesStartedFilter(goal, startedFilter.value);
}

export function usePriorityMode() {
	const { isSearchOpen } = useSearchState();
	const { setGoalWeight, reorderPriorityGoals } = useGoalApi();
	const { currentPlaceId } = usePlaces();

	function matchesCurrentPlace(goal: PrioritizedGoal): boolean {
		if (!filterByCurrentPlace.value) return true;
		return isDoableAtPlace(goal, currentPlaceId.value);
	}

	const visibleGoals = computed(() =>
		prioritizedGoals.value
			.filter(matchesStartedFilterForGoals)
			.filter(matchesCurrentPlace),
	);

	async function loadPrioritizedGoals() {
		if (isLoading.value) return;
		isLoading.value = true;

		try {
			// Backend returnerar redan filtrerade och sorterade goals med
			// weight från goal_relations, så vi behöver inte mappa om eller
			// sortera client-side.
			// Via our own server route so the session cookie is attached.
			const goals = await $fetch<PrioritizedGoal[]>('/api/proxy/goals/prioritized');

			prioritizedGoals.value = goals;

			if (selectedGoalId.value !== null) {
				const stillExists = goals.some((g) => g.id === selectedGoalId.value);
				if (!stillExists) {
					selectedGoalId.value = visibleGoals.value[0]?.id ?? null;
				}
			} else if (visibleGoals.value.length > 0) {
				selectedGoalId.value = visibleGoals.value[0]?.id ?? null;
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
		// Spara till backend. Om målet inte har någon parent-relation ännu
		// (top-level utan explicit parent) skapar vi en till root (ID 1) via
		// UPSERT i relations.post. Annars uppdaterar vi befintlig relation.
		const parentId = goal.parentId ?? ROOT_GOAL_ID;
		try {
			await setGoalWeight(goalId, parentId, newWeight);
		} catch (err) {
			console.error("Failed to persist weight change:", err);
		}
	}

	let prioritySaveChain: Promise<void> = Promise.resolve();

	function moveGoal(direction: "up" | "down") {
		const movingId = selectedGoalId.value;
		if (movingId === null) return;

		const reordered = moveChildId(
			prioritizedGoals.value.map((goal) => goal.id),
			visibleGoals.value.map((goal) => goal.id),
			movingId,
			direction,
		);
		if (!reordered) return;

		const byId = new Map(prioritizedGoals.value.map((goal) => [goal.id, goal]));
		reordered.forEach((id, index) => {
			const goal = byId.get(id);
			if (goal) goal.priorityOrder = index;
		});
		prioritizedGoals.value = reordered
			.map((id) => byId.get(id))
			.filter((goal): goal is PrioritizedGoal => goal !== undefined);
		selectedGoalId.value = movingId;

		const ids = reordered;
		prioritySaveChain = prioritySaveChain
			.then(() => reorderPriorityGoals(ids))
			.catch((err) => console.error("Failed to save priority order:", err));
	}

	function handleKeydown(event: KeyboardEvent) {
		const key = event.key.toLowerCase();

		// Prioritetslägets tangenthantering ska ENDAST vara aktiv på
		// prioritetssidan. På andra sidor (t.ex. goal-sidan) har sidan
		// sin egen vim-hantering och den globala hanteraren ska inte
		// fånga l/j/k och navigera till prioriterade mål.
		const route = useRoute();
		if (route.path !== "/priority") {
			return;
		}

		heldKeys.value.add(key);
		isPriorityMode.value = heldKeys.value.has("p");

		const target = event.target as HTMLElement;
		if (
			target.tagName === "INPUT" ||
			target.tagName === "TEXTAREA" ||
			target.isContentEditable ||
			isSearchFieldFocused(isSearchOpen)
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
			if (key === "arrowup" || key === "arrowdown") {
				event.preventDefault();
				moveGoal(key === "arrowup" ? "up" : "down");
			} else if (key === "l" && selectedGoalId.value !== null) {
				event.preventDefault();
				navigateTo(`/goal/${selectedGoalId.value}`);
			} else if (key === "j") {
				event.preventDefault();
				const list = visibleGoals.value;
				const currentIndex = list.findIndex(
					(g) => g.id === selectedGoalId.value,
				);
				if (currentIndex < list.length - 1) {
					const nextGoal = list[currentIndex + 1];
					if (nextGoal) {
						selectedGoalId.value = nextGoal.id;
					}
				}
			} else if (key === "k") {
				event.preventDefault();
				const list = visibleGoals.value;
				const currentIndex = list.findIndex(
					(g) => g.id === selectedGoalId.value,
				);
				if (currentIndex > 0) {
					const prevGoal = list[currentIndex - 1];
					if (prevGoal) {
						selectedGoalId.value = prevGoal.id;
					}
				}
			}
		}
	}

	function handleKeyup(event: KeyboardEvent) {
		const route = useRoute();
		if (route.path !== "/priority") {
			return;
		}

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
		return visibleGoals.value.findIndex((g) => g.id === selectedGoalId.value);
	}

	function setStartedFilter(filter: "all" | "started" | "not_started") {
		startedFilter.value = filter;
		const list = visibleGoals.value;
		if (
			selectedGoalId.value === null ||
			!list.some((g) => g.id === selectedGoalId.value)
		) {
			selectedGoalId.value = list[0]?.id ?? null;
		}
	}

	// Slår på och av filtret "görbar här".
	function setFilterByCurrentPlace(enabled: boolean) {
		filterByCurrentPlace.value = enabled;
		const list = visibleGoals.value;
		if (
			selectedGoalId.value === null ||
			!list.some((g) => g.id === selectedGoalId.value)
		) {
			selectedGoalId.value = list[0]?.id ?? null;
		}
	}

	return {
		isPriorityMode: readonly(isPriorityMode),
		selectedGoalId: readonly(selectedGoalId),
		prioritizedGoals: readonly(prioritizedGoals),
		visibleGoals,
		isLoading: readonly(isLoading),
		startedFilter: readonly(startedFilter),
		filterByCurrentPlace: readonly(filterByCurrentPlace),
		loadPrioritizedGoals,
		updateWeight,
		handleKeydown,
		handleKeyup,
		selectGoal,
		getSelectedIndex,
		setStartedFilter,
		setFilterByCurrentPlace,
	};
}
