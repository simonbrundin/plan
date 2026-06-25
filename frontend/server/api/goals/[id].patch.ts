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
	const body = await readBody(event);

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

	// Build update dynamically
	const fields: string[] = [];
	const values: any[] = [];
	let paramIndex = 1;

	if (body.title !== undefined) {
		fields.push(`title = $${paramIndex++}`);
		values.push(body.title);
	}
	if (body.icon !== undefined) {
		fields.push(`icon = $${paramIndex++}`);
		values.push(body.icon);
	}
	if (body.finished !== undefined) {
		fields.push(`finished = $${paramIndex++}`);
		values.push(body.finished ? new Date(body.finished) : null);
	}

	if (fields.length === 0) {
		throw createError({
			statusCode: 400,
			message: "No fields to update",
		});
	}

	// Build and execute query
	const query = `UPDATE goals SET ${fields.join(", ")} WHERE id = $${paramIndex}`;
	values.push(goalId);

	await sql.unsafe(query, values);

	const [updatedGoal] = await sql<
		any[]
	>`SELECT * FROM goals WHERE id = ${goalId}`;

	return updatedGoal;
});
