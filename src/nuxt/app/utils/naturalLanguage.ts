/**
 * Naturligspråkig tolkning av inmatning, i stil med Fantastical.
 *
 * Målet är att användaren ska kunna skriva "Lunch med Anna imorgon kl 12"
 * och få en kalenderhändelse, eller "Läs bok" och få ett mål, utan att
 * behöva fylla i formulärfält.
 *
 * Tolkningen är avsiktligt försiktig: okända ord lämnas kvar i titeln.
 */

import { extractEmoji } from "./emojis.ts";

export interface ParsedEntry {
	/** Titel efter att datum, tid och längd tagits bort. */
	title: string
	/**
	 * Emoji som fanns i texten, exempelvis "🏃". Används som ikon för
	 * händelsen och tas bort ur titeln.
	 */
	icon: string | null
	/** Starttid, om texten angav datum eller klockslag. */
	startAt: Date | null
	/** Sluttid, beräknad utifrån längd eller standardlängd. */
	endAt: Date | null
	/** Sant om händelsen ska vara heldag (datum utan klockslag). */
	allDay: boolean
	/** RRULE om texten angav upprepning. */
	recurrenceRule: string | null
	/** Den fras i texten som angav upprepningen, t.ex. "varannan dag". */
	recurrencePhrase: string | null
	/** Sant om texten innehöll ett klockslag. */
	hasTime: boolean
	/** Sant om texten innehöll ett datum. */
	hasDate: boolean
	/**
	 * Sant om datumet skrevs ut i texten. En upprepning som "varje dag"
	 * räknas inte, eftersom den bara anger mönstret.
	 */
	dateExplicit: boolean
	/**
	 * Sant om titeln innehåller ett ord som talar för att det är en
	 * händelse, exempelvis "kalas" eller "möte".
	 */
	eventHint: boolean
}

const DEFAULT_DURATION_MINUTES = 60

/** Standardtid som föreslås för en händelse utan utskrivet klockslag. */
const DEFAULT_EVENT_HOUR = 18

/**
 * Ord som starkt talar för att texten beskriver en händelse snarare än
 * ett mål. Utan detta skulle "Harry Potter Kalas" hamna som ett mål,
 * eftersom inget datum eller klockslag skrevs ut.
 */
const EVENT_KEYWORDS = [
	"kalas",
	"födelsedag",
	"jubileum",
	"möte",
	"lunch",
	"middag",
	"frukost",
	"fika",
	"träning",
	"kondition",
	"gym",
	"konsert",
	"bröllop",
	"fest",
	"afterwork",
	"aw",
	"intervju",
	"presentation",
	"föreläsning",
	"kurs",
	"workshop",
	"release",
	"träff",
	"semester",
	"resa",
	"flyg",
	"frisör",
	"tandläkare",
	"läkare",
	"optiker",
	"restaurang",
	"hundpromenad",
]

const EVENT_KEYWORD_PATTERN = new RegExp(
	`\\b(${EVENT_KEYWORDS.join("|")})\\b`,
	"i",
)

/**
 * Sant om titeln innehåller ett ord som talar för en händelse.
 */
export function looksLikeEvent(title: string): boolean {
	return EVENT_KEYWORD_PATTERN.test(title)
}

/**
 * Tidsmönster i ett tidsintervall.
 *
 * Grupperna är alltid (timme, minut) för början och sedan (timme, minut)
 * för slutet, vilket gör att alla mönter kan tolkas likadant.
 */
const TIME_ANY = String.raw`(\d{1,2})(?:[:.](\d{2}))?`
const TIME_STRICT = String.raw`(\d{1,2})[:.](\d{2})`

/**
 * Mönster för tidsintervall, exempelvis "mellan 23.00 och 07.00".
 *
 * Alla mönster har samma grupper: (start timme, start minut, slut timme,
 * slut minut). Lookbehind-skyddet hindrar att en del av ett datum eller ett
 * klockslag plockas ut som ett intervall, exempelvis "2026-03-10".
 */
