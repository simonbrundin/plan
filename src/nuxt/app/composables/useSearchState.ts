// Delad state - sökpaletten kan öppnas från flera ställen och alla
// konsumenter ska se samma värde.
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
