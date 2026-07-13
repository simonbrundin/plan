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

	if (!goalId || isNaN(goalId)) {
		throw createError({
			statusCode: 400,
			message: "Invalid goal ID",
		});
	}

	// Check if user has access
	const accessResult = await sql`
    SELECT 1 FROM user_goals 
    WHERE user_id = ${user.id} AND goal_id = ${goalId}
    LIMIT 1
  `;

	if (!accessResult || accessResult.length === 0) {
		throw createError({
			statusCode: 403,
			message: "Access denied",
		});
	}

	// Delete in correct order to avoid FK violations
	// 1. Remove from user_goals
	await sql`DELETE FROM user_goals WHERE goal_id = ${goalId}`;

	// 2. Remove from goal_relations (both as parent and child)
	await sql`DELETE FROM goal_relations WHERE parent_id = ${goalId}`;
	await sql`DELETE FROM goal_relations WHERE child_id = ${goalId}`;

	// 3. Finally delete the goal
	await sql`DELETE FROM goals WHERE id = ${goalId}`;

	return { success: true, deleted: goalId };
});