const TIME_RANGE_PATTERNS: RegExp[] = [
	// "mellan 23.00 och 07.00"
	new RegExp(
		String.raw`(?<![\d./\-:])\b(?:mellan|between)\s+(?:kl\.?\s*)?${TIME_ANY}\s+(?:och|and|till|to)\s+(?:kl\.?\s*)?${TIME_ANY}`,
		"i",
	),
	// "från 23.00 till 07.00"
	new RegExp(
		String.raw`(?<![\d./\-:])\b(?:fr[åa]n|from)\s+(?:kl\.?\s*)?${TIME_ANY}\s+(?:till|to)\s+(?:kl\.?\s*)?${TIME_ANY}`,
		"i",
	),
	// "23.00-07.00"
	new RegExp(
		String.raw`(?<![\d./\-:])(?:kl\.?\s*)?${TIME_STRICT}\s*(?:-|–|—|till|to)\s*(?:kl\.?\s*)?${TIME_ANY}`,
		"i",
	),
	// "kl 23-07"
	new RegExp(
		String.raw`(?<![\d./\-:])\bkl\.?\s*(\d{1,2})(?:[:.](\d{2}))?\s*(?:-|–|—|till|to)\s*(?:kl\.?\s*)?(\d{1,2})(?:[:.](\d{2}))?`,
		"i",
	),
	// "14-15"
	new RegExp(
		String.raw`(?<![\d./\-:])(\d{1,2})(?:[:.](\d{2}))?\s*(?:-|–|—)\s*(\d{1,2})(?:[:.](\d{2}))?(?![\d])`,
		"i",
	),
]

interface Weekday {
	names: string[]
	index: number
}

const WEEKDAYS: Weekday[] = [
	{ names: ["söndag", "sön", "sundag", "sunday", "sun"], index: 0 },
	{ names: ["måndag", "mån", "mandag", "monday", "mon"], index: 1 },
	{ names: ["tisdag", "tis", "tuesday", "tue", "tues"], index: 2 },
	{ names: ["onsdag", "ons", "wednesday", "wed"], index: 3 },
	{ names: ["torsdag", "tors", "thursday", "thu", "thurs"], index: 4 },
	{ names: ["fredag", "fre", "friday"], index: 5 },
	{ names: ["lördag", "lör", "lordag", "saturday", "sat"], index: 6 },
]

const MONTHS: Array<{ names: string[]; index: number }> = [
	{ names: ["januari", "january", "jan"], index: 0 },
	{ names: ["februari", "february", "feb"], index: 1 },
	{ names: ["mars", "march", "mar"], index: 2 },
	{ names: ["april", "apr"], index: 3 },
	{ names: ["maj", "may"], index: 4 },
	{ names: ["juni", "june", "jun"], index: 5 },
	{ names: ["juli", "july", "jul"], index: 6 },
	{ names: ["augusti", "august", "aug"], index: 7 },
	{ names: ["september", "sep", "sept"], index: 8 },
	{ names: ["oktober", "october", "oct"], index: 9 },
	{ names: ["november", "nov"], index: 10 },
	{ names: ["december", "dec"], index: 11 },
]

/** Ord som indikerar ett klockslag utan att själva tas bort ur titeln. */
const TIME_HINTS: Array<{ pattern: RegExp; hour: number; minute: number }> = [
	{ pattern: /\blunch\b/i, hour: 12, minute: 0 },
	{ pattern: /\bmiddag\b/i, hour: 18, minute: 0 },
	{ pattern: /\bfrukost\b/i, hour: 8, minute: 0 },
	{ pattern: /\bkvällsmat\b/i, hour: 19, minute: 0 },
]

/** Enkel arbetsyta där matchade delar kan tas bort ur texten. */
class TextBuffer {
	private value: string

	constructor(value: string) {
		this.value = value
	}

	get text(): string {
		return this.value
	}

	/** Tar bort första matchen och returnerar matchen. */
	take(pattern: RegExp): RegExpExecArray | null {
		const match = pattern.exec(this.value)
		if (!match || match.index === undefined) return null

		this.value =
			this.value.slice(0, match.index) +
			" " +
			this.value.slice(match.index + match[0].length)
		return match
	}

