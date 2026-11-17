import { ref, watch } from "vue";

export interface IconResult {
	name: string;
	collection: string;
	full: string;
}

/**
 * Sökning i Iconify-katalogen.
 *
 * Sökningen är ett rent UI-behov och har ingen affärslogik, så den ligger
 * i Nuxt framför att läggas till i Go API:t. Både målens ikonväljare och
 * kalenderns ikonväljare använder den här sökningen.
 */
export function useIconSearch() {
	const searchQuery = ref("");
	const searchResults = ref<IconResult[]>([]);
	const isSearching = ref(false);
	const searchError = ref("");
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	watch(searchQuery, (newQuery) => {
		if (searchTimeout) {
			clearTimeout(searchTimeout);
		}

		// Sök inte med för få tecken - Iconify ger inga relevanta träffar då.
		if (!newQuery || newQuery.trim().length < 2) {
			searchResults.value = [];
			searchError.value = "";
			isSearching.value = false;
			return;
		}

		isSearching.value = true;
		searchTimeout = setTimeout(async () => {
			await performSearch(newQuery);
		}, 300);
	});

	async function performSearch(query: string) {
		searchError.value = "";
		try {
			const response = await fetch(`/api/icons?q=${encodeURIComponent(query)}`);
			if (!response.ok) {
				throw new Error(`Icon search failed with status ${response.status}`);
			}
			const data = await response.json();
			searchResults.value = data.results || [];
		} catch (error) {
			console.error("Failed to search icons:", error);
			searchResults.value = [];
			searchError.value = "Kunde inte söka ikoner just nu. Försök igen om en stund.";
		} finally {
			isSearching.value = false;
		}
	}

	/** Tömmer sökningen och eventuella resultat. */
	function reset() {
		if (searchTimeout) {
			clearTimeout(searchTimeout);
			searchTimeout = null;
		}
		searchQuery.value = "";
		searchResults.value = [];
		searchError.value = "";
		isSearching.value = false;
	}

	return { searchQuery, searchResults, isSearching, searchError, performSearch, reset };
}
