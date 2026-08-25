import { eventHandler, getQuery, sendRedirect } from "h3";

export default eventHandler(async (event) => {
	const query = getQuery(event);

	const token = query.token as string;
	const sub = query.sub as string;
	const email = query.email as string;

	if (!token) {
		return sendRedirect(event, "/?error=auth_failed");
	}

	// Set user session with JWT from Go API
	// accessToken is stored in user object (encrypted in sealed cookie)
	await setUserSession(event, {
		user: {
			id: sub,
			sub: sub,
			email: email,
			accessToken: token,
		},
		loggedInAt: Number(Date.now()),
	});

	return sendRedirect(event, "/");
});
