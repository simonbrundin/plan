import { describe, expect, it } from "bun:test";

import {
	durationOptions,
	millisecondsPerPixel,
	parseGoalDrag,
	snapStepMs,
	snapTimestamp,
	serializeGoalDrag,
	shiftEventTimes,
	timeAtOffset,
} from "../../app/utils/calendarScheduling";

describe("parseGoalDrag", () => {
	it("läser tillbaka ett mål", () => {
		const payload = serializeGoalDrag({
			id: 7,
			title: "Springa",
			icon: "heroicons:star",
			estimated_minutes: 90,
		});

		expect(parseGoalDrag(payload)).toEqual({
			id: 7,
			title: "Springa",
			icon: "heroicons:star",
			estimated_minutes: 90,
		});
	});

	it("tar null när texten inte är ett mål", () => {
		expect(parseGoalDrag(null)).toBeNull();
		expect(parseGoalDrag("inte json")).toBeNull();
		expect(parseGoalDrag("{}")).toBeNull();
		expect(parseGoalDrag('{"id":"sju","title":"Springa"}')).toBeNull();
		expect(parseGoalDrag('{"id":7}')).toBeNull();
	});

	it("hanterar en uppskattning som saknas", () => {
		const payload = serializeGoalDrag({ id: 1, title: "Läsa" });

		expect(parseGoalDrag(payload)?.estimated_minutes).toBeNull();
	});
});

describe("durationOptions", () => {
	it("sätter målets egen uppskattning först", () => {
		const options = durationOptions(90);

		expect(options[0]).toEqual({
			minutes: 90,
			label: "1 h 30 min",
			suggested: true,
		});
	});

	it("bjuder på de vanligaste längderna", () => {
		const options = durationOptions(null);
		const minutes = options.map((option) => option.minutes);

		expect(minutes).toContain(15);
		expect(minutes).toContain(30);
		expect(minutes).toContain(60);
		expect(minutes).toContain(120);
		expect(minutes).toContain(240);
	});

	it("tar inte med samma längd två gånger", () => {
		const options = durationOptions(60);
		const minutes = options.map((option) => option.minutes);

		expect(minutes.filter((value) => value === 60)).toHaveLength(1);
	});

	it("ignorerar en omöjlig uppskattning", () => {
		expect(durationOptions(0).every((option) => option.minutes > 0)).toBe(true);
		expect(durationOptions(-5).every((option) => option.minutes > 0)).toBe(true);
	});
});

describe("timeAtOffset", () => {
	const base = {
		startTime: 1_000_000,
		visibleDuration: 24 * 60 * 60 * 1000,
		height: 100,
	};

	it("ger starttiden vid överkanten", () => {
		expect(timeAtOffset({ ...base, offsetY: 0 })).toBe(base.startTime);
	});

	it("ger slutetiden vid underkanten", () => {
		expect(timeAtOffset({ ...base, offsetY: 100 })).toBe(
			base.startTime + base.visibleDuration,
		);
	});

	it("ger halva tiden vid halva höjden", () => {
		expect(timeAtOffset({ ...base, offsetY: 50 })).toBe(
			base.startTime + base.visibleDuration / 2,
		);
	});

	it("klipper in positioner utanför vyn", () => {
		expect(timeAtOffset({ ...base, offsetY: -50 })).toBe(base.startTime);
		expect(timeAtOffset({ ...base, offsetY: 500 })).toBe(
			base.startTime + base.visibleDuration,
		);
	});

	it("hanterar en vyn utan höjd", () => {
		expect(
			timeAtOffset({ ...base, height: 0, offsetY: 50 }),
		).toBe(base.startTime);
	});
});

describe("millisecondsPerPixel", () => {
	it("räknar ut tiden per pixel", () => {
		expect(millisecondsPerPixel(1000, 100)).toBe(10);
	});

	it("returnerar noll utan höjd", () => {
		expect(millisecondsPerPixel(1000, 0)).toBe(0);
	});
});

describe("snapStepMs", () => {
	it("snappar på minuter i en kort vy", () => {
		expect(snapStepMs(24 * 60 * 60 * 1000)).toBe(15 * 60 * 1000);
	});

	it("snappar på timmar i en månadsvy", () => {
		expect(snapStepMs(30 * 24 * 60 * 60 * 1000)).toBe(60 * 60 * 1000);
	});

	it("snappar på dygn i en årsvy", () => {
		expect(snapStepMs(365 * 24 * 60 * 60 * 1000)).toBe(24 * 60 * 60 * 1000);
	});
});

describe("snapTimestamp", () => {
	it("snappar till närmaste steg", () => {
		expect(snapTimestamp(1_000_000, 15 * 60 * 1000)).toBe(
			Math.round(1_000_000 / (15 * 60 * 1000)) * (15 * 60 * 1000),
		);
	});

	it("lämnar tiden som den är utan steg", () => {
		expect(snapTimestamp(1234, 0)).toBe(1234);
	});
});

describe("shiftEventTimes", () => {
	it("flyttar både början och slut", () => {
		const moved = shiftEventTimes({
			start: 1000,
			end: 2000,
			deltaMs: 500,
		});

		expect(moved).toEqual({ start: 1500, end: 2500 });
	});

	it("behåller händelsens längd", () => {
		const moved = shiftEventTimes({
			start: 1000,
			end: 2000,
			deltaMs: -250,
		});

		expect(moved.end - moved.start).toBe(1000);
	});
});
