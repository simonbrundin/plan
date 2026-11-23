/**
 * Hierarki för "Gör nu"-listan.
 *
 * Ett mål kan ligga under en förälder (via goal_relations). I /now ska ett
 * sådant mål visas indenterat under föräldern, och det ska bara gå att
 * sortera bland sina syskon - föräldern sorteras i sin tur i toppnivålistan.
 *
 * Funktionerna här är rena: de tar en lista med mål och returnerar en träd
 * som gränssnittet kan rita och sortera.
 */

/** En förälder till ett mål. */
export interface HierarchyParent {
	id: number;
}

/** Ett mål som kan ingå i hierarkin. */
export interface HierarchyGoal {
	id: number;
	// Föräldrarna kommer sorterade på vikt (tyngsta först) från servern.
	parents?: HierarchyParent[] | null;
}

/** Trädet som målen bildar. */
export interface Hierarchy {
	/** Föräldern varje mål hänger under. null betyder toppnivå. */
	parentOf: Map<number, number | null>;
	/** Målen på toppnivå, i listordning. */
	roots: number[];
}

/** En rad i den visade listan. */
export interface DisplayRow {
	/** Målets id. */
	id: number;
	/** Antal synliga föräldrar ovanför raden. */
	depth: number;
	/** Föräldern raden visas under. null betyder toppnivå. */
	parentId: number | null;
}

/** Den visade listan, med syskon per rad. */
export interface DisplayTree {
	/** Raderna i den ordning de ska visas. */
	rows: DisplayRow[];
	/** Målen som visas under samma förälder, per mål. */
	siblingsOf: Map<number, number[]>;
}

/**
 * Bygger trädet över målen.
 *
 * Ett mål läggs under den första föräldern som också finns i listan. Finns
 * ingen sådan förälder ligger målet på toppnivå.
 */
export function buildHierarchy(goals: HierarchyGoal[]): Hierarchy {
	const ids = new Set(goals.map((goal) => goal.id));
	const parentOf = new Map<number, number | null>();

	for (const goal of goals) {
		const parent = (goal.parents ?? []).find((p) => ids.has(p.id));
		parentOf.set(goal.id, parent ? parent.id : null);
	}

	breakCycles(parentOf);

	const roots: number[] = [];
	for (const goal of goals) {
		if (parentOf.get(goal.id) === null) {
			roots.push(goal.id);
		}
	}

	return { parentOf, roots };
}

/**
 * Bryter eventuella cykler i förälderekrelationerna.
 *
 * Om A är förälder till B och B är förälder till A finns ingen toppnivå att
 * hänga upp trädet på, och målen skulle försvinna ur listan. Den kedjan
 * bryts då genom att den siste målet i kedjan blir toppnivå.
 */
function breakCycles(parentOf: Map<number, number | null>): void {
	for (const start of parentOf.keys()) {
		const seen = new Set<number>([start]);
		let current = parentOf.get(start) ?? null;

		while (current !== null) {
			if (seen.has(current)) {
				parentOf.set(current, null);
				break;
			}
			seen.add(current);
			current = parentOf.get(current) ?? null;
		}
	}
}

/**
 * Bygger den lista som ska visas.
 *
 * Bara synliga mål tas med. Ett mål läggs under sin närmaste synliga
 * förälder - om föräldern inte syns (t.ex. filtrerad bort) hamnar målet
 * på toppnivå i stället.
 *
 * visibleIds ska vara de synliga målen i listordning.
 */
export function buildDisplayTree(
	visibleIds: number[],
	hierarchy: Hierarchy,
): DisplayTree {
	const visible = new Set(visibleIds);
	const childrenOf = new Map<number | null, number[]>();
	const displayParentOf = new Map<number, number | null>();

	for (const id of visibleIds) {
		const parent = nearestVisibleAncestor(id, hierarchy, visible);
		displayParentOf.set(id, parent);
		const children = childrenOf.get(parent) ?? [];
		children.push(id);
		childrenOf.set(parent, children);
	}

	const rows: DisplayRow[] = [];
	const siblingsOf = new Map<number, number[]>();

	// Djup-först: toppnivåmålen först, sedan deras barn, och så vidare.
	const visit = (parent: number | null, depth: number) => {
		const children = childrenOf.get(parent) ?? [];
		for (const child of children) {
			siblingsOf.set(child, children);
			rows.push({ id: child, depth, parentId: parent });
			visit(child, depth + 1);
		}
	};
	visit(null, 0);

	return { rows, siblingsOf };
}

/**
 * Den närmaste föräldern som också syns i listan.
 *
 * Om ingen förälder syns blir svaret null - då ligger målet på toppnivå.
 */
function nearestVisibleAncestor(
	id: number,
	hierarchy: Hierarchy,
	visible: Set<number>,
): number | null {
	const seen = new Set<number>([id]);
	let current = hierarchy.parentOf.get(id) ?? null;

	while (current !== null && !visible.has(current)) {
		if (seen.has(current)) return null;
		seen.add(current);
		current = hierarchy.parentOf.get(current) ?? null;
	}

	return current;
}

/**
 * Byter plats på två mål i listan.
 *
 * Används när ett mål flyttas upp eller ner: målen byter plats i den
 * sparade listan, medan målen däremellan ligger kvar.
 */
export function swapIds(ids: number[], firstId: number, secondId: number): number[] | null {
	const first = ids.indexOf(firstId);
	const second = ids.indexOf(secondId);
	if (first === -1 || second === -1) return null;

	const reordered = [...ids];
	reordered[first] = secondId;
	reordered[second] = firstId;
	return reordered;
}
