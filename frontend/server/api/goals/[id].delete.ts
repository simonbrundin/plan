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

	// Delete the goal (cascades to user_goals and goal_relations)
	await sql`DELETE FROM goals WHERE id = ${goalId}`;

	return { deleted: true };
});
