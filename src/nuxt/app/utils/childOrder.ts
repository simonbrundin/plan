/**
 * Hjälpfunktioner för att ändra ordningen på en fils undermål.
 *
 * Ordningen lagras i fältet "order" på relationen mellan målet och dess
 * undermål. Listan som visas i gränssnittet kan vara filtrerad (t.ex. dölja
 * klara mål), men den ordning som sparas ska alltid gälla alla undermål.
 */

/**
 * Flyttar ett undermål ett steg upp eller ner i den synliga listan och
 * returnerar den nya ordningen för alla undermål.
 *
 * Dolda undermål (t.ex. klara mål) behåller sin plats - bara den valda
 * målet och dess synliga grann byter plats.
 *
 * Returnerar null om flytten inte är möjlig (t.ex. redan överst eller
 * längst ner i listan).
 */
export function moveChildId(
	allIds: number[],
	visibleIds: number[],
	movingId: number,
	direction: "up" | "down",
): number[] | null {
	const visiblePos = visibleIds.indexOf(movingId);
	if (visiblePos === -1) {
		return null;
	}

	const targetPos = direction === "up" ? visiblePos - 1 : visiblePos + 1;
	if (targetPos < 0 || targetPos >= visibleIds.length) {
		return null;
	}

	const targetId = visibleIds[targetPos];
	const movingPos = allIds.indexOf(movingId);
	const targetPosInAll = allIds.indexOf(targetId);
	if (movingPos === -1 || targetPosInAll === -1) {
		return null;
	}

	const reordered = [...allIds];
	reordered[movingPos] = targetId;
	reordered[targetPosInAll] = movingId;
	return reordered;
}

/**
 * Returnerar nästa lediga ordningsnummer för ett nytt undermål.
 */
export function nextChildOrder(orders: number[]): number {
	const highest = orders.reduce((acc, order) => Math.max(acc, order), -1);
	return highest + 1;
}
