/**
 * fzf-liknande fuzzy-matchning.
 *
 * Sökningen i kommandopaletten ska kännas som fzf: skriv några tecken i
 * valfri ordning och träffarna ska hittas ändå. En vanlig
 * substring-sökning räcker inte, och Fuse.js matchar för lätt när
 * tröskeln höjs - "mk" träffar då "Läkare". Därför görs matchningen här.
 */

/** Tecken som avslutar ett ord. */
const WORD_SEPARATORS = new Set([
	" ",
	"\t",
	"-",
	"_",
	"/",
	"\\",
	".",
	",",
	":",
	";",
	"!",
	"?",
	"(",
	")",
	"[",
	"]",
	"{",
	"}",
	'"',
	"'",
	"<",
	">",
	"|",
	"+",
	"*",
	"&",
	"^",
	"%",
	"#",
	"@",
]);

/** Grundpoängen för varje träffat tecken. */
const BONUS_MATCH = 16;
/** Bonus när träffen börjar i textens början. */
const BONUS_FIRST = 20;
/** Bonus när träffen börjar ett nytt ord. */
const BONUS_BOUNDARY = 16;
/** Bonus när träffen är en camelCase-gräns. */
const BONUS_CAMEL = 12;
/** Bonus när tecken träffas i följd. */
const BONUS_CONSECUTIVE = 12;
/** Kostnad per tecken som hoppas över mellan träffar. */
const PENALTY_GAP = 1;
/** Tak för hoppstraffen, så att en lång lucka inte straffar för mycket. */
const MAX_GAP_PENALTY = 8;

export interface FuzzyMatch {
	/** Högre poäng betyder bättre träff. */
	score: number;
	/** Index i texten som träffades, i stigande ordning. */
	positions: number[];
}

/**
 * Plockar bort diacriter och gör bokstävernas grundform.
 *
 * "mål" och "mal" ska matcha varandra, precis som i fzf.
 */
function foldChar(char: string): string {
	return char.normalize("NFD").replace(/\p{Diacritic}/gu, "");
}

/**
 * Smart case: versaler i sökordet betyder att versaler krävs i träffen.
 */
function isCaseSensitive(pattern: string): boolean {
	return /\p{Lu}/u.test(pattern);
}

/** Sant om tecknet matchar, med hänsyn till smart case. */
function charEquals(
	patternChar: string,
	textChar: string,
	caseSensitive: boolean,
): boolean {
	if (caseSensitive) return patternChar === textChar;
	return patternChar.toLowerCase() === textChar.toLowerCase();
}

/** Sant om positionen i texten är början på ett nytt ord. */
function isWordBoundary(text: string, index: number): boolean {
	if (index === 0) return true;
	return WORD_SEPARATORS.has(text[index - 1] ?? "");
}

/** Sant om positionen i texten är en camelCase-gräns ("målMiddag"). */
function isCamelBoundary(text: string, index: number): boolean {
	if (index === 0) return false;
	const previous = text[index - 1] ?? "";
	const current = text[index] ?? "";
	if (!previous || !current) return false;
	const previousIsLower = previous === previous.toLowerCase();
	const currentIsUpper = current !== current.toLowerCase();
	return previousIsLower && currentIsUpper;
}

/**
 * Matcher ett sökord mot texten med en given startposition.
 *
 * Greedy: nästa tecken tas så tidigt som möjligt. Bonusen för ordgränser
 * och följdförlopp gör att tidiga, sammanhängande träffar vinner.
 */
function matchFrom(
	pattern: string,
	foldedChars: string[],
	text: string,
	start: number,
	caseSensitive: boolean,
): FuzzyMatch | null {
	let score = 0;
	let patternIndex = 0;
	let lastIndex = -1;
	const positions: number[] = [];

	for (let index = start; index < foldedChars.length; index += 1) {
		if (patternIndex >= pattern.length) break;

		const patternChar = pattern[patternIndex] ?? "";
		const textChar = foldedChars[index] ?? "";
		if (!charEquals(patternChar, textChar, caseSensitive)) continue;

		let bonus = 0;
		if (lastIndex === index - 1) {
			bonus = BONUS_CONSECUTIVE;
		} else {
			if (index === 0) {
				bonus = BONUS_FIRST;
			} else if (isWordBoundary(text, index)) {
				bonus = BONUS_BOUNDARY;
			} else if (isCamelBoundary(text, index)) {
				bonus = BONUS_CAMEL;
			}

			const gap = index - lastIndex - 1;
			score -= Math.min(gap, MAX_GAP_PENALTY) * PENALTY_GAP;
		}

		score += BONUS_MATCH + bonus;
		positions.push(index);
		lastIndex = index;
		patternIndex += 1;
	}

	if (patternIndex < pattern.length) return null;

	return { score, positions };
}

/**
 * Matchar ett enda sökord mot texten och returnerar den bästa träffen.
 */
function matchTerm(term: string, text: string): FuzzyMatch | null {
	const pattern = foldChar(term);
	if (!pattern) return null;

	const foldedChars = text.split("").map(foldChar);
	const caseSensitive = isCaseSensitive(term);

	let best: FuzzyMatch | null = null;

	for (let start = 0; start < foldedChars.length; start += 1) {
		const candidate = matchFrom(
			pattern,
			foldedChars,
			text,
			start,
			caseSensitive,
		);
		if (candidate && (!best || candidate.score > best.score)) {
			best = candidate;
		}
	}

	return best;
}

/**
 * Matchar ett sökord mot texten.
 *
 * Flera ord i sökordet är ett AND: alla ord måste hittas någonstans i
 * texten. Poängen blir summan av ordens poäng, så ett exakt prefix ger
 * mer poäng än en träff mitt i en lång text.
 */
export function fuzzyMatch(pattern: string, text: string): FuzzyMatch | null {
	const terms = pattern.trim().split(/\s+/).filter(Boolean);
	if (!terms.length) return { score: 0, positions: [] };

	let score = 0;
	const positions: number[] = [];

	for (const term of terms) {
		const match = matchTerm(term, text);
		if (!match) return null;
		score += match.score;
		positions.push(...match.positions);
	}

	return { score, positions: [...new Set(positions)].sort((a, b) => a - b) };
}

/**
 * Poängen för en sökning, eller null om texten inte matchar.
 */
export function fuzzyScore(pattern: string, text: string): number | null {
	const match = fuzzyMatch(pattern, text);
	return match ? match.score : null;
}
