import { describe, expect, it } from "bun:test";

import type { Place } from "../../app/types/place";
import { isDoableAtAnyPlace, isDoableAtPlace, isDoableAtPlaceOrUnplaced, placeNames } from "../../app/utils/places";

const place = (id: number, name: string): Place => ({
	id,
	name,
	latitude: 59.3293,
	longitude: 18.0686,
	radius_meters: 100,
	created_at: "2026-11-20T00:00:00Z",
	updated_at: "2026-11-20T00:00:00Z",
});

describe("isDoableAtPlace", () => {
	it("ett mål utan platser är inte görbart någonstans", () => {
		expect(isDoableAtPlace({}, 1)).toBe(false);
	});

	it("ett mål med platsen är görbart där", () => {
		expect(isDoableAtPlace({ places: [place(1, "Hem")] }, 1)).toBe(true);
	});

	it("ett mål är inte görbart på en annan plats", () => {
		expect(isDoableAtPlace({ places: [place(1, "Hem")] }, 2)).toBe(false);
	});

	it("flera platser räcker - det räcker att vara på en av dem", () => {
		const goal = { places: [place(1, "Hem"), place(2, "Kontor")] };
		expect(isDoableAtPlace(goal, 1)).toBe(true);
		expect(isDoableAtPlace(goal, 2)).toBe(true);
		expect(isDoableAtPlace(goal, 3)).toBe(false);
	});

	it("utan vald plats är inget mål görbart", () => {
		expect(isDoableAtPlace({ places: [place(1, "Hem")] }, null)).toBe(false);
	});
});

describe("isDoableAtPlaceOrUnplaced", () => {
	it("ett mål utan platser är görbart överallt", () => {
		expect(isDoableAtPlaceOrUnplaced({}, 1)).toBe(true);
	});

	it("ett mål med platsen är görbart där", () => {
		expect(isDoableAtPlaceOrUnplaced({ places: [place(1, "Hem")] }, 1)).toBe(true);
	});

	it("ett mål på en annan plats är inte görbart", () => {
		expect(isDoableAtPlaceOrUnplaced({ places: [place(1, "Hem")] }, 2)).toBe(false);
	});

	it("flera platser räcker - det räcker att vara på en av dem", () => {
		const goal = { places: [place(1, "Hem"), place(2, "Kontor")] };
		expect(isDoableAtPlaceOrUnplaced(goal, 1)).toBe(true);
		expect(isDoableAtPlaceOrUnplaced(goal, 2)).toBe(true);
		expect(isDoableAtPlaceOrUnplaced(goal, 3)).toBe(false);
	});

	it("utan vald plats är inget mål görbart", () => {
		expect(isDoableAtPlaceOrUnplaced({ places: [place(1, "Hem")] }, null)).toBe(false);
	});
});

describe("isDoableAtAnyPlace", () => {
	it("räcker att en av platserna stämmer", () => {
		const goal = { places: [place(2, "Kontor")] };
		expect(isDoableAtAnyPlace(goal, [1, 2])).toBe(true);
		expect(isDoableAtAnyPlace(goal, [1, 3])).toBe(false);
		expect(isDoableAtAnyPlace(goal, [])).toBe(false);
	});
});

describe("placeNames", () => {
	it("listar platsernas namn", () => {
		const goal = { places: [place(1, "Hem"), place(2, "Kontor")] };
		expect(placeNames(goal)).toBe("Hem, Kontor");
	});

	it("ett mål utan platser ger tom sträng", () => {
		expect(placeNames({})).toBe("");
	});
});
