import { sql } from "../../utils/db";
import { eventHandler, getQuery, sendRedirect, setCookie, getCookie } from "h3";
import { withQuery } from "ufo";
import crypto from "node:crypto";

function generateRandomString(length: number = 32): string {
	return crypto.randomBytes(length).toString("base64url").slice(0, length);
}

async function generateCodeChallenge(verifier: string): Promise<string> {
	const encoder = new TextEncoder();
	const data = encoder.encode(verifier);
	const digest = await crypto.subtle.digest("SHA-256", data);
	return Buffer.from(digest).toString("base64url");
}

export default eventHandler(async (event) => {
	const config = useRuntimeConfig();
	const query = getQuery(event);

	const ZITADEL_DOMAIN = config.oauth.zitadel.domain;
	const CLIENT_ID = config.oauth.zitadel.clientId;
	const REDIRECT_URL =
		config.oauth.zitadel.redirectUrl ||
		config.public.appUrl + "/api/auth/zitadel";

	try {
		const parsed = new URL(REDIRECT_URL);
		if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
			throw new Error("protocol not http/https");
		}
	} catch {
		throw new Error(
			`REDIRECT_URL must be an absolute http(s) URL (got: "${REDIRECT_URL}"). ` +
				"Set NUXT_OAUTH_ZITADEL_REDIRECT_URL or NUXT_PUBLIC_APP_URL.",
		);
	}

	// Check for OAuth errors from Zitadel
	if (query.error) {
		return sendRedirect(
			event,
			`/?error=${query.error}&desc=${query.error_description || ""}`,
		);
	}

	// Step 1: No code yet - initiate OAuth
	if (!query.code) {
		const verifier = generateRandomString(64);
		const challenge = await generateCodeChallenge(verifier);
		const state = generateRandomString(32);

		const combinedData = JSON.stringify({ state, verifier });
		setCookie(event, "oauth_data", combinedData, {
			httpOnly: true,
			secure: true,
			sameSite: "lax",
			maxAge: 60 * 10,
			path: "/",
		});

		return sendRedirect(
			event,
			withQuery(`https://${ZITADEL_DOMAIN}/oauth/v2/authorize`, {
				response_type: "code",
				client_id: CLIENT_ID,
				redirect_uri: REDIRECT_URL,
				scope: "openid email profile",
				state,
				code_challenge: challenge,
				code_challenge_method: "S256",
			}),
		);
	}

	// Step 2: Callback with code
	const combinedData = getCookie(event, "oauth_data");
	if (!combinedData) {
		return sendRedirect(event, "/?error=missing_cookie");
	}

	let oauthData: { state: string; verifier: string };
	try {
		oauthData = JSON.parse(combinedData);
	} catch {
		return sendRedirect(event, "/?error=invalid_cookie");
	}

	if (query.state !== oauthData.state) {
		return sendRedirect(event, "/?error=state_mismatch");
	}

	setCookie(event, "oauth_data", "", {
		maxAge: 0,
		path: "/",
	});

	try {
		const tokenResponse = await $fetch(
			`https://${ZITADEL_DOMAIN}/oauth/v2/token`,
			{
				method: "POST",
				headers: {
					"Content-Type": "application/x-www-form-urlencoded",
				},
				body: new URLSearchParams({
					grant_type: "authorization_code",
					client_id: CLIENT_ID,
					redirect_uri: REDIRECT_URL,
					code: query.code as string,
					code_verifier: oauthData.verifier,
				}).toString(),
			},
		);

		if ((tokenResponse as any).error) {
			const error = tokenResponse as any;
			return sendRedirect(
				event,
				`/?error=token_failed&desc=${encodeURIComponent(error.error_description || error.error)}`,
			);
		}

		const tokens = tokenResponse as {
			access_token: string;
			token_type: string;
			id_token?: string;
		};

		const userInfo = await $fetch(
			`https://${ZITADEL_DOMAIN}/oidc/v1/userinfo`,
			{
				headers: {
					Authorization: `Bearer ${tokens.access_token}`,
					Accept: "application/json",
				},
			},
		);

		const user = userInfo as {
			sub: string;
			email: string;
			preferred_username?: string;
		};

		if (!user?.sub || !user?.email) {
			return sendRedirect(event, "/?error=auth_data_missing");
		}

		// Find or create user in database
		let dbUser = await sql<
			any[]
		>`SELECT id, sub, email FROM users WHERE sub = ${user.sub}`;

		if (!dbUser[0]) {
			const existingByEmail = await sql<
				any[]
			>`SELECT id, sub, email FROM users WHERE email = ${user.email}`;
			if (existingByEmail[0]) {
				await sql`UPDATE users SET sub = ${user.sub} WHERE id = ${existingByEmail[0].id}`;
				dbUser = [
					{
						id: existingByEmail[0].id,
						sub: user.sub,
						email: existingByEmail[0].email,
					},
				] as any;
			}
		}

		if (!dbUser[0]) {
			const newUsers = await sql<
				any[]
			>`INSERT INTO users (sub, email) VALUES (${user.sub}, ${user.email}) RETURNING id, sub, email`;
			if (!newUsers[0]) {
				return sendRedirect(event, "/?error=user_creation_failed");
			}
			dbUser = newUsers;
		}

		const finalDbUser = dbUser[0];

		// Ensure user has root goal (ID 1)
		const existingRootGoal = await sql<
			any[]
		>`SELECT 1 FROM user_goals WHERE user_id = ${finalDbUser.id} LIMIT 1`;
		if (!existingRootGoal[0]) {
			await sql`INSERT INTO user_goals (user_id, goal_id) VALUES (${finalDbUser.id}, 1) ON CONFLICT DO NOTHING`;
		}

		await setUserSession(event, {
			user: {
				id: finalDbUser.id,
				sub: String(finalDbUser.sub),
				email: String(finalDbUser.email),
			},
			accessToken: tokens.access_token,
			loggedInAt: Number(Date.now()),
		});

		return sendRedirect(event, "/");
	} catch (error: any) {
		return sendRedirect(event, "/?error=auth_failed");
	}
});
