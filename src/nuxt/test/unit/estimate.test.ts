import { describe, expect, it } from "bun:test";
import {
	formatEstimate,
	parseEstimateMinutes,
} from "../../app/utils/estimate";

describe("parseEstimateMinutes", () => {
	it("tolkar ett rent antal som minuter", () => {
		expect(parseEstimateMinutes("90")).toBe(90);
		expect(parseEstimateMinutes(45)).toBe(45);
	});

	it("tolkar timmar och minuter", () => {
		expect(parseEstimateMinutes("1h")).toBe(60);
		expect(parseEstimateMinutes("1 h")).toBe(60);
		expect(parseEstimateMinutes("1h30")).toBe(90);
		expect(parseEstimateMinutes("1h30m")).toBe(90);
		expect(parseEstimateMinutes("2h")).toBe(120);
	});

	it("tolkar svenska decimaltecken", () => {
		expect(parseEstimateMinutes("1,5h")).toBe(90);
		expect(parseEstimateMinutes("0,5 h")).toBe(30);
	});

	it("tolkar minuter", () => {
		expect(parseEstimateMinutes("45m")).toBe(45);
		expect(parseEstimateMinutes("45 min")).toBe(45);
	});

	it("returnerar null för tomma eller ogiltiga värden", () => {
		expect(parseEstimateMinutes("")).toBeNull();
		expect(parseEstimateMinutes(null)).toBeNull();
		expect(parseEstimateMinutes(undefined)).toBeNull();
		expect(parseEstimateMinutes("abc")).toBeNull();
		expect(parseEstimateMinutes("-30")).toBeNull();
		expect(parseEstimateMinutes(-30)).toBeNull();
	});
});

describe("formatEstimate", () => {
	it("visar timmar och minuter", () => {
		expect(formatEstimate(90)).toBe("1 h 30 min");
		expect(formatEstimate(45)).toBe("45 min");
		expect(formatEstimate(120)).toBe("2 h");
		expect(formatEstimate(60)).toBe("1 h");
	});

	it("returnerar tom sträng när det saknas uppskattning", () => {
		expect(formatEstimate(null)).toBe("");
		expect(formatEstimate(undefined)).toBe("");
		expect(formatEstimate(0)).toBe("");
	});
});
