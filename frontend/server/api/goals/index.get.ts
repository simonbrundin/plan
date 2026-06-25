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

	const goals = await sql<any[]>`
    SELECT g.* FROM goals g
    INNER JOIN user_goals ug ON g.id = ug.goal_id
    WHERE ug.user_id = ${user.id}
    ORDER BY g.created DESC
  `;

	return goals;
});
