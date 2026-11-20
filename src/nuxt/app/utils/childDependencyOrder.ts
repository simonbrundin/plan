/**
 * Undermål som beror på varandra.
 *
 * Ett undermål som väntar på ett annat undermål i samma lista ska ligga
 * direkt under det mål det väntar på - som ett underelement i listan.
 * Då blir indenteringen logisk: det indenterade målet hör till målet
 * ovanför, inte till något längre upp.
 *
 * Listan i gränssnittet visar order-fältet. Därför skrivs order-fältet
 * om till trädets ordning - annars skulle listan visa en ordning som inte
 * gäller när den sparas.
 */

/** Det som behövs för att avgöra om ett mål väntar på ett annat. */
export interface DependencyLike {
	id: number;
	finished?: string | null;
	dependsOn?: { id: number }[] | null;
}

/** Ett undermål i trädet. */
export interface DependencyTreeItem<T> {
	child: T;
	// Nivå i trädet. 0 = ligger direkt i listan.
	depth: number;
	// Id på det mål som målet ligger under. Null för mål på översta nivån.
	parentId: number | null;
}

/**
 * Returnerar beroenden som finns mellan mål i samma lista.
 *
 * Beroenden till mål utanför listan ignoreras - de kan inte påverka
 * strukturen i just den här listan. Beroenden till redan klara mål
 * ignoreras också: de blockerar inte längre något.
 */
function internalDependencies<T extends DependencyLike>(
	children: T[],
): Map<number, number[]> {
	const byId = new Map(children.map((child) => [child.id, child]));
	const deps = new Map<number, number[]>();

	for (const child of children) {
		const ids: number[] = [];

		for (const dep of child.dependsOn ?? []) {
			if (dep.id === child.id) continue;

			const target = byId.get(dep.id);
			if (!target) continue;
			if (target.finished) continue;

			ids.push(dep.id);
		}

		deps.set(child.id, ids);
	}

	return deps;
}

/**
 * Bygger trädet av undermål.
 *
 * Ett mål som väntar på ett annat mål i listan blir ett underelement till
 * det målet. Om målet väntar på flera mål väljs det som ligger längst ned
 * i beroendekedjan - så att kedjan hänger ihop.
 *
 * Cykler (A beroende av B och B beroende av A) går inte att lösa - då
 * behålls listans ursprungliga ordning.
 */
export function buildDependencyTree<T extends DependencyLike>(
	children: T[],
): DependencyTreeItem<T>[] {
	if (children.length === 0) {
		return [];
	}

	const deps = internalDependencies(children);

	// Förälder i trädet: det beroende som ligger längst ned i kedjan.
	const parentIds = new Map<number, number | null>();

	for (const child of children) {
		const waiting = deps.get(child.id) ?? [];
		let parent: number | null = null;

		for (const depId of waiting) {
			const isAncestor = waiting.some((other) => {
				if (other === depId) return false;
				// Om ett annat beroende väntar på det här beroendet så är det
				// en stamfader - då ligger det andra beroendet närmast.
				return (deps.get(other) ?? []).includes(depId);
			});
			if (isAncestor) continue;

			parent = depId;
			break;
		}

		parentIds.set(child.id, parent);
	}

	// Bryt cykler: om kedjan av föräldrar går tillbaka till starten tas
	// föräldern bort - annars skulle trädet aldrig ta slut.
	for (const child of children) {
		const seen = new Set<number>([child.id]);
		let current = parentIds.get(child.id) ?? null;

		while (current !== null && current !== undefined) {
			if (seen.has(current)) {
				parentIds.set(child.id, null);
				break;
			}
			seen.add(current);
			current = parentIds.get(current) ?? null;
		}
	}

	// Undermål per förälder.
	const kids = new Map<number, T[]>();
	for (const child of children) {
		const parent = parentIds.get(child.id);
		if (parent === null || parent === undefined) continue;

		const list = kids.get(parent) ?? [];
		list.push(child);
		kids.set(parent, list);
	}

	const tree: DependencyTreeItem<T>[] = [];
	const visited = new Set<number>();

	const visit = (child: T, depth: number) => {
		if (visited.has(child.id)) return;
		visited.add(child.id);

		tree.push({
			child,
			depth,
			parentId: parentIds.get(child.id) ?? null,
		});

		for (const kid of kids.get(child.id) ?? []) {
			visit(kid, depth + 1);
		}
	};

	// Mål utan förälder i listan - de ligger överst.
	for (const child of children) {
		const parent = parentIds.get(child.id);
		if (parent === null || parent === undefined) {
			visit(child, 0);
		}
	}

	// Mål som fastnat i en cykel läggs till sist.
	for (const child of children) {
		if (!visited.has(child.id)) {
			visit(child, 0);
		}
	}

	return tree;
}

