import { describe, expect, it } from "bun:test";
import { dueTone, formatDueDate, parseDueDate } from "../../app/utils/dueDate";

describe("parseDueDate", () => {
	it("tolkar datumet lokalt, utan tidszon", () => {
		const date = parseDueDate("2026-11-20");
		expect(date).not.toBeNull();
		expect(date?.getFullYear()).toBe(2026);
		expect(date?.getMonth()).toBe(10);
		expect(date?.getDate()).toBe(20);
	});

	it("returnerar null för tomma eller ogiltiga värden", () => {
		expect(parseDueDate(null)).toBeNull();
		expect(parseDueDate("")).toBeNull();
		expect(parseDueDate("2026-11-20T10:00:00Z")).toBeNull();
		expect(parseDueDate("inte ett datum")).toBeNull();
	});
});

describe("formatDueDate", () => {
	it("visar kort datumformat", () => {
		expect(formatDueDate("2026-11-20")).toBe("20 nov.");
	});

	it("returnerar tom sträng när det saknas datum", () => {
		expect(formatDueDate(null)).toBe("");
	});
});

describe("dueTone", () => {
	it("räknar datum som förgått, idag eller kommande", () => {
		const today = new Date();
		const iso = (date: Date) =>
			`${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;

		const yesterday = new Date(today);
		yesterday.setDate(today.getDate() - 1);
		expect(dueTone(iso(yesterday))).toBe("overdue");

		expect(dueTone(iso(today))).toBe("today");

		const tomorrow = new Date(today);
		tomorrow.setDate(today.getDate() + 1);
		expect(dueTone(iso(tomorrow))).toBe("upcoming");
	});
});
