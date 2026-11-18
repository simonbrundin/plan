import { describe, expect, it } from "bun:test";
import {
	duplicateEventPayload,
	isSeriesOccurrence,
	parentEventId,
	toRRuleDate,
	truncateRecurrenceRule,
} from "../../app/utils/calendarEvents";
import type { CalendarEvent } from "../../app/composables/useCalendarApi";

const event = (overrides: Partial<CalendarEvent> = {}): CalendarEvent => ({
	id: 1,
	title: "Löpning",
	description: "Varje morgon",
	icon: "🏃",
	start_at: "2026-11-18T09:00:00Z",
	end_at: "2026-11-18T10:00:00Z",
	time_zone: "Europe/Stockholm",
	all_day: false,
	status: "confirmed",
	calendar_id: 7,
	goal_id: 3,
	...overrides,
});

describe("isSeriesOccurrence", () => {
	it("säger nej till en vanlig händelse", () => {
		expect(isSeriesOccurrence(event())).toBe(false);
	});

	it("säger ja till en förekomst i en serie", () => {
		const occurrence = event({
			occurrence_id: "1-1763456400000",
			recurrence_rule: "FREQ=DAILY",
		});

		expect(isSeriesOccurrence(occurrence)).toBe(true);
	});

	it("säger ja till en förekomst med egen rad", () => {
		const override = event({
			id: 42,
			parent_event_id: 1,
			original_start_at: "2026-11-18T09:00:00Z",
		});

		expect(isSeriesOccurrence(override)).toBe(true);
	});
});

describe("parentEventId", () => {
	it("pekar på seriens masterrad för en förekomst", () => {
		const occurrence = event({ occurrence_id: "1-1763456400000" });

		expect(parentEventId(occurrence)).toBe(1);
	});

	it("pekar på föräldern för en förekomst med egen rad", () => {
		const override = event({ id: 42, parent_event_id: 1 });

		expect(parentEventId(override)).toBe(1);
	});

	it("pekar på händelsen själv när den inte hör till en serie", () => {
		expect(parentEventId(event({ id: 9 }))).toBe(9);
	});
});

describe("duplicateEventPayload", () => {
	it("kopierar titel, beskrivning, ikon, kalender och mål", () => {
		const payload = duplicateEventPayload(event());

		expect(payload.title).toBe("Löpning");
		expect(payload.description).toBe("Varje morgon");
		expect(payload.icon).toBe("🏃");
		expect(payload.calendar_id).toBe(7);
		expect(payload.goal_id).toBe(3);
		expect(payload.start_at).toBe("2026-11-18T09:00:00Z");
		expect(payload.end_at).toBe("2026-11-18T10:00:00Z");
	});

	it("lämnar kalender och mål tomma när händelsen saknar dem", () => {
		const payload = duplicateEventPayload(
			event({ calendar_id: undefined, goal_id: undefined }),
		);

		expect("calendar_id" in payload).toBe(false);
		expect("goal_id" in payload).toBe(false);
	});

	it("kopplar loss en förekomst från sin serie", () => {
		const payload = duplicateEventPayload(
			event({
				occurrence_id: "1-1763456400000",
				recurrence_rule: "FREQ=DAILY",
			}),
		);

		expect("recurrence_rule" in payload).toBe(false);
	});

	it("behåller upprepningsregeln när hela serien kopieras", () => {
		const payload = duplicateEventPayload(
			event({ recurrence_rule: "FREQ=DAILY" }),
		);

		expect(payload.recurrence_rule).toBe("FREQ=DAILY");
	});
});

describe("truncateRecurrenceRule", () => {
	it("lägger till en UNTIL-gräns", () => {
		const rule = truncateRecurrenceRule(
			"FREQ=WEEKLY",
			new Date("2026-11-18T08:59:59Z"),
		);

		expect(rule).toBe("FREQ=WEEKLY;UNTIL=20261118T085959Z");
	});

	it("tar bort en befintlig UNTIL", () => {
		const rule = truncateRecurrenceRule(
			"FREQ=DAILY;UNTIL=20270101T000000Z",
			new Date("2026-11-18T08:59:59Z"),
		);

		expect(rule).toBe("FREQ=DAILY;UNTIL=20261118T085959Z");
	});

	it("tar bort en befintlig COUNT", () => {
		const rule = truncateRecurrenceRule(
			"FREQ=DAILY;COUNT=10",
			new Date("2026-11-18T08:59:59Z"),
		);

		expect(rule).toBe("FREQ=DAILY;UNTIL=20261118T085959Z");
	});

	it("behåller resten av regeln", () => {
		const rule = truncateRecurrenceRule(
			"FREQ=WEEKLY;INTERVAL=2;BYDAY=MO",
			new Date("2026-11-18T08:59:59Z"),
		);

		expect(rule).toBe(
			"FREQ=WEEKLY;INTERVAL=2;BYDAY=MO;UNTIL=20261118T085959Z",
		);
	});
});

describe("toRRuleDate", () => {
	it("skriver tiden i RRULE-format", () => {
		expect(toRRuleDate(new Date("2026-11-18T09:00:00Z"))).toBe(
			"20261118T090000Z",
		);
	});
});
