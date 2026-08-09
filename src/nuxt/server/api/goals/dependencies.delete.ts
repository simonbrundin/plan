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

		const body = await readBody(event);
		const { goalId, dependsOnId } = body;

		if (!goalId || !dependsOnId) {
			throw createError({
				statusCode: 400,
				message: "goalId and dependsOnId are required",
			});
		}

		// Verify user has access to the goal
		const [goalAccess] = await sql<any[]>`
			SELECT 1 FROM user_goals
			WHERE user_id = ${user.id} AND goal_id = ${goalId}
			LIMIT 1
		`;

		if (!goalAccess) {
			throw createError({
				statusCode: 403,
				message: "Access denied",
			});
		}

		// Delete the dependency
		await sql`
			DELETE FROM goal_dependencies
			WHERE goal_id = ${goalId} AND depends_on_id = ${dependsOnId}
		`;

		return { deleted: true };
	} catch (err: any) {
		if (err?.statusCode && err.statusCode < 500) {
			throw err;
		}
		console.error("[dependencies.delete] Unexpected error:", err);
		throw createError({
			statusCode: 500,
			message: err?.message ?? "Internal server error",
		});
	}
});
