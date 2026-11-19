/**
 * En plats är ett område där uppgifter kan slutföras.
 *
 * Platsen är ett område, inte en punkt: latitud/longitude är
 * mittpunkten och radius_meters är cirkelns radie. Radien används för
 * att avgöra om en GPS-position ligger inom platsen.
 */
export interface Place {
	id: number;
	name: string;
	latitude: number;
	longitude: number;
	radius_meters: number;
	icon?: string;
	created_at: string;
	updated_at: string;
}

/**
 * Ett mål eller en kalenderhändelse med sina platser.
 *
 * Flera platser har OR-semantik: uppgiften är görbar om man är på
 * någon av platserna.
 */
export interface PlaceOwner {
	places: Place[];
}

/** Svar från /me/current-place. */
export interface CurrentPlaceResponse {
	place: Place | null;
}
