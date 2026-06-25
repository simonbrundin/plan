export default defineEventHandler(async (event) => {
	const config = useRuntimeConfig();
	const body = await readBody(event);

	const { email, password, username, name } = body;

	// Validering
	if (!email || !password || !username) {
		throw createError({
			statusCode: 400,
			message: "E-post, lösenord och användarnamn krävs",
		});
	}

	if (password.length < 8) {
		throw createError({
			statusCode: 400,
			message: "Lösenordet måste vara minst 8 tecken",
		});
	}

	const zitadelDomain = config.oauth.zitadel.domain;
	const zitadelToken = process.env.ZITADEL_API_TOKEN;

	if (!zitadelToken) {
		throw createError({
			statusCode: 500,
			message: "Zitadel API-token saknas",
		});
	}

	try {
		// Skapa användare via Zitadel Admin API
		const response = await $fetch(`https://${zitadelDomain}/admin/v1/users`, {
			method: "POST",
			headers: {
				Authorization: `Bearer ${zitadelToken}`,
				"Content-Type": "application/json",
			},
			body: {
				userType: {
					human: {
						email: {
							email: email,
							isVerified: false,
						},
						profile: {
							firstName: name || username,
							lastName: "",
						},
						userName: username,
					},
				},
			},
		});

		return {
			success: true,
			message: "Användare skapad",
			userId: (response as any).user?.id,
		};
	} catch (error: any) {
		console.error("Zitadel API error:", error);

		throw createError({
			statusCode: error.statusCode || 500,
			message: error.data?.message || "Kunde inte skapa användare",
		});
	}
});