	/** Tar bort första matchen tillsammans med ett föregående "i ". */
	takeWithPrecedingPreposition(pattern: RegExp): RegExpExecArray | null {
		const match = pattern.exec(this.value)
		if (!match || match.index === undefined) return null

		const before = this.value.slice(0, match.index)
		const trimmed = before.replace(/\bi\s*$/i, "")
		this.value =
			trimmed +
			" " +
			this.value.slice(match.index + match[0].length)
		return match
	}

	/** Tar bort ett antal tecken på given position. */
	removeAt(index: number, length: number) {
		this.value =
			this.value.slice(0, index) + " " + this.value.slice(index + length)
	}
}

function escapeRegExp(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
}

function buildAlternation(names: string[]): string {
	const unique = Array.from(new Set(names)).sort(
		(a, b) => b.length - a.length,
	)
	// Gruppen är nödvändig: utan den binder alternativen till resten av
	// mönstret, exempelvis "(måndag|mån)(?:ar|er)".
	return `(?:${unique.map(escapeRegExp).join("|")})`
}

const WEEKDAY_ALTERNATION = buildAlternation(
	WEEKDAYS.flatMap((weekday) => weekday.names),
)

const MONTH_ALTERNATION = buildAlternation(
	MONTHS.flatMap((month) => month.names),
)

/**
 * Mönster för datumintervall, exempelvis "mellan 12 och 13 mars".
 *
 * Grupperna är (första dag, sista dag, månad).
 */
const DATE_RANGE_PATTERNS: RegExp[] = [
	new RegExp(
		String.raw`(?<![\d./\-:])\b(?:mellan|between)\s+(\d{1,2})\s+(?:och|and)\s+(\d{1,2})\s+(${MONTH_ALTERNATION})\b`,
		"i",
	),
	new RegExp(
		String.raw`(?<![\d./\-:])\b(\d{1,2})\s*(?:-|–|—)\s*(\d{1,2})\s+(${MONTH_ALTERNATION})\b`,
		"i",
	),
	new RegExp(
		String.raw`(?<![\d./\-:])\b(\d{1,2})\s+(?:till|to)\s+(\d{1,2})\s+(${MONTH_ALTERNATION})\b`,
		"i",
	),
]

function findWeekday(text: string): { index: number; matched: string } | null {
	const pattern = new RegExp(`\\b${WEEKDAY_ALTERNATION}\\b`, "i")
	const match = pattern.exec(text)
	if (!match) return null

	const name = match[0].toLowerCase()
	for (const weekday of WEEKDAYS) {
		if (weekday.names.includes(name)) {
			return { index: weekday.index, matched: match[0] }
		}
	}
	return null
}

function findMonth(text: string): number | null {
	const pattern = new RegExp(`\\b${MONTH_ALTERNATION}\\b`, "i")
	const match = pattern.exec(text)
	if (!match) return null

	const name = match[0].toLowerCase()
	for (const month of MONTHS) {
		if (month.names.includes(name)) return month.index
	}
	return null
}

function addDays(date: Date, days: number): Date {
	const result = new Date(date.getTime())
	result.setDate(result.getDate() + days)
	return result
}

function atTime(date: Date, hour: number, minute: number): Date {
	const result = new Date(date.getTime())
	result.setHours(hour, minute, 0, 0)
	return result
}

function startOfDay(date: Date): Date {
	const result = new Date(date.getTime())
	result.setHours(0, 0, 0, 0)
	return result
}

interface ClockTime {
	hour: number
	minute: number
}

function isValidClockTime(time: ClockTime): boolean {
	return (
		time.hour >= 0 &&
		time.hour <= 23 &&
		time.minute >= 0 &&
		time.minute <= 59
	)
}

/**
 * Ett rent timintervall, exempelvis "14-15", är svårt att skilja från ett
 * antal ("3-4 böcker"). Därför accepteras bara timmar som ligger inom
 * rimliga öppettider.
 */
function isPlausibleHourRange(start: ClockTime, end: ClockTime): boolean {
	const first = Math.min(start.hour, end.hour)
	const last = Math.max(start.hour, end.hour)
	return first >= 6 && last <= 23
}

