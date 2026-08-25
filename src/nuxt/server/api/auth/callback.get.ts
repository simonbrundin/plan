import { eventHandler, getQuery, sendRedirect, getCookie } from "h3";

export default eventHandler(async (event) => {
	const query = getQuery(event);

	const token = query.token as string;
	const sub = query.sub as string;
	const email = query.email as string;

	console.log("OAuth callback:", { token: token?.substring(0, 50), sub, email });

	if (!token) {
		console.error("No token in callback");
		return sendRedirect(event, "/?error=auth_failed");
	}

	// Set user session with JWT from Go API
	// accessToken is stored in user object (encrypted in sealed cookie)
	const sessionData = {
		user: {
			id: sub,
			sub: sub,
			email: email,
			accessToken: token,
		},
		loggedInAt: Number(Date.now()),
	};

	console.log("Setting session with data:", JSON.stringify(sessionData, null, 2));

	await setUserSession(event, sessionData);

	// Verify session was set
	const session = await getUserSession(event);
	console.log("Session after setUserSession:", JSON.stringify(session, null, 2));

	return sendRedirect(event, "/");
});
