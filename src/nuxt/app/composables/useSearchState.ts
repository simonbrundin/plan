const isSearchOpen = ref(false);

export function useSearchState() {
	function openSearch() {
		isSearchOpen.value = true;
	}

	function closeSearch() {
		isSearchOpen.value = false;
	}

	return {
		isSearchOpen: readonly(isSearchOpen),
		openSearch,
		closeSearch,
	};
}
