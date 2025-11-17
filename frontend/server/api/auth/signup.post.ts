export default defineEventHandler(async (event) => {
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

  const config = useRuntimeConfig();

  // Hämta Authentik-konfiguration från miljövariabler
  const authentikUrl =
    process.env.NUXT_OAUTH_AUTHENTIK_SERVER_URL || "http://localhost:9000";
  const authentikToken = process.env.AUTHENTIK_API_TOKEN;

  if (!authentikToken) {
    throw createError({
      statusCode: 500,
      message: "Authentik API-token saknas",
    });
  }

  try {
    // Skapa användare via Authentik API
    // Dokumentation: https://docs.goauthentik.io/developer-docs/api/reference#tag/core/operation/core_users_create
    const response = await $fetch(`${authentikUrl}/api/v3/core/users/`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${authentikToken}`,
        "Content-Type": "application/json",
      },
      body: {
        username,
        email,
        name: name || username,
        is_active: true,
        type: "internal",
        // Authentik kan hantera lösenordet på olika sätt
        // Du kan antingen skicka lösenordet direkt eller använda deras flow
        attributes: {
          // Här kan du lägga till custom attribut om du vill
        },
      },
    });

    // Sätt lösenordet separat (om Authentik kräver det)
    // Detta kan variera beroende på din Authentik-konfiguration
    if (response.pk) {
      await $fetch(
        `${authentikUrl}/api/v3/core/users/${response.pk}/set_password/`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${authentikToken}`,
            "Content-Type": "application/json",
          },
          body: {
            password,
          },
        },
      );
    }

    return {
      success: true,
      message: "Användare skapad",
      userId: response.pk,
    };
  } catch (error: any) {
    console.error("Authentik API error:", error);

    // Hantera specifika fel från Authentik
    if (error.statusCode === 400) {
      const errorData = error.data;
      if (errorData?.username) {
        throw createError({
          statusCode: 400,
          message: "Användarnamnet är redan taget",
        });
      }
      if (errorData?.email) {
        throw createError({
          statusCode: 400,
          message: "E-postadressen används redan",
        });
      }
    }

    throw createError({
      statusCode: error.statusCode || 500,
      message: error.data?.detail || "Kunde inte skapa användare",
    });
  }
});
