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

	const body = await readBody(event);
	const { childId, parentId, order, weight } = body;

	if (!childId || !parentId) {
		throw createError({
			statusCode: 400,
			message: "childId and parentId are required",
		});
	}

	// Verify user has access to the parent goal
	const [parentAccess] = await sql<any[]>`
    SELECT 1 FROM user_goals 
    WHERE user_id = ${user.id} AND goal_id = ${parentId}
    LIMIT 1
  `;

	if (!parentAccess) {
		throw createError({
			statusCode: 403,
			message: "Access denied",
		});
	}

	// Build update query
	const fields: string[] = [];
	const values: any[] = [];
	let paramIndex = 1;

	if (order !== undefined) {
		fields.push(`"order" = $${paramIndex++}`);
		values.push(order);
	}
	if (weight !== undefined) {
		fields.push(`weight = $${paramIndex++}`);
		values.push(weight);
	}

	if (fields.length === 0) {
		throw createError({
			statusCode: 400,
			message: "No fields to update",
		});
	}

	values.push(parentId, childId);
	const query = `UPDATE goal_relations SET ${fields.join(", ")} WHERE parent_id = $${paramIndex++} AND child_id = $${paramIndex}`;

	await sql.unsafe(query, values);

	const [updated] = await sql<any[]>`
    SELECT * FROM goal_relations 
    WHERE parent_id = ${parentId} AND child_id = ${childId}
  `;

	return updated;
});
