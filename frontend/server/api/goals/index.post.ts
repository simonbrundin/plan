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

	if (!body.title) {
		throw createError({
			statusCode: 400,
			message: "Title is required",
		});
	}

	// Create the goal
	const [newGoal] = await sql<any[]>`
    INSERT INTO goals (title, icon) 
    VALUES (${body.title}, ${body.icon || "heroicons:star"})
    RETURNING *
  `;

	// Link it to the user
	await sql`
    INSERT INTO user_goals (user_id, goal_id) 
    VALUES (${user.id}, ${newGoal.id})
  `;

	return newGoal;
});