/**
 * Returnerar en ny ordning där ett undermål har flyttats ett steg upp eller
 * ner, eller null om flytten inte är möjlig.
 *
 * Ett mål på översta nivån flyttas tillsammans med sina undermål - de
 * hör ihop. Ett undermål flyttas bara inom sin grupp, så att det aldrig
 * hamnar utanför det mål det väntar på.
 */
export function moveChildInTree(
	allIds: number[],
	tree: DependencyTreeItem<{ id: number }>[],
	movingId: number,
	direction: "up" | "down",
): number[] | null {
	const index = tree.findIndex((item) => item.child.id === movingId);
	if (index === -1) {
		return null;
	}

	const moving = tree[index];

	if (moving.depth === 0) {
		const branches = rootBranches(tree);
		const branchIndex = branches.findIndex((branch) =>
			branch.ids.includes(movingId),
		);
		if (branchIndex === -1) {
			return null;
		}

		const targetIndex = direction === "up" ? branchIndex - 1 : branchIndex + 1;
		const target = branches[targetIndex];
		if (!target) {
			return null;
		}

		return swapBlocks(allIds, branches[branchIndex].ids, target.ids, direction);
	}

	// Undermål flyttas bara inom sin grupp.
	const siblings = tree.filter((item) => item.parentId === moving.parentId);
	const siblingIndex = siblings.findIndex((item) => item.child.id === movingId);
	const target = siblings[siblingIndex + (direction === "up" ? -1 : 1)];
	if (!target) {
		return null;
	}

	return swapIds(allIds, movingId, target.child.id);
}

/** Delar upp trädet i grenar - varje gren är ett mål med sina undermål. */
function rootBranches(
	tree: DependencyTreeItem<{ id: number }>[],
): { ids: number[] }[] {
	const branches: { ids: number[] }[] = [];
	let current: { ids: number[] } | null = null;

	for (const item of tree) {
		if (item.depth === 0) {
			current = { ids: [item.child.id] };
			branches.push(current);
			continue;
		}
		if (current) {
			current.ids.push(item.child.id);
		}
	}

	return branches;
}

/** Byter plats mellan två mål i listan. */
function swapIds(allIds: number[], a: number, b: number): number[] | null {
	const aPos = allIds.indexOf(a);
	const bPos = allIds.indexOf(b);
	if (aPos === -1 || bPos === -1) {
		return null;
	}

	const result = [...allIds];
	result[aPos] = b;
	result[bPos] = a;
	return result;
}

/** Flyttar ett helt block av mål till en annan plats i listan. */
function swapBlocks(
	allIds: number[],
	moving: number[],
	target: number[],
	direction: "up" | "down",
): number[] | null {
	const movingSet = new Set(moving);
	const without = allIds.filter((id) => !movingSet.has(id));

	const anchor = direction === "up" ? target[0] : target[target.length - 1];
	const anchorPos = without.indexOf(anchor);
	if (anchorPos === -1) {
		return null;
	}

	const insertAt = direction === "up" ? anchorPos : anchorPos + 1;
	without.splice(insertAt, 0, ...moving);
	return without;
}
