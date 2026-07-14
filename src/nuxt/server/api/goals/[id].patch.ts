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
	if (body.started !== undefined) {
		fields.push(`started = $${paramIndex++}`);
		values.push(body.started ? new Date(body.started) : null);
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

	try {
		await sql.unsafe(query, values);
	} catch (err: any) {
		// Ignorera errors för started-kolumnen — den finns inte före migrering.
		// Viktigt: om andra fält också uppdateras, försök igen utan started.
		const msg = err?.message ?? "";
		const isMissingStartedColumn =
			(err?.code === "42703" || msg.includes("does not exist")) &&
			msg.includes("started");

		if (!isMissingStartedColumn) throw err;

		if (fields.length === 1) {
			const [existing] = await sql<any[]>`SELECT * FROM goals WHERE id = ${goalId}`;
			return existing;
		}

		const fallbackFields: string[] = [];
		const fallbackValues: any[] = [];
		let i = 1;

		if (body.title !== undefined) {
			fallbackFields.push(`title = $${i++}`);
			fallbackValues.push(body.title);
		}
		if (body.icon !== undefined) {
			fallbackFields.push(`icon = $${i++}`);
			fallbackValues.push(body.icon);
		}
		if (body.finished !== undefined) {
			fallbackFields.push(`finished = $${i++}`);
			fallbackValues.push(body.finished ? new Date(body.finished) : null);
		}

		const fallbackQuery = `UPDATE goals SET ${fallbackFields.join(", ")} WHERE id = $${i}`;
		fallbackValues.push(goalId);
		await sql.unsafe(fallbackQuery, fallbackValues);
	}

	const [updatedGoal] = await sql<
		any[]
	>`SELECT * FROM goals WHERE id = ${goalId}`;

	return updatedGoal;
});