/**
 * Räknar ut sluttiden för ett tidsintervall.
 *
 * Ett slut som är före eller lika med starten betyder att händelsen
 * passerar midnatt, exempelvis "mellan 23.00 och 07.00".
 */
function endTimeFor(start: Date, end: ClockTime): Date {
	const result = atTime(start, end.hour, end.minute)
	if (result.getTime() <= start.getTime()) {
		return addDays(result, 1)
	}
	return result
}

/** Nästa förekomst av en veckodag, där idag räknas som möjlig träff. */
function nextWeekday(now: Date, weekdayIndex: number): Date {
	const days = (weekdayIndex - now.getDay() + 7) % 7
	return addDays(startOfDay(now), days)
}

/**
 * Tolkar naturlig text.
 *
 * @param input användarens text
 * @param now referenstidpunkt, används för "idag", "imorgon" och veckodagar
 */
export function parseNaturalLanguage(
	input: string,
	now: Date = new Date(),
): ParsedEntry {
	// Emoji plockas ut först, så att de inte stör datum- och tidstolkningen.
	const { emoji, text: withoutEmoji } = extractEmoji(input.trim())
	const buffer = new TextBuffer(withoutEmoji)
	const original = buffer.text

	let recurrenceRule: string | null = null
	let recurrencePhrase: string | null = null
	let weekdayForRecurrence: number | null = null

	// Upprepning plockas först ut så att veckodagar inom upprepningen inte
	// misstolkas som ett enstaka datum.
	const pluralWeekday = new RegExp(
		`\\b${WEEKDAY_ALTERNATION}(?:ar|er)\\b`,
		"i",
	)
	const pluralMatch = pluralWeekday.exec(buffer.text)
	if (pluralMatch) {
		const weekday = findWeekday(pluralMatch[0].replace(/(ar|er)$/i, ""))
		if (weekday) {
			recurrenceRule = `FREQ=WEEKLY;BYDAY=${toRRuleDay(weekday.index)}`
			recurrencePhrase = pluralMatch[0]
			weekdayForRecurrence = weekday.index
			buffer.take(pluralWeekday)
		}
	}

	if (!recurrenceRule) {
		const everyWeekday = new RegExp(
			`\\b(?:varje|every)\\s+${WEEKDAY_ALTERNATION}(?:ar|er)?\\b`,
			"i",
		)
		const match = everyWeekday.exec(buffer.text)
		if (match) {
			const weekday = findWeekday(match[0])
			if (weekday) {
				recurrenceRule = `FREQ=WEEKLY;BYDAY=${toRRuleDay(weekday.index)}`
				recurrencePhrase = match[0]
				weekdayForRecurrence = weekday.index
				buffer.take(everyWeekday)
			}
		}
	}

	if (!recurrenceRule) {
		const everyOtherDay = /\b(varannan\s+dag|var\s+annan\s+dag|every\s+other\s+day|every\s+2nd\s+day)\b/i.exec(
			buffer.text,
		)
		if (everyOtherDay) {
			recurrenceRule = "FREQ=DAILY;INTERVAL=2"
			recurrencePhrase = everyOtherDay[0]
			buffer.take(
				/\b(varannan\s+dag|var\s+annan\s+dag|every\s+other\s+day|every\s+2nd\s+day)\b/i,
			)
		}
	}

	if (!recurrenceRule) {
		const daily = /\b(varje\s+dag|dagligen|daily)\b/i.exec(buffer.text)
		if (daily) {
			recurrenceRule = "FREQ=DAILY"
			recurrencePhrase = daily[0]
			buffer.take(/\b(varje\s+dag|dagligen|daily)\b/i)
		}
	}

	if (!recurrenceRule) {
		const everyOtherWeek = /\b(varannan\s+vecka|every\s+other\s+week|biweekly)\b/i.exec(
			buffer.text,
		)
		if (everyOtherWeek) {
			recurrenceRule = "FREQ=WEEKLY;INTERVAL=2"
			recurrencePhrase = everyOtherWeek[0]
			buffer.take(/\b(varannan\s+vecka|every\s+other\s+week|biweekly)\b/i)
		}
	}

	if (!recurrenceRule) {
		const weekly = /\b(varje\s+vecka|veckovis|weekly)\b/i.exec(buffer.text)
		if (weekly) {
			recurrenceRule = "FREQ=WEEKLY"
			recurrencePhrase = weekly[0]
			buffer.take(/\b(varje\s+vecka|veckovis|weekly)\b/i)
		}
	}

	if (!recurrenceRule) {
		const monthly = /\b(varje\s+m[åa]nad|m[åa]nadsvis|monthly)\b/i.exec(
			buffer.text,
		)
		if (monthly) {
			recurrenceRule = "FREQ=MONTHLY"
			recurrencePhrase = monthly[0]
			buffer.take(
				/\b(varje\s+m[åa]nad|m[åa]nadsvis|monthly)\b/i,
			)
		}
	}

	if (!recurrenceRule) {
		const yearly = /\b(varje\s+[åa]r|[åa]rligen|yearly)\b/i.exec(buffer.text)
		if (yearly) {
			recurrenceRule = "FREQ=YEARLY"
			recurrencePhrase = yearly[0]
			buffer.take(/\b(varje\s+[åa]r|[åa]rligen|yearly)\b/i)
		}
	}

	// Längd
	let durationMinutes: number | null = null

	const hourDuration = buffer.takeWithPrecedingPreposition(
		/(\d+(?:[.,]\d+)?)\s*(?:timme|timmar|tim|hour|hours|hr|h)\b/i,
	)
	if (hourDuration) {
		durationMinutes = Math.round(parseFloat(hourDuration[1].replace(",", ".")) * 60)
	} else {
		const minuteDuration = buffer.takeWithPrecedingPreposition(
			/(\d+)\s*(?:minut|minuter|min|minutes?)\b/i,
		)
		if (minuteDuration) {
			durationMinutes = parseInt(minuteDuration[1], 10)
		} else if (/\b(?:en|ett|1)\s*halvtimme\b/i.test(buffer.text)) {
			durationMinutes = 30
			buffer.take(/\b(?:en|ett|1)\s*halvtimme\b/i)
		} else if (/\b(?:en|ett|1)\s*kvart\b/i.test(buffer.text)) {
			durationMinutes = 15
			buffer.take(/\b(?:en|ett|1)\s*kvart\b/i)
		}
	}

	// Datumintervall, exempelvis "mellan 12 och 13 mars". Detta hanteras
	// före tidsintervall så att dagar inte misstolkas som klockslag.
	let rangeStartDate: Date | null = null
	let rangeEndDate: Date | null = null

	for (const pattern of DATE_RANGE_PATTERNS) {
		const match = pattern.exec(buffer.text)
		if (!match || match.index === undefined) continue

		const month = findMonth(match[3])
		if (month === null) continue

		const firstDay = parseInt(match[1], 10)
		const lastDay = parseInt(match[2], 10)
		const year = now.getFullYear()
		let first = new Date(year, month, firstDay)
		let last = new Date(year, month, lastDay)
		if (first.getTime() < startOfDay(now).getTime()) {
			first = new Date(year + 1, month, firstDay)
			last = new Date(year + 1, month, lastDay)
		}
		rangeStartDate = startOfDay(first)
		rangeEndDate = addDays(startOfDay(last), 1)
		buffer.removeAt(match.index, match[0].length)
		break
	}

	// Tidsintervall, exempelvis "mellan 23.00 och 07.00". Detta plockas
	// ut före enskilda klockslag så att båda tiderna hittas.
	let rangeStart: ClockTime | null = null
	let rangeEnd: ClockTime | null = null

	for (const pattern of TIME_RANGE_PATTERNS) {
		const match = pattern.exec(buffer.text)
		if (!match || match.index === undefined) continue

		// Ett intervall som följs av ett månadsnamn är sannolikt ett
		// datumintervall, exempelvis "mellan 12 och 13 mars".
		const rest = buffer.text.slice(match.index + match[0].length)
		if (
			new RegExp(String.raw`^\s*(?:${MONTH_ALTERNATION})\b`, "i").test(rest)
		) {
			continue
		}

		const start: ClockTime = {
			hour: parseInt(match[1], 10),
			minute: match[2] ? parseInt(match[2], 10) : 0,
		}
		const end: ClockTime = {
			hour: parseInt(match[3], 10),
			minute: match[4] ? parseInt(match[4], 10) : 0,
		}
		if (!isValidClockTime(start) || !isValidClockTime(end)) continue

		// Ett rent timintervall utan minuter är svårt att skilja från ett
		// antal, exempelvis "3-4 böcker". Därför krävs det att timmarna
		// ligger inom rimliga öppettider.
		if (!match[2] && !match[4] && !isPlausibleHourRange(start, end)) {
			continue
		}

		rangeStart = start
		rangeEnd = end
		buffer.removeAt(match.index, match[0].length)
		break
	}

	// Klockslag
	let hour: number | null = null
	let minute = 0

	if (rangeStart) {
		hour = rangeStart.hour
		minute = rangeStart.minute
	} else {
		const swedishTime = buffer.take(
			/\bkl\.?\s*(\d{1,2})(?:[:.](\d{2}))?\b/i,
		)
		if (swedishTime) {
			hour = parseInt(swedishTime[1], 10)
			minute = swedishTime[2] ? parseInt(swedishTime[2], 10) : 0
		} else {
			const clockTime = buffer.take(/\b(\d{1,2})[:.](\d{2})\b/i)
			if (clockTime) {
				hour = parseInt(clockTime[1], 10)
				minute = parseInt(clockTime[2], 10)
			} else {
				const ampm = buffer.take(/\b(\d{1,2})\s*(am|pm)\b/i)
				if (ampm) {
					const value = parseInt(ampm[1], 10)
					const isPm = ampm[2].toLowerCase() === "pm"
					hour = isPm && value < 12 ? value + 12 : value
				}
			}
		}
	}

	// Ord som "lunch" ger ett klockslag men behålls i titeln.
	if (hour === null) {
		for (const hint of TIME_HINTS) {
			if (hint.pattern.test(original)) {
				hour = hint.hour
				minute = hint.minute
				break
			}
		}
	}

	// Datum
	let date: Date | null = null
	let dateExplicit = false

	const isoDate = buffer.take(/\b(\d{4})-(\d{1,2})-(\d{1,2})\b/)
	if (isoDate) {
		date = new Date(
			Number(isoDate[1]),
			Number(isoDate[2]) - 1,
			Number(isoDate[3]),
		)
	}

	if (!date) {
		const numericDate = buffer.take(
			/\b(\d{1,2})[/.](\d{1,2})(?:[/.](\d{2,4}))?\b/,
		)
		if (numericDate) {
			const day = parseInt(numericDate[1], 10)
			const month = parseInt(numericDate[2], 10) - 1
			let year = now.getFullYear()
			if (numericDate[3]) {
				const raw = parseInt(numericDate[3], 10)
				year = raw < 100 ? 2000 + raw : raw
			}
			date = new Date(year, month, day)
			// Utan år väljs det närmaste kommande datumet.
			if (!numericDate[3] && date.getTime() < startOfDay(now).getTime()) {
				date = new Date(year + 1, month, day)
			}
		}
	}

	if (!date) {
		const dayMonth = buffer.take(
			new RegExp(`\\b(\\d{1,2})\\s+${MONTH_ALTERNATION}\\b`, "i"),
		)
		if (dayMonth) {
			const month = findMonth(dayMonth[0])
			if (month !== null) {
				const day = parseInt(dayMonth[1], 10)
				date = new Date(now.getFullYear(), month, day)
				if (date.getTime() < startOfDay(now).getTime()) {
					date = new Date(now.getFullYear() + 1, month, day)
				}
			}
		}
	}

	if (!date) {
		const monthDay = buffer.take(
			new RegExp(`\\b${MONTH_ALTERNATION}\\s+(\\d{1,2})\\b`, "i"),
		)
		if (monthDay) {
			const month = findMonth(monthDay[0])
			if (month !== null) {
				const day = parseInt(monthDay[1], 10)
				date = new Date(now.getFullYear(), month, day)
				if (date.getTime() < startOfDay(now).getTime()) {
					date = new Date(now.getFullYear() + 1, month, day)
				}
			}
		}
	}

	if (!date) {
		const tomorrow = buffer.take(
			/\b(i\s+morgon|imorgon|tomorrow)\b/i,
		)
		if (tomorrow) {
			date = addDays(startOfDay(now), 1)
		}
	}

	if (!date) {
		const dayAfter = buffer.take(/\b(övermorgon|day\s+after\s+tomorrow)\b/i)
		if (dayAfter) {
			date = addDays(startOfDay(now), 2)
		}
	}

	if (!date) {
		const today = buffer.take(/\b(idag|today)\b/i)
		if (today) {
			date = startOfDay(now)
		}
	}

	if (!date) {
		const relative = buffer.takeWithPrecedingPreposition(
			/\bom\s+(\d+)\s*(dag|dagar|veckor?|m[åa]nader?|manader?|[åa]r|days?|weeks?|months?|years?)\b/i,
		)
		if (relative) {
			const amount = parseInt(relative[1], 10)
			const unit = relative[2].toLowerCase()
			if (/veck|week/.test(unit)) {
				date = addDays(startOfDay(now), amount * 7)
			} else if (/m[åa]nad|manad|month/.test(unit)) {
				date = new Date(now.getFullYear(), now.getMonth() + amount, now.getDate())
			} else if (/[åa]r|year/.test(unit)) {
				date = new Date(now.getFullYear() + amount, now.getMonth(), now.getDate())
			} else {
				date = addDays(startOfDay(now), amount)
			}
		}
	}

	if (date) {
		dateExplicit = true
		// "start" är en markör i exempel som "start mån 30 nov",
		// inte en del av händelsens titel.
		buffer.take(/\b(?:med\s+)?start\b/i)
		const dateWeekday = findWeekday(buffer.text)
		if (dateWeekday) {
			buffer.take(new RegExp(`\\b${escapeRegExp(dateWeekday.matched)}\\b`, "i"))
		}
	}

	if (!date && recurrenceRule) {
		// En upprepning utan uttryckligt datum startar från närmsta
		// förekomst av den angivna veckan.
		date =
			weekdayForRecurrence !== null
				? nextWeekday(now, weekdayForRecurrence)
				: startOfDay(now)
	}

	if (!date && weekdayForRecurrence === null) {
		const weekday = findWeekday(buffer.text)
		if (weekday) {
			date = nextWeekday(now, weekday.index)
			buffer.take(new RegExp(`\\b${escapeRegExp(weekday.matched)}\\b`, "i"))
		}
	}

	if (rangeStart && rangeEnd) {
		// Tidsintervallet hanterar redan midnatt när sluttiden är lika med
		// eller tidigare än starttiden. Ta därför bort den redundanta
		// förklaringen i exempel som "17.00 till 17.00 dagen efter".
		buffer.take(/\b(?:dagen\s+efter|nästa\s+dag)\b/i)
	}

	const title = buffer.text.replace(/\s+/g, " ").trim()
	const hasTime = hour !== null
	const eventHint = looksLikeEvent(title)
	let hasDate = date !== null

	let startAt: Date | null = null
	let endAt: Date | null = null
	let allDay = false

	if (rangeStartDate && rangeEndDate) {
		allDay = true
		startAt = rangeStartDate
		endAt = rangeEndDate
	} else if (date && hasTime) {
		startAt = atTime(date, hour as number, minute)
		endAt = rangeEnd
			? endTimeFor(startAt, rangeEnd)
			: new Date(
					startAt.getTime() +
						(durationMinutes ?? DEFAULT_DURATION_MINUTES) * 60 * 1000,
				)
	} else if (date && !hasTime) {
		allDay = true
		startAt = startOfDay(date)
		endAt = addDays(startAt, 1)
	} else if (!date && hasTime) {
		startAt = atTime(now, hour as number, minute)
		if (rangeEnd) {
			// Ett intervall ligger kvar på dagens datum så länge det
			// inte redan är slut.
			endAt = endTimeFor(startAt, rangeEnd)
			if (endAt.getTime() <= now.getTime()) {
				startAt = atTime(addDays(now, 1), hour as number, minute)
				endAt = endTimeFor(startAt, rangeEnd)
			}
		} else {
			if (startAt.getTime() <= now.getTime()) {
				startAt = atTime(addDays(now, 1), hour as number, minute)
			}
			endAt = new Date(
				startAt.getTime() +
					(durationMinutes ?? DEFAULT_DURATION_MINUTES) * 60 * 1000,
			)
		}
	}

	// Titeln talar för en händelse men inget datum eller klockslag skrevs
	// ut. Då föreslår vi kvällen idag (eller imorgon om den redan passerat)
	// i stället för att tappa texten på ett mål.
	if (!startAt && eventHint) {
		const evening = atTime(now, DEFAULT_EVENT_HOUR, 0)
		startAt = evening.getTime() > now.getTime() ? evening : atTime(addDays(now, 1), DEFAULT_EVENT_HOUR, 0)
		endAt = new Date(
			startAt.getTime() +
				(durationMinutes ?? DEFAULT_DURATION_MINUTES) * 60 * 1000,
		)
		hasDate = true
	}

	return {
		title,
		icon: emoji,
		startAt,
		endAt,
		allDay,
		recurrenceRule,
		recurrencePhrase,
		hasTime,
		hasDate,
		dateExplicit,
		eventHint,
	}
}

