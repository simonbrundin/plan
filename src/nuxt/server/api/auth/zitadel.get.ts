import { eventHandler, getQuery, sendRedirect } from "h3";
import { withQuery } from "ufo";
import { useRuntimeConfig } from "#imports";
import type { H3Event } from "h3";

// Zitadel OAuth handler using PKCE
// Exchanges Zitadel tokens for a longer-lived session token from Go API
export default eventHandler(async (event: H3Event) => {
  const config = useRuntimeConfig(event);
  const zitadelConfig = config.oauth?.zitadel;
  const goApiUrl = config.public.goApiUrl || "http://localhost:8080";
  
  if (!zitadelConfig?.clientId || !zitadelConfig?.domain) {
    console.error("Zitadel OAuth not configured");
    return sendRedirect(event, "/login?error=missing_config");
  }

  const query = getQuery(event);
  
  // If no code, redirect to Zitadel login
  if (!query.code) {
    const state = crypto.randomUUID();
    const verifier = crypto.randomUUID().replace(/-/g, '') + crypto.randomUUID().replace(/-/g, '');
    
    // Store state and verifier in cookies
    setCookie(event, 'oauth_zitadel_state', state, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 600, // 10 minutes
      path: '/'
    });
    setCookie(event, 'oauth_zitadel_verifier', verifier, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 600,
      path: '/'
    });

    // Calculate code challenge for PKCE
    const codeChallenge = await calculateCodeChallenge(verifier);
    
    const authorizationURL = `https://${zitadelConfig.domain}/oauth/v2/authorize`;
    const redirectUrl = zitadelConfig.redirectUrl || `${config.public.appUrl}/api/auth/zitadel`;
    
    return sendRedirect(event, withQuery(authorizationURL, {
      response_type: 'code',
      client_id: zitadelConfig.clientId,
      redirect_uri: redirectUrl,
      scope: 'openid email profile',
      state,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    }));
  }

  // Validate state
  const state = query.state as string;
  const savedState = getCookie(event, 'oauth_zitadel_state');
  const verifier = getCookie(event, 'oauth_zitadel_verifier');
  
  if (!state || state !== savedState) {
    console.error("Invalid OAuth state");
    return sendRedirect(event, "/login?error=invalid_state");
  }

  // Clear state cookies
  deleteCookie(event, 'oauth_zitadel_state');
  deleteCookie(event, 'oauth_zitadel_verifier');

  // Exchange code for tokens
  const tokenURL = `https://${zitadelConfig.domain}/oauth/v2/token`;
  const redirectUrl = zitadelConfig.redirectUrl || `${config.public.appUrl}/api/auth/zitadel`;
  
  try {
    const tokenResponse = await $fetch<{
      access_token: string;
      refresh_token?: string;
      id_token?: string;
      token_type: string;
      expires_in: number;
      error?: string;
    }>(tokenURL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        client_id: zitadelConfig.clientId,
        redirect_uri: redirectUrl,
        code: query.code as string,
        code_verifier: verifier || '',
      }).toString(),
    });

    if (tokenResponse.error) {
      console.error("Token exchange failed:", tokenResponse);
      return sendRedirect(event, "/login?error=token_failed");
    }

    // Get user info using access token
    const userInfoURL = `https://${zitadelConfig.domain}/oidc/v1/userinfo`;
    const userInfo = await $fetch<{
      sub: string;
      email?: string;
      name?: string;
      given_name?: string;
      family_name?: string;
    }>(userInfoURL, {
      headers: {
        Authorization: `Bearer ${tokenResponse.access_token}`,
        Accept: 'application/json'
      }
    });

    // Exchange Zitadel token for a longer-lived session token from Go API
    let sessionToken = tokenResponse.access_token;
    try {
      const sessionResponse = await $fetch<{
        session: string;
        user_id: string;
      }>(`${goApiUrl}/auth/session`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${tokenResponse.access_token}`,
        },
      });
      sessionToken = sessionResponse.session;
      console.log("Got session token from Go API");
    } catch (apiError) {
      console.warn("Failed to get session from Go API, using Zitadel token directly:", apiError);
      // Fall back to using Zitadel token
    }

    // Set user session with the session token
    await setUserSession(event, {
      user: {
        id: userInfo.sub,
        sub: userInfo.sub,
        email: userInfo.email,
        name: userInfo.name || `${userInfo.given_name || ''} ${userInfo.family_name || ''}`.trim(),
        sessionToken: sessionToken,
        refreshToken: tokenResponse.refresh_token,
      },
      loggedInAt: Date.now(),
    });

    console.log("Zitadel OAuth success for user:", userInfo.sub);
    return sendRedirect(event, "/");

  } catch (error) {
    console.error("Zitadel OAuth error:", error);
    return sendRedirect(event, "/login?error=auth_failed");
  }
});

// Calculate PKCE code challenge
async function calculateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}
