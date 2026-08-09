import { sql } from "../../utils/db";
import { getCurrentUser } from "../../utils/auth";

export default defineEventHandler(async (event) => {
	try {
		const user = await getCurrentUser(event);

		if (!user) {
			throw createError({
				statusCode: 401,
				message: "Unauthorized",
			});
		}

		// Safely read body
		let body: any = {};
		try {
			body = (await readBody(event)) ?? {};
		} catch (_) {
			body = {};
		}

		const goalId = Number(body.goalId);
		const dependsOnId = Number(body.dependsOnId);

		if (
			!goalId ||
			!dependsOnId ||
			Number.isNaN(goalId) ||
			Number.isNaN(dependsOnId)
		) {
			throw createError({
				statusCode: 400,
				message: "goalId and dependsOnId are required and must be numbers",
			});
		}

		if (goalId === dependsOnId) {
			throw createError({
				statusCode: 400,
				message: "A goal cannot depend on itself",
			});
		}

		// Verify user has access to both goals
		const [goalAccess] = await sql<any[]>`
			SELECT 1 FROM user_goals
			WHERE user_id = ${user.id} AND goal_id = ${goalId}
			LIMIT 1
		`;
		const [dependsOnAccess] = await sql<any[]>`
			SELECT 1 FROM user_goals
			WHERE user_id = ${user.id} AND goal_id = ${dependsOnId}
			LIMIT 1
		`;

		if (!goalAccess || !dependsOnAccess) {
			throw createError({
				statusCode: 403,
				message: "Access denied",
			});
		}

		// Check for existing dependency (to avoid duplicates)
		const [existing] = await sql<any[]>`
			SELECT 1 FROM goal_dependencies
			WHERE goal_id = ${goalId} AND depends_on_id = ${dependsOnId}
			LIMIT 1
		`;

		if (existing) {
			// Already exists, just return it
			const [existingDep] = await sql<any[]>`
				SELECT * FROM goal_dependencies
				WHERE goal_id = ${goalId} AND depends_on_id = ${dependsOnId}
			`;
			return existingDep;
		}

		// Cycle detection: Check if adding goalId→dependsOnId would create a cycle
		// We need to check if there's already a path from dependsOnId to goalId
		// If yes, adding this dependency would create a cycle
		const cyclePath = await checkForCycle(goalId, dependsOnId);
		if (cyclePath.length > 0) {
			throw createError({
				statusCode: 400,
				message: `Cykel upptäckt: Detta skulle skapa en cykel genom: ${cyclePath.join(" → ")} → ${goalId}`,
			});
		}

		// Create the dependency
		const [dependency] = await sql<any[]>`
			INSERT INTO goal_dependencies (goal_id, depends_on_id)
			VALUES (${goalId}, ${dependsOnId})
			RETURNING *
		`;

		return dependency;
	} catch (err: any) {
		if (err?.statusCode && err.statusCode < 500) {
			throw err;
		}
		console.error("[dependencies.post] Unexpected error:", err);
		throw createError({
			statusCode: 500,
			message: err?.message ?? "Internal server error",
		});
	}
});

// DFS to detect cycles in the dependency graph
async function checkForCycle(
	startGoalId: number,
	endGoalId: number,
): Promise<number[]> {
	// Check if there's a path from endGoalId to startGoalId
	// If yes, adding startGoalId→endGoalId would create a cycle
	const visited = new Set<number>();
	const path: number[] = [];

	async function dfs(currentId: number): Promise<boolean> {
		if (currentId === startGoalId) {
			return true; // Found a path back to start - this would create a cycle
		}

		if (visited.has(currentId)) {
			return false;
		}

		visited.add(currentId);
		path.push(currentId);

		// Get all goals that currentId depends on (currentId → depends_on)
		const dependencies = await sql<any[]>`
			SELECT depends_on_id FROM goal_dependencies
			WHERE goal_id = ${currentId}
		`;

		for (const dep of dependencies) {
			if (await dfs(dep.depends_on_id)) {
				return true;
			}
		}

		path.pop();
		return false;
	}

	await dfs(endGoalId);
	return path.slice(0, -1); // Return path without the start node
}
