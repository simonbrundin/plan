import { beforeEach, describe, expect, it } from "bun:test";

import { useGoalVerification } from "../../app/composables/useGoalVerification";

type Anrop = { url: string; method: string; body?: unknown };

const anrop: Anrop[] = [];
let svar: (url: string, options: any) => unknown = () => ({});

// Nuxt-funktionerna importeras automatiskt i appen. I enhetstestet
// sätts de på globalen - annars finns ingen runtime.
(globalThis as any).useRuntimeConfig = () => ({
	public: { goApiUrl: "/api/v1" },
});

beforeEach(() => {
	anrop.length = 0;
	svar = () => ({});
	(globalThis as any).$fetch = async (url: string, options: any = {}) => {
		anrop.push({
			url,
			method: options?.method ?? "GET",
			body: options?.body,
		});
		return svar(url, options);
	};
});

describe("useGoalVerification", () => {
	it("läser plan, bevis och omdöme för ett mål", async () => {
		svar = () => ({
			verification: { id: 3, goal_id: 540, type: "checklista", criterion: "Alla punkter" },
			evidence: [{ id: 1, goal_id: 540, type: "text", value: "Kollade av" }],
			result: { id: 9, goal_id: 540, outcome: "passed" },
		});

		const api = useGoalVerification();
		const state = await api.loadVerification(540);

		expect(anrop[0].url).toBe("/api/v1/goals/540/verification");
		expect(state.verification?.type).toBe("checklista");
		expect(state.evidence).toHaveLength(1);
		expect(state.result?.outcome).toBe("passed");
	});

	it("sätter verifieringsplanen", async () => {
		svar = () => ({});

		const api = useGoalVerification();
		await api.saveVerification(540, {
			type: "bevis",
			criterion: "Bevis finns",
			strength: "stark",
		});

		expect(anrop[0].method).toBe("PUT");
		expect(anrop[0].url).toBe("/api/v1/goals/540/verification");
		expect((anrop[0].body as any).criterion).toBe("Bevis finns");
	});

	it("tar bort verifieringsplanen", async () => {
		svar = () => ({});

		const api = useGoalVerification();
		await api.removeVerification(540);

		expect(anrop[0].method).toBe("DELETE");
		expect(anrop[0].url).toBe("/api/v1/goals/540/verification");
	});

	it("lämnar bevis", async () => {
		svar = () => ({});

		const api = useGoalVerification();
		await api.addEvidence(540, { type: "text", value: "Klart" });

		expect(anrop[0].method).toBe("POST");
		expect(anrop[0].url).toBe("/api/v1/goals/540/evidence");
		expect((anrop[0].body as any).value).toBe("Klart");
	});

	it("kör verifieringen och godkänner omdömet", async () => {
		svar = () => ({ id: 9, goal_id: 540, outcome: "passed" });

		const api = useGoalVerification();
		await api.runVerification(540);
		await api.approveResult(9);

		expect(anrop[0].method).toBe("POST");
		expect(anrop[0].url).toBe("/api/v1/goals/540/verification/run");
		expect(anrop[1].method).toBe("POST");
		expect(anrop[1].url).toBe("/api/v1/verification-results/9/approve");
	});

	it("avslutar målet och läser tillbaka hur det avslutades", async () => {
		svar = () => ({ goal_id: 540, from: "verify", to: "completed", completion_kind: "verifierat" });

		const api = useGoalVerification();
		const utfall = await api.completeGoal(540);

		expect(anrop[0].method).toBe("POST");
		expect(anrop[0].url).toBe("/api/v1/goals/540/complete");
		expect(utfall.completion_kind).toBe("verifierat");
	});
});