function toRRuleDay(index: number): string {
	return ["SU", "MO", "TU", "WE", "TH", "FR", "SA"][index]
}

/**
 * Avgör om en tolkad inmatning ska bli ett mål eller en händelse.
 *
 * I automatiskt läge blir det en händelse om texten innehåller datum,
 * klockslag eller upprepning. Även ett ord som tydligt talar för en
 * händelse, som "kalas" eller "möte", gör att det blir en händelse.
 */
export function resolveQuickAddMode(
	entry: ParsedEntry,
	mode: "auto" | "event" | "goal" = "auto",
): "event" | "goal" {
	if (mode !== "auto") return mode;
	return entry.hasTime ||
		entry.dateExplicit ||
		entry.recurrenceRule ||
		entry.eventHint
		? "event"
		: "goal";
}

/**
 * Tar fram titeln på ett mål från en tolkad inmatning.
 *
 * Ett mål behåller hela texten - mål har inget datumfält, så att tappa
 * bort ord bara skulle förlora information. En upprepningsfras är ett
 * undantag: den blir regel i stället för text och ska inte stå kvar i
 * titeln.
 */
export function goalTitle(text: string, entry: ParsedEntry): string {
	const trimmed = text.trim();
	const phrase = entry.recurrencePhrase;
	if (!phrase) return trimmed;

	return trimmed.replace(phrase, "").replace(/\s{2,}/g, " ").trim();
}

