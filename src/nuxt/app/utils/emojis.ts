/**
 * Emoji för kalenderhändelser.
 *
 * Katalogen är avsiktligt liten och handplockad. Den ska täcka de vanligaste
 * kategorierna i en kalender - arbete, hälsa, resor, socialt - utan att
 * bli en fullständig emoji-databas som bara är laddad men sällan använd.
 *
 * Sökordet är på svenska, eftersom det är vad användaren skriver.
 */

export interface EmojiEntry {
	/** Emoji-tecknet. */
	emoji: string
	/** Sökord på svenska,|matchar mot fritext. */
	keywords: string[]
}

export const EMOJI_CATALOGUE: EmojiEntry[] = [
	// Arbete och möten
	{ emoji: "💼", keywords: ["jobb", "arbete", "möte", "office", "kontor"] },
	{ emoji: "📅", keywords: ["möte", "kalender", "plan", "agenda"] },
	{ emoji: "📊", keywords: ["rapport", "statistik", "presentation"] },
	{ emoji: "🗓️", keywords: ["plan", "kalender", "schema"] },
	{ emoji: "✍️", keywords: ["skriva", "anteckning", "signera"] },
	{ emoji: "💻", keywords: ["kod", "dator", "utveckling", "programmering"] },
	{ emoji: "🔧", keywords: ["verktyg", "fixa", "reparation"] },
	{ emoji: "🏗️", keywords: ["bygg", "projekt", "arbete"] },

	// Hälsa och träning
	{ emoji: "🏃", keywords: ["löpning", "träning", "jogging", "run"] },
	{ emoji: "🏋️", keywords: ["gym", "styrka", "träning"] },
	{ emoji: "🚴", keywords: ["cykel", "cycling", "träning"] },
	{ emoji: "⚽", keywords: ["fotboll", "idrott", "match"] },
	{ emoji: "🏊", keywords: ["sim", "bad", "pool"] },
	{ emoji: "🧘", keywords: ["yoga", "meditation", "andning"] },
	{ emoji: "💊", keywords: ["medicin", "tablett", "hälsa"] },
	{ emoji: "🩺", keywords: ["läkare", "vård", "hälsa", "besök"] },
	{ emoji: "🦷", keywords: ["tandläkare", "tänder", "besök"] },
	{ emoji: "💇", keywords: ["frisör", "hår", "salong"] },
	{ emoji: "🛏️", keywords: ["sömn", "sova", "vilа", "vila"] },
	{ emoji: "😴", keywords: ["sömn", "sova", "trött"] },

	// Mat
	{ emoji: "☕", keywords: ["kaffe", "fika", "espresso"] },
	{ emoji: "🍎", keywords: ["frukt", "äpple", "snack"] },
	{ emoji: "🥗", keywords: ["sallad", "middag", "hälsosamt"] },
	{ emoji: "🍕", keywords: ["pizza", "middag"] },
	{ emoji: "🍔", keywords: ["hamburger", "middag", "lunch"] },
	{ emoji: "🌮", keywords: ["taco", "middag"] },
	{ emoji: "🍣", keywords: ["sushi", "middag"] },
	{ emoji: "🍜", keywords: ["nudlar", "middag"] },
	{ emoji: "🥐", keywords: ["frukost", "bulle", "bageri"] },
	{ emoji: "🍰", keywords: ["tårt", "fyllig", "kaka"] },
	{ emoji: "🍺", keywords: ["öl", "pubs"] },
	{ emoji: "🍷", keywords: ["vin", "dryck"] },

	// Resor och plats
	{ emoji: "✈️", keywords: ["flyg", "resa", "flygplan"] },
	{ emoji: "🚗", keywords: ["bil", "resa", "kör"] },
	{ emoji: "🚆", keywords: ["tåg", "resa"] },
	{ emoji: "🚌", keywords: ["buss", "resa"] },
	{ emoji: "🏨", keywords: ["hotell", "övernattning"] },
	{ emoji: "🗺️", keywords: ["resa", "karta", "viktigt"] },
	{ emoji: "🏖️", keywords: ["semester", "strand", "sol"] },
	{ emoji: "⛰️", keywords: ["berg", "vandring", "friluftsliv"] },
	{ emoji: "🏕️", keywords: ["tält", "camping", "friluftsliv"] },

	// Hem och sysslor
	{ emoji: "🏠", keywords: ["hem", "hus", "villa"] },
	{ emoji: "🔑", keywords: ["nyckel", "lås", "hem"] },
	{ emoji: "🧹", keywords: ["städning", "hushåll"] },
	{ emoji: "🧺", keywords: ["tvätt", "kläder"] },
	{ emoji: "🛠️", keywords: ["bygga", "verktyg"] },
	{ emoji: "🪴", keywords: ["växt", "blomma", "trädgård"] },
	{ emoji: "🐕", keywords: ["hund", "hundpromenad", "husdjur"] },
	{ emoji: "🐈", keywords: ["katt", "husdjur"] },

	// Lärande och kreativt
	{ emoji: "📚", keywords: ["bok", "läsning", "studie"] },
	{ emoji: "🎓", keywords: ["examen", "skola", "studie"] },
	{ emoji: "✏️", keywords: ["skriva", "anteckning"] },
	{ emoji: "🎨", keywords: ["målning", "konst", "skapande"] },
	{ emoji: "🎸", keywords: ["gitarr", "musik", "spela"] },
	{ emoji: "🎧", keywords: ["musik", "hörlurar", "podcast"] },
	{ emoji: "🎬", keywords: ["film", "bio", "video"] },
	{ emoji: "📷", keywords: ["foto", "kamera"] },
	{ emoji: "🎮", keywords: ["spel", "datorspel"] },
	{ emoji: "🧠", keywords: ["hjärna", "tänkande", "idé"] },

	// Socialt och glädje
	{ emoji: "🎉", keywords: ["fest", "jubileum", "glädje", "party"] },
	{ emoji: "🎂", keywords: ["födelsdag", "tårt", "fest"] },
	{ emoji: "🎁", keywords: ["present", "gåva"] },
	{ emoji: "💝", keywords: ["kärlek", "present"] },
	{ emoji: "🥂", keywords: ["skål", "fest"] },
	{ emoji: "🎵", keywords: ["musik", "konsert"] },
	{ emoji: "🎭", keywords: ["teater", "föreställning"] },
	{ emoji: "👥", keywords: ["möte", "grupp", "team"] },
	{ emoji: "💬", keywords: ["prat", "samtal", "möte"] },
	{ emoji: "❤️", keywords: ["kärlek", "hjärta"] },
	{ emoji: "🔥", keywords: ["viktig", "urgent", "eld"] },
	{ emoji: "✅", keywords: ["klar", "klart", "klapp", "check"] },
	{ emoji: "⏰", keywords: ["klocka", "tid", "alarm", "viktig"] },
	{ emoji: "🔔", keywords: ["notis", "alarm", "påminnelse"] },
	{ emoji: "🚗", keywords: ["bil", "resa"] },
	{ emoji: "💡", keywords: ["idé", "tips", "inspiration"] },
]

