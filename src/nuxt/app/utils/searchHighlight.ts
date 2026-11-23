/**
 * Markering i en sökresultatlista.
 *
 * -1 betyder att inget är markerat. Från det läget markerar nedåt första
 * träffen och uppåt sista träffen. Markeringen stannar alltid inom listan.
 *
 * Sökfälten använder detta för att Enter ska skapa ett nytt mål så länge
 * inget befintligt mål är markerat. Ett befintligt mål väljs bara när
 * användaren har markerat det med piltangenterna.
 */
export function moveSearchHighlight(
	index: number,
	resultCount: number,
	step: number,
): number {
	if (resultCount <= 0) return -1;
	if (index < 0) return step > 0 ? 0 : resultCount - 1;

	const next = index + step;
	return Math.min(Math.max(next, 0), resultCount - 1);
}
