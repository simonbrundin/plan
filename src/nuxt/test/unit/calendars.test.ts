import { describe, expect, it } from "bun:test";
import {
	findDefaultCalendar,
	findDefaultCalendarId,
} from "../../app/utils/calendars";

const calendar = (id: number, isDefault = false) =>
	({
		id,
		user_id: 1,
		name: `Kalender ${id}`,
		color: "#6366f1",
		is_default: isDefault,
		sort_order: 0,
	}) as const;

describe("findDefaultCalendar", () => {
	it("hittar kalendern som är satt som standard", () => {
		const calendars = [calendar(1), calendar(2, true)];

		expect(findDefaultCalendar(calendars)?.id).toBe(2);
	});

	it("returnerar null när ingen kalender är standard", () => {
		expect(findDefaultCalendar([calendar(1), calendar(2)])).toBeNull();
	});

	it("returnerar null när listan är tom", () => {
		expect(findDefaultCalendar([])).toBeNull();
	});
});

describe("findDefaultCalendarId", () => {
	it("returnerar id:t för standardkalendern", () => {
		expect(findDefaultCalendarId([calendar(1, true), calendar(2)])).toBe(1);
	});

	it("returnerar null när ingen kalender är standard", () => {
		expect(findDefaultCalendarId([calendar(1), calendar(2)])).toBeNull();
	});
});