/**
 * Mönster som matchar ett enskilt emoji, inklusive flaggor, hudtoner och
 * sekvenser sammanbundna med ZWJ (t.ex. "👨‍👩‍👧").
 *
 * Endast tecken med emoji-presentation matchas, så att vanliga symboler
 * som "©" eller "▶" inte blir felaktigt uppfattade som emoji.
 */
const EMOJI_PATTERN = new RegExp(
	String.raw`(?:\p{Regional_Indicator}){2}|(?:\p{Emoji_Presentation}|\p{Extended_Pictographic}\uFE0F)(?:\uFE0F|\p{Emoji_Modifier})*(?:\u200D(?:\p{Emoji_Presentation}|\p{Extended_Pictographic}\uFE0F)(?:\uFE0F|\p{Emoji_Modifier})*)*`,
	"gu",
)

/**
 * Plockar ut det första emoji som finns i texten.
 *
 * Returnerar emoji:t och texten utan det. Emoji sommar längre fram i
 * meningen tas också bort, eftersom de är lika relevanta som ett i
 * början. Användaren kan alltså skriva "🏃 Löpning kl 18" lika väl som
 * "Löpning kl 18 🏃".
 */
export function extractEmoji(text: string): { emoji: string | null; text: string } {
	const match = EMOJI_PATTERN.exec(text)
	if (!match || match.index === undefined) {
		EMOJI_PATTERN.lastIndex = 0
		return { emoji: null, text }
	}

	EMOJI_PATTERN.lastIndex = 0
	const emoji = match[0]
	const without = `${text.slice(0, match.index)} ${text.slice(match.index + emoji.length)}`
	return { emoji, text: without.replace(/\s+/g, " ").trim() }
}

/**
 * Söker i katalogen efter emoji som matchar fritext.
 *
 * Matchar på sökord och på emoji-tecknet självt. Tom text ger en tom
 * lista, så att väljaren kan visa ett "ingen träffar"-meddelande.
 */
export function searchEmojis(query: string): EmojiEntry[] {
	const trimmed = query.trim().toLowerCase()
	if (!trimmed) return []

	return EMOJI_CATALOGUE.filter(
		(entry) =>
			entry.emoji === trimmed ||
			entry.keywords.some((keyword) => keyword.includes(trimmed)),
	)
}
