import { describe, expect, it } from "bun:test";

import type { CalendarEventLike } from "../../app/utils/calendarSeries";
import {
	describeRecurrence,
	isDailyRecurrence,
	recurrenceSeriesKey,
} from "../../app/utils/calendarSeries";

const event = (
	overrides: Partial<CalendarEventLike> = {},
): CalendarEventLike => ({
	id: 1,
	title: "Löpning",
	start_at: "2027-03-01T09:00:00.000Z",
	end_at: "2027-03-01T10:00:00.000Z",
	...overrides,
});

describe("describeRecurrence", () => {
	it("beskriver en veckoregel", () => {
		expect(describeRecurrence("FREQ=WEEKLY")).toBe("varje vecka");
	});

	it("beskriver en regel med intervall", () => {
		expect(describeRecurrence("FREQ=DAILY;INTERVAL=3")).toBe("var 3:e dag");
	});

	it("beskriver en regel med flera delar", () => {
		expect(describeRecurrence("FREQ=WEEKLY;BYDAY=MO,WE")).toBe("varje vecka");
	});

	it("returnerar tom etikett när regeln saknas", () => {
		expect(describeRecurrence(null)).toBe("");
		expect(describeRecurrence("")).toBe("");
	});
});

describe("recurrenceSeriesKey", () => {
	it("ger förekomster i samma serie samma nyckel", () => {
		const first = event({ id: 7, recurrence_rule: "FREQ=WEEKLY" });
		const second = event({
			id: 7,
			recurrence_rule: "FREQ=WEEKLY",
			start_at: "2027-03-08T09:00:00.000Z",
			end_at: "2027-03-08T10:00:00.000Z",
		});

		expect(recurrenceSeriesKey(first)).toBe("series:7");
		expect(recurrenceSeriesKey(second)).toBe(recurrenceSeriesKey(first));
	});

	it("räknar en override som en del av serien", () => {
		const override = event({ id: 99, parent_event_id: 7 });

		expect(recurrenceSeriesKey(override)).toBe("series:7");
	});

	it("lämnar enstaka händelser utan serie", () => {
		expect(recurrenceSeriesKey(event())).toBeNull();
	});
});

describe("isDailyRecurrence", () => {
	it("känner igen en daglig regel", () => {
		expect(isDailyRecurrence("FREQ=DAILY")).toBe(true);
		expect(isDailyRecurrence("FREQ=DAILY;INTERVAL=1")).toBe(true);
	});

	it("räknar inte varannan dag som daglig", () => {
		expect(isDailyRecurrence("FREQ=DAILY;INTERVAL=2")).toBe(false);
	});

	it("räknar inte veckoseries som dagliga", () => {
		expect(isDailyRecurrence("FREQ=WEEKLY")).toBe(false);
	});

	it("returnerar falskt när regeln saknas", () => {
		expect(isDailyRecurrence(null)).toBe(false);
		expect(isDailyRecurrence("")).toBe(false);
	});
});
