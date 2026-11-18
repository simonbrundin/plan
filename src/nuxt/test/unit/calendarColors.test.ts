import { describe, expect, it } from "bun:test";
import {
	DEFAULT_EVENT_COLOR,
	isEventHidden,
	normalizeHexColor,
	resolveEventColor,
} from "../../app/utils/calendarColors";

const calendars = [
	{ id: 1, color: "#ef4444" },
	{ id: 2, color: "#10b981" },
];

describe("resolveEventColor", () => {
	it("ärver kalenderns färg", () => {
		const color = resolveEventColor({ calendar_id: 2 }, calendars);

		expect(color).toBe("#10b981");
	});

	it("faller tillbaka på standardfärgen för händelser utan kalender", () => {
		const color = resolveEventColor({}, calendars);

		expect(color).toBe(DEFAULT_EVENT_COLOR);
	});

	it("faller tillbaka på standardfärgen om kalendern inte finns", () => {
		const color = resolveEventColor({ calendar_id: 99 }, calendars);

		expect(color).toBe(DEFAULT_EVENT_COLOR);
	});

	it("fungerar med ett uppslag av kalendrar", () => {
		const map = new Map(calendars.map((calendar) => [calendar.id, calendar]));

		expect(resolveEventColor({ calendar_id: 1 }, map)).toBe("#ef4444");
	});
});

describe("isEventHidden", () => {
	it("döljer händelser i dolda kalendrar", () => {
		expect(isEventHidden({ calendar_id: 1 }, [1, 2])).toBe(true);
	});

	it("visar händelser i synliga kalendrar", () => {
		expect(isEventHidden({ calendar_id: 3 }, [1, 2])).toBe(false);
	});

	it("visar händelser utan kalender oavsett vad", () => {
		expect(isEventHidden({}, [1, 2])).toBe(false);
	});
});

describe("normalizeHexColor", () => {
	it("lämnar en giltig sexsiffrig kod som den är", () => {
		expect(normalizeHexColor("#6366f1")).toBe("#6366f1");
	});

	it("expanderar tre siffror till sex", () => {
		expect(normalizeHexColor("#abc")).toBe("#aabbcc");
	});

	it("sänker versaler", () => {
		expect(normalizeHexColor("#AABBCC")).toBe("#aabbcc");
	});

	it("trimmar mellanslag", () => {
		expect(normalizeHexColor("  #6366f1 ")).toBe("#6366f1");
	});

	it("avvisar en kod utan hash", () => {
		expect(normalizeHexColor("6366f1")).toBeNull();
	});

	it("avvisar fel längd", () => {
		expect(normalizeHexColor("#12345")).toBeNull();
	});

	it("avvisar ogiltiga tecken", () => {
		expect(normalizeHexColor("#6366fG")).toBeNull();
	});

	it("avvisar tomt värde", () => {
		expect(normalizeHexColor("")).toBeNull();
	});
});
