import { describe, expect, it } from "bun:test";
import { todayDateString } from "../../app/utils/dueDate";
import {
	describeGoalRecurrence,
	isRecurringGoal,
	isWaitingForNextOccurrence,
	recurrenceCount,
} from "../../app/utils/recurrence";

// Datumformatet på samma sätt som appen: lokalt kalenderdatum.
const localDate = (offsetDays: number): string => {
	const date = new Date();
	date.setDate(date.getDate() + offsetDays);
	const pad = (value: number) => String(value).padStart(2, "0");
	return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
};

describe("isRecurringGoal", () => {
	it("säger att ett mål med regel är återkommande", () => {
		expect(isRecurringGoal({ recurrence_rule: "FREQ=DAILY" })).toBe(true);
	});

	it("säger att ett mål utan regel inte är återkommande", () => {
		expect(isRecurringGoal({ recurrence_rule: null })).toBe(false);
		expect(isRecurringGoal({ recurrence_rule: "" })).toBe(false);
	});
});

describe("isWaitingForNextOccurrence", () => {
	it("väntar när nästa förekomst ligger i framtiden", () => {
		expect(
			isWaitingForNextOccurrence({
				recurrence_rule: "FREQ=DAILY",
				recurrence_next: localDate(1),
			}),
		).toBe(true);
	});

	it("väntar inte när nästa förekomst är idag", () => {
		expect(
			isWaitingForNextOccurrence({
				recurrence_rule: "FREQ=DAILY",
				recurrence_next: localDate(0),
			}),
		).toBe(false);
	});

	it("väntar inte när förekomsten är förfallen", () => {
		expect(
			isWaitingForNextOccurrence({
				recurrence_rule: "FREQ=DAILY",
				recurrence_next: localDate(-1),
			}),
		).toBe(false);
	});

	it("väntar inte när serien är slut", () => {
		expect(
			isWaitingForNextOccurrence({
				recurrence_rule: "FREQ=DAILY",
				recurrence_next: null,
			}),
		).toBe(false);
	});

	it("väntar inte för ett vanligt mål", () => {
		expect(
			isWaitingForNextOccurrence({
				recurrence_rule: null,
				recurrence_next: "2030-01-01",
			}),
		).toBe(false);
	});
});

describe("recurrenceCount", () => {
	it("räknar slutförda förekomster", () => {
		expect(recurrenceCount({ recurrence_count: 3 })).toBe(3);
	});

	it("räknar noll när antalet saknas", () => {
		expect(recurrenceCount({ recurrence_count: undefined })).toBe(0);
	});
});

describe("describeGoalRecurrence", () => {
	it("beskriver daglig upprepning", () => {
		expect(describeGoalRecurrence("FREQ=DAILY")).toBe("varje dag");
	});

	it("beskriver varannan dag", () => {
		expect(describeGoalRecurrence("FREQ=DAILY;INTERVAL=2")).toBe("var annan dag");
	});

	it("beskriver veckodag", () => {
		expect(describeGoalRecurrence("FREQ=WEEKLY;BYDAY=MO")).toBe("varje måndag");
	});

	it("beskriver varannan vecka", () => {
		expect(describeGoalRecurrence("FREQ=WEEKLY;INTERVAL=2")).toBe("var annan vecka");
	});

	it("beskriver månad och år", () => {
		expect(describeGoalRecurrence("FREQ=MONTHLY")).toBe("varje månad");
		expect(describeGoalRecurrence("FREQ=YEARLY")).toBe("varje år");
	});

	it("tar bort RRULE-prefixet", () => {
		expect(describeGoalRecurrence("RRULE:FREQ=DAILY")).toBe("varje dag");
	});

	it("visar okända regler som de är", () => {
		expect(describeGoalRecurrence("FREQ=BOGUS")).toBe("FREQ=BOGUS");
	});

	it("ger tom sträng utan regel", () => {
		expect(describeGoalRecurrence(null)).toBe("");
	});
});
