import { describe, expect, it } from "bun:test";

import {
	describeCompletion,
	getVerificationOutcomePresentation,
	getVerificationStrengthPresentation,
	getVerificationTypePresentation,
	needsApproval,
} from "../../app/utils/goalVerification";

describe("goalVerification", () => {
	it("känner igen alla verifieringstyper", () => {
		for (const typ of [
			"checklista",
			"delmal",
			"kalender",
			"bevis",
			"matvarde",
			"kod",
		]) {
			expect(getVerificationTypePresentation(typ).label).not.toBe(
				"Okänd typ (undefined)",
			);
			expect(getVerificationTypePresentation(typ).label).not.toContain(
				"Okänd typ",
			);
		}
	});

	it("visar en okänd typ som okänd i stället för att krascha", () => {
		const presentation = getVerificationTypePresentation("troligen");
		expect(presentation.label).toContain("troligen");
		expect(presentation.colorClass).toBe("text-ink-dim");
	});

	it("visar en plan utan typ som 'Ingen plan'", () => {
		expect(getVerificationTypePresentation(null).label).toBe("Ingen plan");
		expect(getVerificationTypePresentation(undefined).label).toBe("Ingen plan");
	});

	it("känner igen styrkorna", () => {
		expect(getVerificationStrengthPresentation("stark").label).toBe(
			"Starkt bevis",
		);
		expect(getVerificationStrengthPresentation("svag").label).toBe(
			"Svagt bevis",
		);
	});

	it("känner igen omdömena", () => {
		expect(getVerificationOutcomePresentation("passed").label).toBe(
			"Godkänd",
		);
		expect(getVerificationOutcomePresentation("failed").label).toBe(
			"Underkänd",
		);
		expect(getVerificationOutcomePresentation("pending").label).toBe(
			"Väntar på bevis",
		);
		expect(getVerificationOutcomePresentation(null).label).toBe(
			"Inget omdöme",
		);
	});

	it("beskriver hur målet avslutades", () => {
		expect(describeCompletion("verifierat")?.label).toBe("Verifierat");
		expect(describeCompletion("deklarerat")?.label).toBe("Deklarerat");
		expect(describeCompletion("")).toBeNull();
		expect(describeCompletion(null)).toBeNull();
	});

	it("kräver godkännande så länge omdömet inte är beslutat", () => {
		expect(needsApproval(null)).toBe(false);
		expect(
			needsApproval({
				id: 1,
				goal_id: 2,
				outcome: "passed",
				created_at: "2026-11-28",
			}),
		).toBe(true);
		expect(
			needsApproval({
				id: 1,
				goal_id: 2,
				outcome: "passed",
				decided_at: "2026-11-28",
				created_at: "2026-11-28",
			}),
		).toBe(false);
	});
});
