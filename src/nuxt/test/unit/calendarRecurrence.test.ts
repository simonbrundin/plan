import { describe, expect, it } from "bun:test";

import type { CalendarEvent } from "../../app/composables/useCalendarApi";
import {
	CalendarExpander,
	expandCalendarEvents,
	expandSeries,
} from "../../app/utils/calendarRecurrence";

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

const event = (overrides: Partial<CalendarEvent> = {}): CalendarEvent => ({
	id: 1,
	title: "Händelse",
	description: "",
	start_at: "2026-11-17T09:00:00.000Z",
	end_at: "2026-11-17T10:00:00.000Z",
	time_zone: "UTC",
	all_day: false,
	status: "confirmed",
	...overrides,
});

describe("expandSeries", () => {
	it("expanderar en daglig serie till en förekomst per dag", () => {
		const series = event({
			id: 7,
			title: "Löpning",
			recurrence_rule: "FREQ=DAILY",
		});

		const occurrences = expandSeries(
			series,
			Date.parse("2026-11-17T00:00:00.000Z"),
			Date.parse("2026-11-20T00:00:00.000Z"),
		);

		expect(occurrences.map((item) => item.start_at)).toEqual([
			"2026-11-17T09:00:00.000Z",
			"2026-11-18T09:00:00.000Z",
			"2026-11-19T09:00:00.000Z",
		]);
		expect(occurrences[0].original_start_at).toBe("2026-11-17T09:00:00.000Z");
		expect(occurrences[0].occurrence_id).toBe(
			`7-${Date.parse("2026-11-17T09:00:00.000Z")}`,
		);
	});

	it("begränsar serien till antalet förekomster i regeln", () => {
		const series = event({
			id: 8,
			recurrence_rule: "FREQ=DAILY;COUNT=3",
		});

		const occurrences = expandSeries(
			series,
			Date.parse("2026-11-17T00:00:00.000Z"),
			Date.parse("2026-12-17T00:00:00.000Z"),
		);

		expect(occurrences).toHaveLength(3);
	});

	it("returnerar inget för en händelse utan upprepningsregel", () => {
		expect(
			expandSeries(
				event(),
				Date.parse("2026-11-17T00:00:00.000Z"),
				Date.parse("2026-11-18T00:00:00.000Z"),
			),
		).toHaveLength(0);
	});
});

describe("expandCalendarEvents", () => {
	it("tar med enstaka händelser som ligger i fönstret", () => {
		const inside = event({ id: 1, title: "Fika" });
		const outside = event({
			id: 2,
			title: "Sent",
			start_at: "2026-12-01T09:00:00.000Z",
			end_at: "2026-12-01T10:00:00.000Z",
		});

		const events = expandCalendarEvents(
			[inside, outside],
			Date.parse("2026-11-17T00:00:00.000Z"),
			Date.parse("2026-11-18T00:00:00.000Z"),
		);

		expect(events.map((item) => item.title)).toEqual(["Fika"]);
	});

	it("ersätter en förekomst med sin override", () => {
		const series = event({
			id: 7,
			title: "Löpning",
			recurrence_rule: "FREQ=DAILY",
		});
		const override = event({
			id: 9,
			title: "Löpning (ute)",
			parent_event_id: 7,
			original_start_at: "2026-11-18T09:00:00.000Z",
			start_at: "2026-11-18T18:00:00.000Z",
			end_at: "2026-11-18T19:00:00.000Z",
		});

		const events = expandCalendarEvents(
			[series, override],
			Date.parse("2026-11-17T00:00:00.000Z"),
			Date.parse("2026-11-20T00:00:00.000Z"),
		);

		expect(events).toHaveLength(3);
		expect(events[1].title).toBe("Löpning (ute)");
		expect(events[1].start_at).toBe("2026-11-18T18:00:00.000Z");
	});

	it("tar bort en förekomst när override är avstängd", () => {
		const series = event({
			id: 7,
			title: "Löpning",
			recurrence_rule: "FREQ=DAILY",
		});
		const override = event({
			id: 9,
			title: "Löpning",
			parent_event_id: 7,
			original_start_at: "2026-11-18T09:00:00.000Z",
			status: "cancelled",
		});

		const events = expandCalendarEvents(
			[series, override],
			Date.parse("2026-11-17T00:00:00.000Z"),
			Date.parse("2026-11-20T00:00:00.000Z"),
		);

		expect(events.map((item) => item.start_at)).toEqual([
			"2026-11-17T09:00:00.000Z",
			"2026-11-19T09:00:00.000Z",
		]);
	});
});

describe("CalendarExpander", () => {
	it("expanderar serier lokalt utan att fråga servern", () => {
		const expander = new CalendarExpander();
		expander.setEvents([
			event({
				id: 7,
				title: "Löpning",
				recurrence_rule: "FREQ=DAILY",
			}),
		]);

		const from = Date.parse("2026-11-17T00:00:00.000Z");
		const to = Date.parse("2026-11-19T00:00:00.000Z");

		expect(expander.visible(from, to)).toHaveLength(2);
		// Samma fönster igen ger samma resultat - cachen används.
		expect(expander.visible(from, to)).toHaveLength(2);
		// Ett annat fönster expanderas om.
		expect(
			expander.visible(Date.parse("2026-11-18T00:00:00.000Z"), Date.parse("2026-11-20T00:00:00.000Z")),
		).toHaveLength(2);
	});

	it("glömmer gamla händelser när nya laddas", () => {
		const expander = new CalendarExpander();
		expander.setEvents([event({ id: 1, title: "Fika" })]);

		const from = Date.parse("2026-11-17T00:00:00.000Z");
		const to = Date.parse("2026-11-18T00:00:00.000Z");

		expect(expander.visible(from, to)).toHaveLength(1);

		expander.setEvents([event({ id: 2, title: "Möte" })]);

		expect(expander.visible(from, to)).toHaveLength(1);
		expect(expander.visible(from, to)[0].title).toBe("Möte");
	});
});
