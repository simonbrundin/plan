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

// Den ordning som ska sparas och om ett sparande pågår. Detta ligger på
// modulnivå så att alla som anropar composablen delar samma kö - annars
// kan en instans spara en äldre ordning efter att en annan sparat en
// nyare, och det är den sista sparade ordningen som gäller.
let pendingOrder: number[] | null = null;
let orderSaveInFlight = false;
let orderSaver: ((ids: number[]) => Promise<unknown>) | null = null;
let unloadListenerRegistered = false;

// Sidor kan laddas om direkt efter att ordningen ändrats. Då måste den
// sista ordningen hinna sparas, annars ser det ut som att omsorteringen
// aldrig sparades. keepalive låter förfrågan klart även om sidan stängs.
function flushOrderOnUnload() {
	if (pendingOrder === null || !orderSaver) return;

	const ids = pendingOrder;
	pendingOrder = null;

	void fetch("/api/proxy/goals/priority/reorder", {
		method: "PATCH",
		headers: { "content-type": "application/json" },
		body: JSON.stringify({ goalIds: ids }),
		keepalive: true,
	}).catch(() => {
		// Sidan stängs - det finns inget att rapportera.
	});
}

function matchesStartedFilterForGoals(goal: PrioritizedGoal): boolean {
	return matchesStartedFilter(goal, startedFilter.value);
}

export function usePriorityMode() {
	const { isSearchOpen } = useSearchState();
	const { setGoalWeight, reorderPriorityGoals } = useGoalApi();
	const { currentPlaceId } = usePlaces();
	// Spara-funktionen sätts på modulnivå så att sidans pagehide-handler
	// kan spara den sista ordningen även om sidan laddas om direkt.
	orderSaver = reorderPriorityGoals;

	if (import.meta.client && !unloadListenerRegistered) {
		unloadListenerRegistered = true;
		window.addEventListener("pagehide", flushOrderOnUnload);
	}


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

	// Sparar ordningen. Endast den SENASTE ordningen sparas - om flera
	// tangenttryck kommer i rad ska vi inte skicka en request per tryck,
	// utan den sista ordningen vinner. Annars hamnar en kö av requests
	// efter varandra, och den som sparar sist kan vara en gammal
	// ordning från flera tangenttryck tidigare.
	async function savePendingOrder() {
		if (orderSaveInFlight || pendingOrder === null || !orderSaver) return;

		const ids = pendingOrder;
		pendingOrder = null;
		orderSaveInFlight = true;
		try {
			await orderSaver(ids);
		} catch (err) {
			console.error("Failed to save priority order:", err);
			// Listan i gränssnittet är då inte längre i synk med servern.
			await loadPrioritizedGoals();
		} finally {
			orderSaveInFlight = false;
			// En ny ordning kan ha kommit in under spartiden.
			if (pendingOrder !== null) {
				void savePendingOrder();
			}
		}
	}

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

		pendingOrder = [...reordered];
		void savePendingOrder();
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

		// Viktläget aktiveras med Shift+P (håll ned). Ett vanligt "p"
		// växlar i stället mellan prioritetssidan och föregående sida,
		// så de två tangenterna får inte krocka.
		heldKeys.value.add(event.key);
		isPriorityMode.value = heldKeys.value.has("P");

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

		// Tangenter kan released med olika modifier-tillstånd, så både
		// "p" och "P" tas bort ur listan över nedtryckta tangenter.
		heldKeys.value.delete(event.key);
		heldKeys.value.delete(event.key.toLowerCase());
		heldKeys.value.delete(event.key.toUpperCase());
		isPriorityMode.value = heldKeys.value.has("P");
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
