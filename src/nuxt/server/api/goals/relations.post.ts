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

		// Safely read body — may be null/empty if no body sent
		let body: any = {};
		try {
			body = (await readBody(event)) ?? {};
		} catch (_) {
			body = {};
		}

		const childId = Number(body.childId);
		const parentId = Number(body.parentId);
		const order = body.order !== undefined ? Number(body.order) : 0;
		const weight = body.weight !== undefined ? Number(body.weight) : 10;

		if (
			!childId ||
			!parentId ||
			Number.isNaN(childId) ||
			Number.isNaN(parentId)
		) {
			throw createError({
				statusCode: 400,
				message: "childId and parentId are required and must be numbers",
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
      INSERT INTO goal_relations (parent_id, child_id, "order", weight)
      VALUES (${parentId}, ${childId}, ${order}, ${weight})
      ON CONFLICT (parent_id, child_id) DO UPDATE SET
        "order" = ${order},
        weight = ${weight}
      RETURNING *
    `;

		return relation;
	} catch (err: any) {
		// Re-throw H3 errors (401/400/403) as-is
		if (err?.statusCode && err.statusCode < 500) {
			throw err;
		}
		// Log and convert unexpected errors to 500 with helpful message
		console.error("[relations.post] Unexpected error:", err);
		throw createError({
			statusCode: 500,
			message: err?.message ?? "Internal server error",
		});
	}
});
