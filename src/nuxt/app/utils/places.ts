import type { Place } from "~/types/place";

/**
 * Ett mål eller en kalenderhändelse som kan ha platser.
 */
export interface PlaceOwner {
	places?: Place[];
}

/**
 * Ett mål är görbart på en plats om någon av målets platser är den
 * valda platsen.
 *
 * Flera platser på ett mål har OR-semantik - det räcker att vara på
 * någon av dem. Ett mål utan platser är aldrig görbart på en plats.
 */
export function isDoableAtPlace(
	goal: PlaceOwner,
	placeId: number | null,
): boolean {
	if (placeId === null) return false;
	return (goal.places ?? []).some((place) => place.id === placeId);
}

/**
 * Samma som isDoableAtPlace, men ett mål utan någon plats räknas också
 * som görbart.
 *
 * Ett mål utan plats är inte bundet till någon särskild plats och kan
 * därför göras var som helst - det ska inte försvinna när användaren
 * filtrerar listan på sin nuvarande plats.
 *
 * Om ingen plats är vald ("Utan plats") visas bara mål utan plats. De
 * är de enda som är "görbara" när man inte har bestämt var man
 * befinner sig - ett mål med plats hör till en specifik plats och syns
 * när den platsen är vald.
 */
export function isDoableAtPlaceOrUnplaced(
	goal: PlaceOwner,
	placeId: number | null,
): boolean {
	const places = goal.places ?? [];
	if (placeId === null) return places.length === 0;
	if (places.length === 0) return true;
	return places.some((place) => place.id === placeId);
}

/**
 * Ett mål är görbart om man är på någon av de angivna platserna.
 */
export function isDoableAtAnyPlace(
	goal: PlaceOwner,
	placeIds: number[],
): boolean {
	return placeIds.some((placeId) => isDoableAtPlace(goal, placeId));
}

/**
 * Platsnamnen på ett mål, som en kommaseparerad sträng.
 *
 * Används för att visa vilka platser ett mål gäller - t.ex. i tooltip.
 */
export function placeNames(goal: PlaceOwner): string {
	return (goal.places ?? []).map((place) => place.name).join(", ");
}
