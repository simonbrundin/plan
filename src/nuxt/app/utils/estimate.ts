// Uppskattad tidsåtgång lagras i minuter (hela tal).
//
// Användaren kan skriva "90", "1h", "1h30", "1,5 h" eller "45m" - alla
// tolkas till antal minuter.

const PLAIN = /^(\d+(?:[.,]\d+)?)$/;
const HOURS = /^(\d+(?:[.,]\d+)?)\s*(?:h|hr|hour|hours|tim|timmar)$/;
const MINUTES = /^(\d+(?:[.,]\d+)?)\s*(?:m|min|minut|minuter)$/;
const MIXED = /^(\d+(?:[.,]\d+)?)\s*h\s*(\d+(?:[.,]\d+)?)?\s*(?:m|min|minut|minuter)?$/;

function toNumber(value: string): number {
	return Number(value.replace(",", "."));
}

/**
 * Tolkar en inmatning som antal minuter.
 *
 * Returnerar null om värdet inte kan tolkas, är negativt eller är tomt.
 */
export function parseEstimateMinutes(
	input: string | number | null | undefined,
): number | null {
	if (input === null || input === undefined) return null;
	if (typeof input === "number") {
		if (!Number.isFinite(input) || input < 0) return null;
		return Math.round(input);
	}

	const text = input.trim().toLowerCase();
	if (!text) return null;

	const mixed = MIXED.exec(text);
	if (mixed) {
		const hours = toNumber(mixed[1]);
		const rest = mixed[2] ? toNumber(mixed[2]) : 0;
		return Math.round(hours * 60 + rest);
	}

	const hours = HOURS.exec(text);
	if (hours) {
		return Math.round(toNumber(hours[1]) * 60);
	}

	const minutes = MINUTES.exec(text);
	if (minutes) {
		return Math.round(toNumber(minutes[1]));
	}

	const plain = PLAIN.exec(text);
	if (plain) {
		return Math.round(toNumber(plain[1]));
	}

	return null;
}

/**
 * Formaterar antal minuter på svenska, t.ex. "1 h 30 min" eller "45 min".
 *
 * Returnerar tom sträng om det inte finns någon uppskattning.
 */
export function formatEstimate(minutes: number | null | undefined): string {
	if (minutes === null || minutes === undefined) return "";
	if (!Number.isFinite(minutes) || minutes <= 0) return "";

	const hours = Math.floor(minutes / 60);
	const rest = Math.round(minutes % 60);
	if (hours === 0) return `${rest} min`;
	if (rest === 0) return `${hours} h`;
	return `${hours} h ${rest} min`;
}
