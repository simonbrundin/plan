import { eventHandler } from "h3";
import { useRuntimeConfig } from "#imports";

// Refresh Zitadel access token using refresh token
export default eventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const zitadelConfig = config.oauth?.zitadel;
  
  if (!zitadelConfig?.clientId || !zitadelConfig?.domain) {
    throw createError({
      statusCode: 500,
      message: "Zitadel OAuth not configured"
    });
  }

  // Get current session
  const session = await getUserSession(event);
  const refreshToken = session.user?.refreshToken;
  
  if (!refreshToken) {
    throw createError({
      statusCode: 401,
      message: "No refresh token available"
    });
  }

  // Exchange refresh token for new access token
  const tokenURL = `https://${zitadelConfig.domain}/oauth/v2/token`;
  
  try {
    const tokenResponse = await $fetch<{
      access_token: string;
      refresh_token?: string;
      expires_in: number;
      error?: string;
    }>(tokenURL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        grant_type: 'refresh_token',
        client_id: zitadelConfig.clientId,
        refresh_token: refreshToken,
      }).toString(),
    });

    if (tokenResponse.error) {
      console.error("Token refresh failed:", tokenResponse);
      throw createError({
        statusCode: 401,
        message: "Failed to refresh token"
      });
    }

    // Calculate new expiry time
    const expiresAt = Date.now() + (tokenResponse.expires_in * 1000);
    
    // Update session with new tokens
    await setUserSession(event, {
      user: {
        ...session.user,
        accessToken: tokenResponse.access_token,
        refreshToken: tokenResponse.refresh_token || refreshToken, // Use new refresh token if provided
        expiresAt: expiresAt,
      },
      loggedInAt: session.loggedInAt,
    });

    console.log("Token refreshed successfully");
    return { success: true, expiresAt };

  } catch (error) {
    console.error("Token refresh error:", error);
    throw createError({
      statusCode: 401,
      message: "Token refresh failed"
    });
  }
});
