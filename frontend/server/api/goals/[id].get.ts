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

	// Get child relations
	const childRelations = await sql<any[]>`
    SELECT gr.*, g.* FROM goal_relations gr
    INNER JOIN goals g ON g.id = gr.child_id
    WHERE gr.parent_id = ${goalId}
    ORDER BY gr.order ASC
  `;

	// Get parent relations
	const parentRelations = await sql<any[]>`
    SELECT gr.*, g.* FROM goal_relations gr
    INNER JOIN goals g ON g.id = gr.parent_id
    WHERE gr.child_id = ${goalId}
  `;

	// Get all user goals
	const allGoals = await sql<any[]>`
    SELECT g.* FROM goals g
    INNER JOIN user_goals ug ON g.id = ug.goal_id
    WHERE ug.user_id = ${user.id}
  `;

	return {
		goal,
		children: childRelations,
		parents: parentRelations,
		allGoals,
	};
});
