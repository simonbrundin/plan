interface NavigationItem {
	id: string | number;
	title: string;
	finished: boolean;
	weight?: number;
}

type NavigationMode = "goal" | "parent" | "child";
type PageType = "goal" | "inbox" | "priority" | "other";

interface NavigationState {
	selectedIndex: number;
	mode: NavigationMode;
	items: NavigationItem[];
	goalMode: "main" | "child" | "parent";
}

interface GoalPageCallbacks {
	onSelectChild?: (index: number) => void;
	onSelectParent?: (index: number) => void;
	onSelectGoal?: () => void;
	onOpenChild?: (index: number) => void;
	onOpenParent?: (index: number) => void;
	onMoveChildUp?: () => void;
	onMoveChildDown?: () => void;
	onToggleFinished?: () => void;
	onDelete?: () => void;
	onEdit?: (atBeginning: boolean) => void;
	onToggleParentSearch?: () => void;
	onToggleChildSearch?: () => void;
}

const currentPage = ref<PageType>("other");
const currentGoalId = ref<number | null>(null);
const navigationState = ref<NavigationState>({
	selectedIndex: -1,
	mode: "child",
	items: [],
	goalMode: "main",
});

const isEnabled = ref(false);

let goalPageCallbacks: GoalPageCallbacks | null = null;

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

export function useGlobalNavigation() {
	const { isSearchOpen } = useSearchState();

	function enableNavigation() {
		isEnabled.value = true;
	}

	function disableNavigation() {
		isEnabled.value = false;
	}

	function setPage(page: PageType) {
		currentPage.value = page;
		if (page !== "goal") {
			navigationState.value = {
				selectedIndex: -1,
				mode: "child",
				items: [],
				goalMode: "main",
			};
		}
	}

	function setCurrentGoal(goalId: number) {
		currentGoalId.value = goalId;
		navigationState.value = {
			selectedIndex: -1,
			mode: "child",
			items: [],
			goalMode: "main",
		};
	}

	function setNavigationItems(
		items: NavigationItem[],
		mode: NavigationMode = "child",
	) {
		navigationState.value = {
			...navigationState.value,
			items,
			mode,
		};
	}

	function setSelectedIndex(index: number) {
		navigationState.value.selectedIndex = index;
	}

	function setMode(mode: NavigationMode) {
		navigationState.value.mode = mode;
	}

	function registerGoalPageCallbacks(callbacks: GoalPageCallbacks) {
		goalPageCallbacks = callbacks;
	}

	function unregisterGoalPageCallbacks() {
		goalPageCallbacks = null;
	}

	function isInputFocused(): boolean {
		const activeElement = document.activeElement;
		if (!activeElement) return false;
		return (
			activeElement.tagName === "INPUT" ||
			activeElement.tagName === "TEXTAREA" ||
			(activeElement as HTMLElement).isContentEditable
		);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!isEnabled.value) return;

		if (isSearchFieldFocused(isSearchOpen)) {
			return;
		}

		const key = event.key.toLowerCase();

		// Mellanslag ska fungera normalt i textarea/input.
		if (isInputFocused()) return;

		if (isSearchFieldFocused(isSearchOpen)) return;

		if (currentPage.value !== "goal") return;
		if (!goalPageCallbacks) return;

		const state = navigationState.value;

		if (key === "d") {
			event.preventDefault();
			goalPageCallbacks.onToggleFinished?.();
			return;
		}

		if (key === "x") {
			event.preventDefault();
			goalPageCallbacks.onDelete?.();
			return;
		}

		if (key === "i" || key === "a") {
			event.preventDefault();
			goalPageCallbacks.onEdit?.(key === "i");
			return;
		}

		if (key === "enter") {
			event.preventDefault();
			if (state.mode === "parent") {
				goalPageCallbacks.onToggleParentSearch?.();
			} else {
				goalPageCallbacks.onToggleChildSearch?.();
			}
			return;
		}

		if (key === "arrowup") {
			event.preventDefault();
			goalPageCallbacks.onMoveChildUp?.();
			return;
		}

		if (key === "arrowdown") {
			event.preventDefault();
			goalPageCallbacks.onMoveChildDown?.();
			return;
		}

		if (key === "j") {
			event.preventDefault();
			if (state.mode === "parent") {
				const newIndex = Math.min(
					state.selectedIndex + 1,
					state.items.length - 1,
				);
				goalPageCallbacks.onSelectParent?.(newIndex);
			} else if (state.mode === "goal") {
				if (state.items.length > 0) {
					goalPageCallbacks.onSelectChild?.(0);
				}
			} else {
				if (state.selectedIndex === -1) {
					goalPageCallbacks.onSelectChild?.(0);
				} else {
					const newIndex = Math.min(
						state.selectedIndex + 1,
						state.items.length - 1,
					);
					goalPageCallbacks.onSelectChild?.(newIndex);
				}
			}
			return;
		}

		if (key === "k") {
			event.preventDefault();
			if (state.mode === "parent") {
				const newIndex = Math.max(state.selectedIndex - 1, 0);
				goalPageCallbacks.onSelectParent?.(newIndex);
			} else if (state.mode === "goal") {
				if (state.items.length > 0 && currentGoalId.value !== 1) {
					goalPageCallbacks.onSelectParent?.(0);
				}
			} else {
				if (state.selectedIndex === 0) {
					goalPageCallbacks.onSelectGoal?.();
				} else {
					const newIndex = Math.max(state.selectedIndex - 1, 0);
					goalPageCallbacks.onSelectChild?.(newIndex);
				}
			}
			return;
		}

		if (key === "h") {
			event.preventDefault();
			if (state.mode === "parent") {
				goalPageCallbacks.onOpenParent?.(state.selectedIndex);
			} else if (state.mode === "goal" && currentGoalId.value !== 1) {
				goalPageCallbacks.onSelectParent?.(0);
			}
			return;
		}

		if (key === "l") {
			event.preventDefault();
			if (state.mode === "parent") {
				goalPageCallbacks.onOpenParent?.(state.selectedIndex);
			} else if (state.mode === "goal") {
				if (state.items.length > 0) {
					goalPageCallbacks.onSelectChild?.(0);
				}
			} else {
				if (state.selectedIndex >= 0) {
					goalPageCallbacks.onOpenChild?.(state.selectedIndex);
				}
			}
			return;
		}
	}

	return {
		currentPage: readonly(currentPage),
		currentGoalId: readonly(currentGoalId),
		navigationState: readonly(navigationState),
		isEnabled: readonly(isEnabled),
		enableNavigation,
		disableNavigation,
		setPage,
		setCurrentGoal,
		setNavigationItems,
		setSelectedIndex,
		setMode,
		registerGoalPageCallbacks,
		unregisterGoalPageCallbacks,
		handleKeydown,
	};
}
