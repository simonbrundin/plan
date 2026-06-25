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
	const { childId, parentId, order = 0, weight = 10 } = body;

	if (!childId || !parentId) {
		throw createError({
			statusCode: 400,
			message: "childId and parentId are required",
		});
	}

	// Verify user has access to both goals
	const [childAccess] = await sql<any[]>`
    SELECT 1 FROM user_goals 
    WHERE user_id = ${user.id} AND goal_id = ${childId}
    LIMIT 1
  `;
	const [parentAccess] = await sql<any[]>`
    SELECT 1 FROM user_goals 
    WHERE user_id = ${user.id} AND goal_id = ${parentId}
    LIMIT 1
  `;

	if (!childAccess || !parentAccess) {
		throw createError({
			statusCode: 403,
			message: "Access denied",
		});
	}

	// Create the relation
	const [relation] = await sql<any[]>`
    INSERT INTO goal_relations (parent_id, child_id, \`order\`, weight)
    VALUES (${parentId}, ${childId}, ${order}, ${weight})
    ON CONFLICT (parent_id, child_id) DO UPDATE SET
      \`order\` = ${order},
      weight = ${weight}
    RETURNING *
  `;

	return relation;
});
