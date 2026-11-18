import { describe, expect, it } from "bun:test";

import type { CalendarEventLike } from "../../app/utils/calendarDetail";
import {
	describeRecurrence,
	groupEventsByDay,
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

describe("viktade förekomster", () => {
	it("räknar vikterna så att antalet blir exakt", () => {
		const groups = groupEventsByMonth([
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				occurrence_weight: 12,
			}),
		]);

		expect(groups[0].count).toBe(12);
	});

	it("räknar en serie och en enstaka händelse var för sig", () => {
		const groups = groupEventsByMonth([
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				occurrence_weight: 12,
			}),
			event({ id: 8, title: "Fika" }),
		]);

		expect(groups[0].count).toBe(13);
		expect(groups[0].entries).toHaveLength(2);
	});

	it("märker markören som ett urval när en serie tunnats ut", () => {
		const groups = groupEventsByMonth([
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				occurrence_weight: 12,
			}),
		]);

		expect(groups[0].approximate).toBe(true);
	});

	it("lämnar en markör utan urval när inget tunnats ut", () => {
		const groups = groupEventsByMonth([event({ id: 1, title: "Fika" })]);

		expect(groups[0].approximate).toBe(false);
	});

	it("skriver ut att urvalet visas i tooltipen", () => {
		const groups = groupEventsByMonth([
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				occurrence_weight: 12,
				occurrence_period: "month",
			}),
		]);

		expect(groups[0].title).toContain("visar 1 av 12 per månad");
	});

	it("skriver ut vilken period serien tunnats ut i", () => {
		const groups = groupEventsByMonth([
			event({
				id: 7,
				recurrence_rule: "FREQ=HOURLY",
				occurrence_weight: 24,
				occurrence_period: "day",
			}),
		]);

		expect(groups[0].title).toContain("visar 1 av 24 per dag");
	});

	it("räknar uttunnade förekomster som en post i serien", () => {
		const groups = groupEventsByMonth([
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				occurrence_weight: 6,
			}),
			event({
				id: 7,
				recurrence_rule: "FREQ=DAILY",
				start_at: "2027-03-02T09:00:00.000Z",
				end_at: "2027-03-02T10:00:00.000Z",
				occurrence_weight: 6,
			}),
		]);

		expect(groups[0].entries).toHaveLength(1);
		expect(groups[0].entries[0].count).toBe(12);
		expect(groups[0].entries[0].shown).toBe(2);
		expect(groups[0].title).toContain("Löpning ×12");
	});
});

describe("serier under samma prick", () => {
	it("listar varje serie för sig", () => {
		const groups = groupEventsByDay([
			event({ id: 7, recurrence_rule: "FREQ=DAILY", title: "Löpning" }),
			event({
				id: 8,
				recurrence_rule: "FREQ=WEEKLY",
				title: "Styrning",
				start_at: "2027-03-01T14:00:00.000Z",
				end_at: "2027-03-01T15:00:00.000Z",
			}),
			event({
				id: 9,
				title: "Fika",
				start_at: "2027-03-01T15:00:00.000Z",
				end_at: "2027-03-01T16:00:00.000Z",
			}),
		]);

		expect(groups).toHaveLength(1);
		expect(groups[0].entries.map((entry) => entry.title)).toEqual([
			"Fika",
			"Löpning",
			"Styrning",
		]);
		expect(groups[0].count).toBe(3);
	});

	it("märker serier och enstaka händelser var för sig", () => {
		const groups = groupEventsByDay([
			event({ id: 7, recurrence_rule: "FREQ=DAILY", title: "Löpning" }),
			event({
				id: 9,
				title: "Fika",
				start_at: "2027-03-01T15:00:00.000Z",
				end_at: "2027-03-01T16:00:00.000Z",
			}),
		]);

		const recurring = groups[0].entries.filter((entry) => entry.recurring);
		const single = groups[0].entries.filter((entry) => !entry.recurring);

		expect(recurring).toHaveLength(1);
		expect(single).toHaveLength(1);
		expect(single[0].title).toBe("Fika");
	});

	it("räknar en override som en del av serien", () => {
		const groups = groupEventsByDay([
			event({ id: 7, recurrence_rule: "FREQ=DAILY", title: "Löpning" }),
			event({
				id: 8,
				title: "Löpning",
				parent_event_id: 7,
				recurrence_rule: "FREQ=DAILY",
				start_at: "2027-03-01T18:00:00.000Z",
				end_at: "2027-03-01T19:00:00.000Z",
			}),
		]);

		expect(groups[0].entries).toHaveLength(1);
		expect(groups[0].entries[0].count).toBe(2);
	});
});
