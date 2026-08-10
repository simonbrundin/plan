import { eventHandler, getQuery, sendRedirect } from "h3";
import { withQuery } from "ufo";

export default eventHandler(async (event) => {
	const config = useRuntimeConfig();
	const query = getQuery(event);

	// Go API URL
	const GO_API_URL = config.public.goApiUrl || "http://localhost:8080";
	const CALLBACK_URL = GO_API_URL + "/api/v1/auth/callback";

	// Check for OAuth errors
	if (query.error) {
		return sendRedirect(
			event,
			`/?error=${query.error}&desc=${query.error_description || ""}`,
		);
	}

	const code = query.code as string;
	const state = query.state as string;

	if (!code || !state) {
		return sendRedirect(event, "/?error=missing_params");
	}

	// Exchange code for token via Go API
	try {
		const callbackURL = withQuery(CALLBACK_URL, { code, state });
		const response = await $fetch<{
			access_token: string;
			user: {
				sub: string;
				email: string;
				name?: string;
				preferred_username?: string;
			};
		}>(callbackURL, {
			method: "GET",
		});

		// Set user session with data from Go API
		await setUserSession(event, {
			user: {
				id: response.user.sub, // Use sub as ID for now
				sub: response.user.sub,
				email: response.user.email,
			},
			accessToken: response.access_token,
			loggedInAt: Number(Date.now()),
		});

		return sendRedirect(event, "/");
	} catch (error: any) {
		console.error("Auth callback error:", error);
		return sendRedirect(event, "/?error=auth_failed");
	}
});
