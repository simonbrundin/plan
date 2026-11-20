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

/**
 * Returnerar den plats i listan där ett nytt undermål ska läggas in.
 *
 * Det nya målet läggs under det markerade undermålet. Om huvudmålet är
 * markerat läggs det först i listan (alltså som barn nr 1). Finns inget
 * markerat undermål läggs det sist.
 *
 * allIds = alla undermål i sparad ordning
 * visibleIds = de undermål som syns i listan (kan vara filtrerade)
 * selectedIndex = index i visibleIds för det markerade undermålet
 */
export function newChildInsertIndex(
	allIds: number[],
	visibleIds: number[],
	selectedIndex: number,
	goalSelected: boolean,
): number {
	if (goalSelected) {
		return 0;
	}

	const selectedId = visibleIds[selectedIndex];
	if (selectedId === undefined) {
		return allIds.length;
	}

	const position = allIds.indexOf(selectedId);
	if (position === -1) {
		return allIds.length;
	}

	return position + 1;
}

/**
 * Returnerar listan med ett nytt undermål insatt på en given plats.
 */
export function insertChildId(
	allIds: number[],
	insertIndex: number,
	newId: number,
): number[] {
	const ids = [...allIds];
	const index = Math.max(0, Math.min(insertIndex, ids.length));
	ids.splice(index, 0, newId);
	return ids;
}
