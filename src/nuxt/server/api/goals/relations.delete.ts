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
	const { childId, parentId } = body;

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

	// Delete the relation
	await sql`
    DELETE FROM goal_relations 
    WHERE parent_id = ${parentId} AND child_id = ${childId}
  `;

	return { deleted: true };
});
