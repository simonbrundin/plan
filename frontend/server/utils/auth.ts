import { sql } from "./db";

export async function getCurrentUser(event: any) {
	const session = await getUserSession(event);

	const userSub = (session?.user as any)?.sub;
	if (!userSub) {
		return null;
	}

	const [user] = await sql<any[]>`
    SELECT * FROM users WHERE sub = ${userSub} LIMIT 1
  `;

	return user || null;
}

export function requireAuth(event: any) {
	const user = event.context.user;
	if (!user) {
		throw createError({
			statusCode: 401,
			message: "Unauthorized",
		});
	}
	return user;
}
