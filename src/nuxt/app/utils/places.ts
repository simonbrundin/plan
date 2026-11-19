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