/**
 * Formaterar en tolkad inmatning som en kort förklaring, exempelvis
 * "ons 12 dec 12:00-13:00" eller "heldag 12 dec".
 */
export function describeParsedEntry(entry: ParsedEntry): string {
	if (!entry.startAt) return "Inget datum hittades"

	const date = entry.startAt
	if (entry.allDay) {
		return `Heldag ${new Intl.DateTimeFormat("sv-SE", {
			day: "numeric",
			month: "short",
		}).format(date)}`
	}

	const time = new Intl.DateTimeFormat("sv-SE", {
		hour: "2-digit",
		minute: "2-digit",
	}).format(date)

	const end = entry.endAt
		? new Intl.DateTimeFormat("sv-SE", {
				hour: "2-digit",
				minute: "2-digit",
			}).format(entry.endAt)
		: null

	const day = new Intl.DateTimeFormat("sv-SE", {
		weekday: "short",
		day: "numeric",
		month: "short",
	}).format(date)

	if (!end) return `${day} ${time}`

	// Ett intervall som passerar midnatt ska visa att sluttiden ligger
	// på nästa dag, exempelvis "23:00-07:00 (+1 dag)".
	const spansMidnight =
		entry.endAt !== null &&
		new Date(entry.endAt.getFullYear(), entry.endAt.getMonth(), entry.endAt.getDate()).getTime() !==
			new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime()

	return spansMidnight
		? `${day} ${time}-${end} (+1 dag)`
		: `${day} ${time}-${end}`
}
