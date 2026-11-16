/**
 * Gör om en målbeskrivning till en kort rad som passar i en lista.
 * Radbrytningar och upprepade mellanrum tas bort, och texten kortas
 * ned med en ellipsis om den blir för lång.
 */
export function descriptionPreview(
	text: string | null | undefined,
	maxLength = 120,
): string {
	if (!text) return ""

	const collapsed = text.replace(/\s+/g, " ").trim()
	if (collapsed.length <= maxLength) return collapsed

	return `${collapsed.slice(0, maxLength - 1).trimEnd()}…`
}
