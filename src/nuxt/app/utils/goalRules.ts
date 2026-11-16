/**
 * Regler för eviga mål - mål som pågår hela livet (t.ex. "kroppen") och
 * därför aldrig kan slutföras.
 */

/**
 * Ett mål med evigt-flaggan går inte att slutföra. Övriga mål gör det.
 *
 * Används både i gränssnittet (där knappen ska vara en no-op) och som
 * spegel av regeln i API:t, som avvisar försök att sätta "finished" på ett
 * evigt mål.
 */
export function canBeFinished(goal: {
	eternal?: boolean | null;
}): boolean {
	return !goal.eternal;
}
