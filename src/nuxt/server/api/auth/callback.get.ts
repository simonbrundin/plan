import { eventHandler, getQuery, sendRedirect, getCookie } from "h3";

export default eventHandler(async (event) => {
	const query = getQuery(event);

	const sessionToken = query.session as string;
	const sub = query.sub as string;
	const email = query.email as string;

	console.log("OAuth callback:", { session: sessionToken?.substring(0, 50), sub, email });

	if (!sessionToken) {
		console.error("No session in callback");
		return sendRedirect(event, "/?error=auth_failed");
	}

	// Set user session with our own session token (not the Zitadel token)
	const sessionData = {
		user: {
			id: sub,
			sub: sub,
			email: email,
			sessionToken: sessionToken, // Our own session token
		},
		loggedInAt: Number(Date.now()),
	};

	console.log("Setting session with data:", JSON.stringify(sessionData, null, 2));

	await setUserSession(event, sessionData);

	// Verify session was set
	const savedSession = await getUserSession(event);
	console.log("Session after setUserSession:", JSON.stringify(savedSession, null, 2));

	return sendRedirect(event, "/");
});
