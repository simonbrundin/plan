import type { CurrentPlaceResponse, Place } from "~/types/place";

/**
 * apiProxy skickar requesten till Go API:t via Nuxt:s proxy.
 *
 * Sessionen ligger i Go API:t, så proxyn vidarebefordrar cookien och
 * valideringen sker där.
 */
const apiProxy = async <T>(
	path: string,
	options: RequestInit = {},
): Promise<T> => {
	return await $fetch<T>(`/api/proxy/${path}`, {
		...options,
	});
};

export interface CreatePlaceInput {
	name: string;
	latitude: number;
	longitude: number;
	radius_meters?: number;
	icon?: string;
}

export interface UpdatePlaceInput {
	name?: string;
	latitude?: number;
	longitude?: number;
	radius_meters?: number;
	icon?: string;
}

/**
 * usePlaces ger åtkomst till användarens platser och kopplingar mellan
 * platser och mål/händelser.
 */
export function usePlaces() {
	/** Alla användarens platser. */
	const places = useState<Place[]>("places", () => []);

	/** Användarens valda "jag är här"-plats. */
	const currentPlaceId = useState<number | null>("current-place-id", () => null);

	/** Laddar platserna och den nuvarande platsen. */
	const fetchPlaces = async (): Promise<void> => {
		const result = await apiProxy<Place[]>("places");
		places.value = result ?? [];
	};

	/** Laddar den nuvarande platsen ("jag är här"). */
	const fetchCurrentPlace = async (): Promise<void> => {
		const result = await apiProxy<CurrentPlaceResponse>("me/current-place");
		currentPlaceId.value = result?.place?.id ?? null;
	};

	/** Skapar en ny plats. */
	const createPlace = async (input: CreatePlaceInput): Promise<Place> => {
		const place = await apiProxy<Place>("places", {
			method: "POST",
			body: input,
		});
		await fetchPlaces();
		return place;
	};

	/** Uppdaterar en plats. */
	const updatePlace = async (
		placeId: number,
		input: UpdatePlaceInput,
	): Promise<Place> => {
		const place = await apiProxy<Place>(`places/${placeId}`, {
			method: "PATCH",
			body: input,
		});
		await fetchPlaces();
		return place;
	};

	/** Tar bort en plats. Kopplingar till mål och händelser försvinner. */
	const deletePlace = async (placeId: number): Promise<void> => {
		await apiProxy(`places/${placeId}`, { method: "DELETE" });
		await fetchPlaces();
	};

	/** Sätter (eller rensar) användarens "jag är här"-plats. */
	const setCurrentPlace = async (placeId: number | null): Promise<void> => {
		await apiProxy("me/current-place", {
			method: "PUT",
			body: { place_id: placeId },
		});
		currentPlaceId.value = placeId;
	};

	/**
	 * Returnerar de platser vars radie innehåller positionen.
	 *
	 * Positionen skickas till API:t som räknar avståndet mot radien.
	 * Råpositionen sparas inte.
	 */
	const findNearbyPlaces = async (
		latitude: number,
		longitude: number,
	): Promise<Place[]> => {
		const result = await apiProxy<Place[]>(
			`places/nearby?lat=${latitude}&lon=${longitude}`,
		);
		return result ?? [];
	};

	/** Hämtar platserna kopplade till ett mål. */
	const fetchGoalPlaces = async (goalId: number): Promise<Place[]> => {
		const result = await apiProxy<Place[]>(`goals/${goalId}/places`);
		return result ?? [];
	};

	/** Ersätter platserna på ett mål. Tom lista tar bort alla platser. */
	const setGoalPlaces = async (
		goalId: number,
		placeIds: number[],
	): Promise<Place[]> => {
		return await apiProxy<Place[]>(`goals/${goalId}/places`, {
			method: "PUT",
			body: { place_ids: placeIds },
		});
	};

	/** Hämtar platserna kopplade till en kalenderhändelse. */
	const fetchEventPlaces = async (eventId: number): Promise<Place[]> => {
		const result = await apiProxy<Place[]>(`calendar-events/${eventId}/places`);
		return result ?? [];
	};

	/** Ersätter platserna på en kalenderhändelse. */
	const setEventPlaces = async (
		eventId: number,
		placeIds: number[],
	): Promise<Place[]> => {
		return await apiProxy<Place[]>(`calendar-events/${eventId}/places`, {
			method: "PUT",
			body: { place_ids: placeIds },
		});
	};

	return {
		places,
		currentPlaceId,
		fetchPlaces,
		fetchCurrentPlace,
		createPlace,
		updatePlace,
		deletePlace,
		setCurrentPlace,
		findNearbyPlaces,
		fetchGoalPlaces,
		setGoalPlaces,
		fetchEventPlaces,
		setEventPlaces,
	};
}
