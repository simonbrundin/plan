import { apiRequestWithAuth } from "~/utils/api";
import type {
	GoalEvidence,
	VerificationPlan,
	VerificationResult,
} from "~/utils/goalVerification";

/**
 * Svaret på GET /goals/:id/verification.
 *
 * Planen är valfri - API:t svarar ändå med en tom plan, så att
 * klienten kan visa samma vy hela tiden.
 */
export interface VerificationState {
	verification: VerificationPlan;
	evidence: GoalEvidence[];
	result?: VerificationResult | null;
	/**
	 * Sant om API:t som kör inte stödjer verifieringen ännu.
	 *
	 * Rust-API:t svarar 501 på de här vägarna tills fasmaskinen och
	 * verifieringen portats. Då ska vyn säga det rakt ut i stället för
	 * att tyst ingenting hända.
	 */
	unsupported?: boolean;
}

/**
 * Svarar på om felet betyder att API:t saknar stödet för funktionen.
 */
export function isApiFeatureMissing(error: any): boolean {
	const status = error?.statusCode ?? error?.response?.status;
	return status === 501;
}

/**
 * API-anropen för verifiering och bevis.
 *
 * Planen är valfri - därför svarar GET med en tom plan i stället för
 * 404 när ett mål saknar plan.
 */
export function useGoalVerification() {
	const loadVerification = async (
		goalId: number,
	): Promise<VerificationState> => {
		try {
			const svar = await apiRequestWithAuth<VerificationState>(
				`goals/${goalId}/verification`,
			);
			return {
				verification: svar.verification,
				evidence: svar.evidence ?? [],
				result: svar.result ?? null,
			};
		} catch (error) {
			// 501 betyder att API:t som kör inte har den här funktionen
			// ännu - det är inte ett fel i användarens data.
			if (isApiFeatureMissing(error)) {
				return {
					verification: { id: 0, goal_id: goalId, type: "", criterion: "", strength: "" },
					evidence: [],
					result: null,
					unsupported: true,
				};
			}
			throw error;
		}
	};

	const saveVerification = async (
		goalId: number,
		plan: { type: string; criterion: string; strength: string },
	): Promise<VerificationPlan> => {
		return await apiRequestWithAuth<VerificationPlan>(
			`goals/${goalId}/verification`,
			{ method: "PUT", body: plan },
		);
	};

	const removeVerification = async (goalId: number): Promise<void> => {
		await apiRequestWithAuth(`goals/${goalId}/verification`, {
			method: "DELETE",
		});
	};

	const addEvidence = async (
		goalId: number,
		evidence: { type: string; value: string; source?: string },
	): Promise<GoalEvidence> => {
		return await apiRequestWithAuth<GoalEvidence>(
			`goals/${goalId}/evidence`,
			{ method: "POST", body: evidence },
		);
	};

	const runVerification = async (
		goalId: number,
	): Promise<VerificationResult> => {
		return await apiRequestWithAuth<VerificationResult>(
			`goals/${goalId}/verification/run`,
			{ method: "POST" },
		);
	};

	// Att godkänna ett omdöme är människans beslut - därför ligger
	// anropet här, och inte bland agentens.
	const approveResult = async (
		resultId: number,
	): Promise<VerificationResult> => {
		return await apiRequestWithAuth<VerificationResult>(
			`verification-results/${resultId}/approve`,
			{ method: "POST" },
		);
	};

	// Att avslå ett omdöme är också människans beslut.
	const rejectResult = async (
		resultId: number,
	): Promise<VerificationResult> => {
		return await apiRequestWithAuth<VerificationResult>(
			`verification-results/${resultId}/reject`,
			{ method: "POST" },
		);
	};

	// "Jag är klar" - avslutar målet med en handling.
	//
	// Svaret innehåller hur målet avslutades - "deklarerat" eller
	// "verifierat" - så att vyn kan visa det direkt.
	const completeGoal = async (
		goalId: number,
		reason?: string,
	): Promise<{ completion_kind?: string }> => {
		return await apiRequestWithAuth<{ completion_kind?: string }>(
			`goals/${goalId}/complete`,
			{
				method: "POST",
				body: reason ? { reason } : {},
			},
		);
	};

	return {
		isApiFeatureMissing,
		loadVerification,
		saveVerification,
		removeVerification,
		addEvidence,
		runVerification,
		approveResult,
		rejectResult,
		completeGoal,
	};
}
