import { eventHandler, sendRedirect, setCookie, getCookie } from "h3";
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
	
	// Go API URL for callback
	const GO_API_URL = config.public.goApiUrl || "http://localhost:8080";
	const CALLBACK_URL = GO_API_URL + "/api/v1/auth/callback";
	
	// Frontend URL to redirect after login
	const FRONTEND_URL = config.public.appUrl + "/api/auth/callback";

	try {
		const parsed = new URL(CALLBACK_URL);
		if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
			throw new Error("protocol not http/https");
		}
	} catch {
		throw createError({
			statusCode: 400,
			message: "Invalid callback URL configuration",
		});
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

		const combinedData = JSON.stringify({ state, verifier, frontendUrl: FRONTEND_URL });
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
				redirect_uri: CALLBACK_URL,
				scope: "openid email profile",
				state,
				code_challenge: challenge,
				code_challenge_method: "S256",
			}),
		);
	}

	// Step 2: Callback from Zitadel with code - redirect to Go API
	console.log("=== CALLBACK WITH CODE ===");

	const combinedData = getCookie(event, "oauth_data");
	if (!combinedData) {
		return sendRedirect(event, "/?error=missing_cookie");
	}

	let oauthData: { state: string; verifier: string; frontendUrl: string };
	try {
		oauthData = JSON.parse(combinedData);
	} catch {
		return sendRedirect(event, "/?error=invalid_cookie");
	}

	if (query.state !== oauthData.state) {
		return sendRedirect(event, "/?error=state_mismatch");
	}

	// Clear the cookie
	setCookie(event, "oauth_data", "", {
		maxAge: 0,
		path: "/",
	});

	// Redirect to Go API callback with the code and state
	return sendRedirect(
		event,
		withQuery(`${GO_API_URL}/api/v1/auth/callback`, {
			code: query.code,
			state: query.state,
		}),
	);
});
