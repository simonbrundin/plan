import { defineStore } from "pinia";
import type { Goal, GoalDependency, GoalRelation, GoalStatus } from "~/types/goal";

// Ett pågående anrop mot servern. Det ligger utanför state eftersom en
// Promise inte hör hemma i en serialiserbar store.
let inflightLoad: Promise<void> | null = null;
let inflightStatusLoad: Promise<void> | null = null;

export const useGoalsStore = defineStore("goals", {
	// Skip hydration is enforced globally by the pinia-hydration plugin.

	state: () => ({
		goals: [] as Goal[],
		relations: [] as GoalRelation[],
		dependencies: [] as GoalDependency[],
		statuses: [] as GoalStatus[],
		statusesLoaded: false,
		isLoaded: false,
		// När grafen senast hämtades (ms sedan epoch). Används för att
		// veta om data är färsk nog eller måste hämtas om.
		loadedAt: 0,
		error: null as string | null,
	}),

	getters: {
		// Hitta mål baserat på ID
		getGoalById: (state) => (id: number) => {
			return state.goals.find((g) => g.id === id);
		},

		// Sök mål baserat på text
		searchGoals: (state) => (query: string) => {
			if (!query.trim()) return [];
			const lowerQuery = query.toLowerCase();
			return state.goals.filter((g) =>
				g.title.toLowerCase().includes(lowerQuery),
			);
		},

		// Hämta föräldrar för ett mål
		getParents: (state) => (goalId: number) => {
			const parentIds = state.relations
				.filter((r) => r.child_id === goalId)
				.map((r) => r.parent_id);

			return state.goals.filter((g) => parentIds.includes(g.id));
		},

		// Hämta barn för ett mål
		getChildren: (state) => (goalId: number) => {
			const childIds = state.relations
				.filter((r) => r.parent_id === goalId)
				.map((r) => r.child_id);

			return state.goals.filter((g) => childIds.includes(g.id));
		},

		// Antalet undermål per mål, räknat ur relationerna.
		//
		// Används för att visa hur många barn ett mål har utan att varje
		// rad behöver fråga servern. "finished" är antalet klara barn.
		childStats: (state) => {
			const finishedById = new Map(
				state.goals.map((goal) => [goal.id, !!goal.finished]),
			);
			const stats = new Map<number, { total: number; finished: number }>();

			for (const relation of state.relations) {
				const current = stats.get(relation.parent_id) ?? {
					total: 0,
					finished: 0,
				};
				current.total += 1;
				if (finishedById.get(relation.child_id)) {
					current.finished += 1;
				}
				stats.set(relation.parent_id, current);
			}

			return stats;
		},

		// Hämta grundmål (mål utan förälder)
		getRootGoals: (state) => {
			const childIds = new Set(state.relations.map((r) => r.child_id));
			return state.goals.filter((g) => !childIds.has(g.id));
		},

		// Hämta inbox-mål
		getInboxGoals: (state) => {
			return state.goals.filter((g) => g.inbox === 1);
		},
	},

	actions: {
		async loadStatuses() {
			if (this.statusesLoaded) return;
			if (inflightStatusLoad) return inflightStatusLoad;

			inflightStatusLoad = (async () => {
				try {
					const { loadGoalStatuses } = useGoalApi();
					this.statuses = await loadGoalStatuses();
					this.statusesLoaded = true;
				} finally {
					inflightStatusLoad = null;
				}
			})();

			return inflightStatusLoad;
		},

		// Ladda hela målgraphen från servern.
		//
		// force = true hämtar alltid om. Annars hoppas hämtningen över
		// om grafen redan är laddad - det gör att byten mellan målsidor
		// inte kräver en serverfråga.
		async loadGraph(force = false) {
			if (!force && this.isLoaded) {
				return;
			}

			// Ett anrop pågår redan - dela det i stället för att köra två.
			if (inflightLoad) {
				return inflightLoad;
			}

			inflightLoad = (async () => {
				try {
					const { loadGoalGraph } = useGoalApi();
					const graph = await loadGoalGraph();

					this.goals = graph.goals ?? [];
					this.relations = graph.relations ?? [];
					this.dependencies = graph.dependencies ?? [];
					this.isLoaded = true;
					this.loadedAt = Date.now();
					this.error = null;
				} catch (error: any) {
					this.error =
						error?.data?.message || error?.message || "Kunde inte hämta målen";
					throw error;
				} finally {
					inflightLoad = null;
				}
			})();

			return inflightLoad;
		},

		// Sätter en ny ordning på ett måls barn (lokalt, utan serverfråga).
		setChildOrder(parentId: number, childIds: number[]) {
			const orderById = new Map(childIds.map((id, index) => [id, index]));

			this.relations = this.relations.map((relation) => {
				if (relation.parent_id !== parentId) {
					return relation;
				}
				const order = orderById.get(relation.child_id);
				return order === undefined ? relation : { ...relation, order };
			});
		},

		// Ladda alla mål och relationer
		async loadGoals() {
			// This method is now handled in components
			// Keeping for compatibility
			this.isLoaded = true;
		},

		// Lägg till ett nytt mål
		addGoal(goal: Goal) {
			this.goals.push(goal);
		},

		// Uppdatera ett mål
		updateGoal(id: number, updates: Partial<Goal>) {
			const index = this.goals.findIndex((g) => g.id === id);
			if (index !== -1) {
				this.goals[index] = Object.assign(
					{},
					this.goals[index],
					updates,
				) as Goal;
			}
		},

		// Ta bort ett mål
		removeGoal(id: number) {
			this.goals = this.goals.filter((g) => g.id !== id);
			// Ta även bort relaterade relationer
			this.relations = this.relations.filter(
				(r) => r.child_id !== id && r.parent_id !== id,
			);
			this.dependencies = this.dependencies.filter(
				(d) => d.goal_id !== id && d.depends_on_id !== id,
			);
		},

		// Lägg till en relation
		addRelation(childId: number, parentId: number) {
			// Kolla om relationen redan finns
			const exists = this.relations.some(
				(r) => r.child_id === childId && r.parent_id === parentId,
			);

			if (!exists) {
				this.relations.push({
					parent_id: parentId,
					child_id: childId,
					order: 0,
					weight: 0,
				});
			}
		},

		// Ta bort en relation
		removeRelation(childId: number, parentId: number) {
			this.relations = this.relations.filter(
				(r) => !(r.child_id === childId && r.parent_id === parentId),
			);
		},

		// Clear error state
		clearError() {
			this.error = null;
		},

		// Flytta mål från inbox till huvudstrukturen
		moveGoalFromInbox(id: number) {
			const goal = this.goals.find((g) => g.id === id);
			if (goal) {
				goal.inbox = 0;
			}
		},

		// Reset store state (useful for error recovery)
		reset() {
			this.goals = [];
			this.relations = [];
			this.dependencies = [];
			this.isLoaded = false;
			this.loadedAt = 0;
			this.error = null;
		},
	},
});
