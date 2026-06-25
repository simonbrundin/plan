import { db } from "../../utils/db";
import { users } from "../../../drizzle/schema";
import { eq } from "drizzle-orm";
import { eventHandler, getQuery, sendRedirect, setCookie, getCookie } from "h3";
import { withQuery } from "ufo";
import crypto from "node:crypto";

const ZITADEL_DOMAIN =
	process.env.NUXT_OAUTH_ZITADEL_DOMAIN || "auth.simonbrundin.com";
const CLIENT_ID =
	process.env.NUXT_OAUTH_ZITADEL_CLIENT_ID || "378094068590182893";
const REDIRECT_URL =
	process.env.NUXT_OAUTH_ZITADEL_REDIRECT_URL ||
	"http://localhost:3000/api/auth/zitadel";

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
	const query = getQuery(event);

	console.log("=== ZITADEL HANDLER ===");

	// Check for OAuth errors from Zitadel
	if (query.error) {
		console.error("OAuth error from Zitadel:", query);
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

		console.log("=== INITIATING OAuth ===");
		console.log("Verifier length:", verifier.length);
		console.log("Challenge length:", challenge.length);

		// Store both state and verifier in a single cookie
		const combinedData = JSON.stringify({ state, verifier });
		setCookie(event, "oauth_data", combinedData, {
			httpOnly: true,
			secure: false,
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
	console.log("=== CALLBACK ===");

	const combinedData = getCookie(event, "oauth_data");
	if (!combinedData) {
		console.error("Missing oauth_data cookie!");
		return sendRedirect(event, "/?error=missing_cookie");
	}

	let oauthData: { state: string; verifier: string };
	try {
		oauthData = JSON.parse(combinedData);
	} catch {
		console.error("Invalid oauth_data cookie!");
		return sendRedirect(event, "/?error=invalid_cookie");
	}

	console.log("State from query:", query.state);
	console.log("State from cookie:", oauthData.state);

	if (query.state !== oauthData.state) {
		console.error("State mismatch!");
		return sendRedirect(event, "/?error=state_mismatch");
	}

	// Clear the cookie
	setCookie(event, "oauth_data", "", {
		maxAge: 0,
		path: "/",
	});

	console.log("=== TOKEN REQUEST ===");

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

		console.log("Token response:", JSON.stringify(tokenResponse, null, 2));

		if ((tokenResponse as any).error) {
			const error = tokenResponse as any;
			console.error("Token error:", error.error, error.error_description);
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
		const accessToken = tokens.access_token;

		// Get user info
		const userInfo = await $fetch(
			`https://${ZITADEL_DOMAIN}/oidc/v1/userinfo`,
			{
				headers: {
					Authorization: `Bearer ${accessToken}`,
					Accept: "application/json",
				},
			},
		);

		console.log("User info:", JSON.stringify(userInfo, null, 2));

		const user = userInfo as {
			sub: string;
			email: string;
			preferred_username?: string;
		};

		if (!user?.sub || !user?.email) {
			console.error("Missing user data:", { user });
			return sendRedirect(event, "/?error=auth_data_missing");
		}

		// Find or create user in database
		let dbUser = await db.query.users.findFirst({
			where: eq(users.sub, user.sub),
		});

		if (!dbUser) {
			dbUser = await db.query.users.findFirst({
				where: eq(users.email, user.email),
			});
			if (dbUser) {
				console.log("Found user by email:", dbUser.email);
				await db
					.update(users)
					.set({ sub: user.sub })
					.where(eq(users.id, dbUser.id));
				dbUser.sub = user.sub;
			}
		}

		if (!dbUser) {
			console.log("Creating new user:", user.email);
			dbUser = await db
				.insert(users)
				.values({
					sub: user.sub,
					email: user.email,
				})
				.returning()
				.then(([newUser]) => newUser);
		}

		if (!dbUser) {
			console.error("Failed to find or create user");
			return sendRedirect(event, "/?error=user_creation_failed");
		}

		await setUserSession(event, {
			user: {
				id: Number(dbUser.id),
				sub: String(dbUser.sub),
				email: String(dbUser.email),
			},
			loggedInAt: Number(Date.now()),
		});

		console.log(`Success: ${dbUser.email}`);
		return sendRedirect(event, "/");
	} catch (error: any) {
		console.error("=== FATAL ERROR ===");
		console.error(error);

		if (error?.data) {
			console.error("Error data:", error.data);
		}

		return sendRedirect(event, "/?error=auth_failed");
	}
});
