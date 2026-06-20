import { db } from "../../utils/db";
import { users } from "../../../drizzle/schema";
import { eq } from "drizzle-orm";
import {
	eventHandler,
	getQuery,
	sendRedirect,
	setCookie,
	getCookie,
	createError,
} from "h3";
import { withQuery } from "ufo";

const ZITADEL_DOMAIN =
	process.env.NUXT_OAUTH_ZITADEL_DOMAIN || "auth.simonbrundin.com";
const CLIENT_ID =
	process.env.NUXT_OAUTH_ZITADEL_CLIENT_ID || "378094068590182893";
const REDIRECT_URI = "http://localhost:3000/auth/zitadel";
const AUTHORIZATION_URL = `https://${ZITADEL_DOMAIN}/oauth/v2/authorize`;
const TOKEN_URL = `https://${ZITADEL_DOMAIN}/oauth/v2/token`;

function generateRandomString(length: number = 32): string {
	const chars =
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
	const array = new Uint8Array(length);
	crypto.getRandomValues(array);
	return Array.from(array, (byte) => chars[byte % chars.length]).join("");
}

async function generateCodeChallenge(verifier: string): Promise<string> {
	const encoder = new TextEncoder();
	const data = encoder.encode(verifier);
	const digest = await crypto.subtle.digest("SHA-256", data);
	return btoa(String.fromCharCode(...new Uint8Array(digest)))
		.replace(/\+/g, "-")
		.replace(/\//g, "_")
		.replace(/=+$/, "");
}

export default eventHandler(async (event) => {
	const query = getQuery(event);

	console.log("=== ZITADEL HANDLER ===");
	console.log("Query:", query);

	// Check for OAuth errors from Zitadel
	if (query.error) {
		console.error("OAuth error from Zitadel:", query);
		return sendRedirect(
			event,
			`/?error=${query.error}&desc=${query.error_description || ""}`,
		);
	}

	// Step 1: If no code, redirect to Zitadel for authorization
	if (!query.code) {
		const verifier = generateRandomString(64);
		const challenge = await generateCodeChallenge(verifier);
		const state = generateRandomString(32);

		console.log("=== INITIATING OAuth ===");
		console.log("Verifier:", verifier.substring(0, 10) + "...");
		console.log("Challenge:", challenge.substring(0, 10) + "...");
		console.log("State:", state);

		// Store combined state + verifier in ONE cookie
		const combinedValue = `${state}:${verifier}`;
		const cookieOptions = {
			httpOnly: true,
			secure: false,
			sameSite: "lax" as const,
			maxAge: 60 * 10,
			path: "/",
		};

		setCookie(event, "auth_oauth_data", combinedValue, cookieOptions);
		console.log("Combined cookie set, length:", combinedValue.length);

		console.log("=== COOKIES SET ===");
		console.log("Verifier cookie set:", verifier);
		console.log("State cookie set:", state);

		return sendRedirect(
			event,
			withQuery(AUTHORIZATION_URL, {
				response_type: "code",
				client_id: CLIENT_ID,
				redirect_uri: REDIRECT_URI,
				scope: "openid email profile",
				state: state,
				code_challenge: challenge,
				code_challenge_method: "S256",
			}),
		);
	}

	// Step 2: Callback with code - validate state
	const allCookies = event.node.req.headers.cookie;
	console.log("=== ALL REQUEST COOKIES ===");
	console.log(allCookies);
	// Read combined cookie and split
	const combinedCookie = getCookie(event, "auth_oauth_data");
	if (!combinedCookie) {
		console.error("Missing combined auth cookie!");
		return sendRedirect(event, "/?error=missing_cookie");
	}
	const [savedState, verifier] = combinedCookie.split(":");

	console.log("=== CALLBACK ===");
	console.log("Query state:", query.state);
	console.log("Saved state:", savedState);
	console.log("Saved verifier:", verifier ? "exists" : "missing");

	if (!savedState || query.state !== savedState) {
		console.error("State mismatch!");
		return sendRedirect(event, "/?error=state_mismatch");
	}

	if (!verifier) {
		console.error("Missing PKCE verifier!");
		return sendRedirect(event, "/?error=missing_verifier");
	}

	// Clear cookies immediately
	setCookie(event, "oauth_pkce_state", "", {
		maxAge: 0,
		path: "/auth/zitadel",
	});
	setCookie(event, "oauth_pkce_verifier", "", {
		maxAge: 0,
		path: "/auth/zitadel",
	});

	// Exchange code for tokens
	console.log("=== TOKEN REQUEST ===");
	console.log("Token URL:", TOKEN_URL);
	console.log("Redirect URI:", REDIRECT_URI);
	console.log("Code:", (query.code as string).substring(0, 20) + "...");

	try {
		const tokenResponse = await $fetch(TOKEN_URL, {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
			},
			body: new URLSearchParams({
				grant_type: "authorization_code",
				client_id: CLIENT_ID,
				redirect_uri: REDIRECT_URI,
				code: query.code as string,
				code_verifier: verifier,
			}).toString(),
		});

		console.log("=== TOKEN RESPONSE ===");
		console.log(JSON.stringify(tokenResponse, null, 2));

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

		// Get user info from Zitadel OIDC endpoint
		const userInfo = await $fetch(
			`https://${ZITADEL_DOMAIN}/oidc/v1/userinfo`,
			{
				headers: {
					Authorization: `Bearer ${accessToken}`,
					Accept: "application/json",
				},
			},
		);

		console.log("=== USER INFO ===");
		console.log(JSON.stringify(userInfo, null, 2));

		const user = userInfo as {
			sub: string;
			email: string;
			preferred_username?: string;
		};

		if (!user?.sub || !user?.email) {
			console.error("Zitadel OAuth: Missing required user data", { user });
			return sendRedirect(event, "/?error=auth_data_missing");
		}

		// Find or create user in database
		let dbUser = await db.query.users.findFirst({
			where: eq(users.sub, user.sub),
		});

		// If not found by sub, try to find by email (in case user was created with different sub)
		if (!dbUser) {
			dbUser = await db.query.users.findFirst({
				where: eq(users.email, user.email),
			});
			if (dbUser) {
				console.log("Found user by email, updating sub:", dbUser.email);
				// Update the sub to match Zitadel's sub
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
			console.error("Zitadel OAuth: Failed to find or create user");
			return sendRedirect(event, "/?error=user_creation_failed");
		}

		// Set session using nuxt-auth-utils
		await setUserSession(event, {
			user: {
				id: Number(dbUser.id),
				sub: String(dbUser.sub),
				email: String(dbUser.email),
			},
			loggedInAt: Number(Date.now()),
		});

		console.log(
			`Zitadel OAuth: Successfully authenticated user ${dbUser.email}`,
		);
		return sendRedirect(event, "/");
	} catch (error: any) {
		console.error("=== FATAL ERROR ===");
		console.error(error);

		// Försök extrahera Zitadel-fel
		if (error?.data) {
			console.error("Zitadel error:", error.data);
		}

		return sendRedirect(event, "/?error=auth_failed");
	}
});
