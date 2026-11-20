/**
 * Matchning av leader-tangentsekvenser.
 *
 * Leader-läget bygger upp en sekvens tangent för tangent ("g" + "p").
 * Här avgörs om sekvensen är ett kommando, om fler tangenter väntas
 * eller om sekvensen ska avbrytas.
 */

export type LeaderAction =
	| { type: "run"; keys: string }
	| { type: "pending"; keys: string }
	| { type: "reset" };

/**
 * @param commands Alla tillgängliga tangentsekvenser, t.ex. ["gi", "gg", "gp"]
 * @param pending Sekvensen som hittills är tryckt
 * @param key Nästa tangent
 */
export function matchLeaderKey(
	commands: string[],
	pending: string,
	key: string,
): LeaderAction {
	const candidate = pending + key;

	if (commands.includes(candidate)) {
		return { type: "run", keys: candidate };
	}

	if (commands.some((command) => command.startsWith(candidate))) {
		return { type: "pending", keys: candidate };
	}

	return { type: "reset" };
}
