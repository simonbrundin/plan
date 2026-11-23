const LAST_PAGE_KEY = "plan-priority-toggle-last-page";

const lastVisitedPage = ref<string | null>(null);
const nowPage = "/now";

function getStoredLastPage(): string | null {
	if (typeof window === "undefined") return null;
	return localStorage.getItem(LAST_PAGE_KEY);
}

function setStoredLastPage(page: string) {
	if (typeof window === "undefined") return;
	localStorage.setItem(LAST_PAGE_KEY, page);
}

export function usePriorityToggle() {
	const route = useRoute();
	const router = useRouter();

	// Track route changes to remember the last non-priority page
	watch(
		() => route.path,
		(newPath) => {
			if (newPath !== nowPage) {
				lastVisitedPage.value = newPath;
				setStoredLastPage(newPath);
			}
		},
		{ immediate: true },
	);

	// Initialize from storage on first load
	onMounted(() => {
		const stored = getStoredLastPage();
		if (stored) {
			lastVisitedPage.value = stored;
		}
	});

	function togglePriority() {
		const currentPath = route.path;
		const isOnNowPage = currentPath === nowPage;

		if (isOnNowPage) {
			// Go back to last visited page
			const target = lastVisitedPage.value || "/";
			router.push(target);
		} else {
			// Go to the actionable goals page
			router.push(nowPage);
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		const key = event.key.toLowerCase();

		// Don't trigger if user is typing in an input
		const target = event.target as HTMLElement;
		const isInInput =
			target.tagName === "INPUT" ||
			target.tagName === "TEXTAREA" ||
			target.isContentEditable;

		if (isInInput) return;

		// Check if search is open
		const { isSearchOpen } = useSearchState();
		if (isSearchOpen.value) return;

		// På målsidan är 'p' platsväljarens tangent (platsen där målet
		// kan slutföras). Gör nu nås då via tangentmenyn.
		if (route.path.startsWith("/goal/")) return;

		// Toggle on 'p' key press (not hold). Shift+P används för
		// ordningsläget på Gör nu och ska inte växla sida.
		if (
			key === "p" &&
			!event.shiftKey &&
			!event.ctrlKey &&
			!event.metaKey &&
			!event.altKey
		) {
			event.preventDefault();
			togglePriority();
		}
	}

	return {
		lastVisitedPage: readonly(lastVisitedPage),
		togglePriority,
		handleKeydown,
	};
}
