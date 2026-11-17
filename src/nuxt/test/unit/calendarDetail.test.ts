import { describe, expect, it } from "bun:test";

import type { CalendarEventLike } from "../../app/utils/calendarDetail";
import {
	describeRecurrence,
	groupEventsByMonth,
	groupEventsByYear,
	recurrenceSeriesKey,
} from "../../app/utils/calendarDetail";

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

describe("groupEventsByMonth", () => {
	it("räknar varje förekomst i en serie", () => {
		const groups = groupEventsByMonth([
			event({ id: 7, recurrence_rule: "FREQ=DAILY" }),
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				start_at: "2027-03-02T09:00:00.000Z",
				end_at: "2027-03-02T10:00:00.000Z",
			}),
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				start_at: "2027-03-03T09:00:00.000Z",
				end_at: "2027-03-03T10:00:00.000Z",
			}),
		]);

		expect(groups).toHaveLength(1);
		expect(groups[0].count).toBe(3);
		expect(groups[0].recurring).toBe(true);
	});

	it("skriver ut serien en gång med sitt antal", () => {
		const groups = groupEventsByMonth([
			event({ id: 7, recurrence_rule: "FREQ=WEEKLY" }),
			event({
				id: 7,
				recurrence_rule: "FREQ=WEEKLY",
				start_at: "2027-03-08T09:00:00.000Z",
				end_at: "2027-03-08T10:00:00.000Z",
			}),
		]);

		expect(groups[0].title).toContain("Löpning ×2");
	});

	it("räknar enstaka händelser var för sig", () => {
		const groups = groupEventsByMonth([
			event({ id: 1, title: "Fika" }),
			event({ id: 2, title: "Lunch" }),
		]);

		expect(groups[0].count).toBe(2);
		expect(groups[0].title).toContain("Fika");
		expect(groups[0].title).toContain("Lunch");
		expect(groups[0].recurring).toBe(false);
	});

	it("behåller förekomster i senare månader", () => {
		const groups = groupEventsByMonth([
			event({ id: 7, recurrence_rule: "FREQ=DAILY" }),
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				start_at: "2027-04-05T09:00:00.000Z",
				end_at: "2027-04-05T10:00:00.000Z",
			}),
		]);

		expect(groups).toHaveLength(2);
		expect(groups[0].count).toBe(1);
		expect(groups[1].count).toBe(1);
	});
});

describe("groupEventsByYear", () => {
	it("räknar alla förekomster i ett år", () => {
		const groups = groupEventsByYear([
			event({ id: 7, recurrence_rule: "FREQ=DAILY" }),
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				start_at: "2027-04-05T09:00:00.000Z",
				end_at: "2027-04-05T10:00:00.000Z",
			}),
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				start_at: "2028-04-05T09:00:00.000Z",
				end_at: "2028-04-05T10:00:00.000Z",
			}),
		]);

		expect(groups).toHaveLength(2);
		expect(groups[0].count).toBe(2);
		expect(groups[1].count).toBe(1);
	});
});
