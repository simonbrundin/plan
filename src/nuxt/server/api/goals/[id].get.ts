import { sql } from "../../utils/db";
import { getCurrentUser } from "../../utils/auth";

export default defineEventHandler(async (event) => {
	const user = await getCurrentUser(event);

	if (!user) {
		throw createError({
			statusCode: 401,
			message: "Unauthorized",
		});
	}

	const goalId = parseInt(getRouterParam(event, "id") || "0");

	if (!goalId) {
		throw createError({
			statusCode: 400,
			message: "Invalid goal ID",
		});
	}

	// Check if user has access
	const [access] = await sql<any[]>`
    SELECT 1 FROM user_goals 
    WHERE user_id = ${user.id} AND goal_id = ${goalId}
    LIMIT 1
  `;

	if (!access) {
		throw createError({
			statusCode: 403,
			message: "Access denied",
		});
	}

	// Get all user goals (needed early for dependency enrichment)
	const allGoals = await sql<any[]>`
    SELECT g.* FROM goals g
    INNER JOIN user_goals ug ON g.id = ug.goal_id
    WHERE ug.user_id = ${user.id}
  `;

	// Build a map of goal_id -> goal for quick lookup
	const allGoalsMap = new Map<number, any>();
	for (const g of allGoals) {
		allGoalsMap.set(g.id, g);
	}

	// Get the goal
	const [goal] = await sql<any[]>`
    SELECT * FROM goals WHERE id = ${goalId}
  `;

	if (!goal) {
		throw createError({
			statusCode: 404,
			message: "Goal not found",
		});
	}

	// Get child relations with their dependencies
	const childRelations = await sql<any[]>`
    SELECT gr.*, g.* FROM goal_relations gr
    INNER JOIN goals g ON g.id = gr.child_id
    WHERE gr.parent_id = ${goalId}
    ORDER BY gr.order ASC
  `;

	// Get dependencies for each child
	const childIds = childRelations.map((c) => c.id);
	const childDependenciesMap = new Map<
		number,
		{ dependsOn: any[]; blocking: any[] }
	>();

	if (childIds.length > 0) {
		// Get all dependencies where the child is the dependent (dependsOn)
		const deps = await sql<any[]>`
      SELECT gd.*, g.title as depends_on_title FROM goal_dependencies gd
      INNER JOIN goals g ON g.id = gd.depends_on_id
      WHERE gd.goal_id = ANY(${childIds})
    `;

		// Get all dependencies where the child is the dependency (blocking)
		const blocks = await sql<any[]>`
      SELECT gd.*, g.title as dependent_title FROM goal_dependencies gd
      INNER JOIN goals g ON g.id = gd.goal_id
      WHERE gd.depends_on_id = ANY(${childIds})
    `;

		// Build the map for each child
		for (const id of childIds) {
			childDependenciesMap.set(id, {
				dependsOn: deps
					.filter((d) => d.goal_id === id)
					.map((d) => ({
						depends_on_id: d.depends_on_id,
						title: d.depends_on_title,
						finished: null,
					})),
				blocking: blocks
					.filter((b) => b.depends_on_id === id)
					.map((b) => ({
						goal_id: b.goal_id,
						title: b.dependent_title,
					})),
			});
		}
	}

	// Merge dependencies into child relations (with correct finished status)
	const childrenWithDeps = childRelations.map((child) => {
		const deps = childDependenciesMap.get(child.id)?.dependsOn || [];
		const blocking = childDependenciesMap.get(child.id)?.blocking || [];

		// Enrich with finished status from allGoals
		const enrichedDeps = deps.map((d) => {
			const goalInfo = allGoalsMap.get(d.depends_on_id);
			return {
				...d,
				finished: goalInfo?.finished || null,
			};
		});

		return {
			...child,
			dependsOn: enrichedDeps,
			blocking: blocking,
		};
	});

	console.log(
		`[GET /api/goals/${goalId}] Returning ${childrenWithDeps.length} children with dependencies`,
	);

	// Get parent relations
	const parentRelations = await sql<any[]>`
    SELECT gr.*, g.* FROM goal_relations gr
    INNER JOIN goals g ON g.id = gr.parent_id
    WHERE gr.child_id = ${goalId}
  `;

	// Get dependencies for this goal
	const dependencies = await sql<any[]>`
    SELECT gd.*, g.* FROM goal_dependencies gd
    INNER JOIN goals g ON g.id = gd.depends_on_id
    WHERE gd.goal_id = ${goalId}
    ORDER BY g.created ASC
  `;

	// Get goals that depend on this goal (blocking)
	const blocking = await sql<any[]>`
    SELECT gd.*, g.* FROM goal_dependencies gd
    INNER JOIN goals g ON g.id = gd.goal_id
    WHERE gd.depends_on_id = ${goalId}
    ORDER BY g.created ASC
  `;

	return {
		goal,
		children: childrenWithDeps,
		parents: parentRelations,
		allGoals,
		dependencies,
		dependsOn: dependencies,
		blocking,
	};
});
