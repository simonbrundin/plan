import type {
	ChildWithDeps,
	GetGoalResponse,
	Goal,
	GoalDependency,
	GoalRelation,
} from '~/types/goal';

// Målgraphen: alla mål, relationerna mellan dem och beroendena.
export interface GoalGraph {
	goals: Goal[];
	relations: GoalRelation[];
	dependencies: GoalDependency[];
}

// Bygger en målsidas data ur grafen.
//
// Grafen innehåller allt som en målsida visar: målet, dess barn (med
// ordning och vikt), dess föräldrar och dess beroenden. Eftersom ingenting
// behöver fråga servern går det att byta målsida direkt - bara en
// beräkning i klienten.
export function buildGoalData(
	graph: GoalGraph,
	id: number,
): GetGoalResponse | null {
	const goal = graph.goals.find((g) => g.id === id);
	if (!goal) {
		return null;
	}

	const childRelations = graph.relations
		.filter((relation) => relation.parent_id === id)
		// Sorterar på ordning, och på id vid lika ordning - samma sortering
		// som servern gör, så att båda ger samma lista.
		.sort((a, b) => a.order - b.order || a.child_id - b.child_id);

	const children: ChildWithDeps[] = childRelations
		.map((relation) => {
			const child = graph.goals.find((g) => g.id === relation.child_id);
			if (!child) {
				return null;
			}
			return {
				...child,
				order: relation.order,
				weight: relation.weight,
				dependsOn: dependsOnOf(graph, relation.child_id),
				blocking: blockingOf(graph, relation.child_id),
			};
		})
		.filter((child): child is ChildWithDeps => child !== null);

	const parents = graph.relations
		.filter((relation) => relation.child_id === id)
		.map((relation) => graph.goals.find((g) => g.id === relation.parent_id))
		.filter((parent): parent is Goal => Boolean(parent));

	return {
		goal: {
			...goal,
			childRelations: childRelations.map((relation) => ({
				child_id: relation.child_id,
				order: relation.order,
				weight: relation.weight,
			})),
			parentRelations: parents,
		},
		children,
		parents,
		allGoals: graph.goals,
		dependencies: graph.dependencies.filter((d) => d.goal_id === id),
		dependsOn: dependsOnOf(graph, id),
		blocking: blockingOf(graph, id),
	};
}

// Mål som ett mål beror på.
export function dependsOnOf(graph: GoalGraph, id: number): Goal[] {
	return graph.dependencies
		.filter((dependency) => dependency.goal_id === id)
		.map((dependency) =>
			graph.goals.find((g) => g.id === dependency.depends_on_id),
		)
		.filter((goal): goal is Goal => Boolean(goal));
}

// Mål som beror på ett mål.
export function blockingOf(graph: GoalGraph, id: number): Goal[] {
	return graph.dependencies
		.filter((dependency) => dependency.depends_on_id === id)
		.map((dependency) => graph.goals.find((g) => g.id === dependency.goal_id))
		.filter((goal): goal is Goal => Boolean(goal));
}
