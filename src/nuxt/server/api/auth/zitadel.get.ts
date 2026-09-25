import { eventHandler, getQuery, sendRedirect } from "h3";
import { withQuery } from "ufo";
import { useRuntimeConfig } from "#imports";
import type { H3Event } from "h3";

// Zitadel OAuth handler using PKCE
// Stores Zitadel tokens with expiry time for automatic refresh
export default eventHandler(async (event: H3Event) => {
  const config = useRuntimeConfig(event);
  const zitadelConfig = config.oauth?.zitadel;
  
  console.log("Zitadel OAuth callback started");
  
  if (!zitadelConfig?.clientId || !zitadelConfig?.domain) {
    console.error("Zitadel OAuth not configured");
    return sendRedirect(event, "/login?error=missing_config");
  }

  const query = getQuery(event);
  
  // If no code, redirect to Zitadel login
  if (!query.code) {
    console.log("No code in callback, redirecting to Zitadel login");
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
      scope: 'openid email profile offline_access', // offline_access for refresh token
      state,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    }));
  }

  // Validate state
  const state = query.state as string;
  const savedState = getCookie(event, 'oauth_zitadel_state');
  const verifier = getCookie(event, 'oauth_zitadel_verifier');
  
  console.log("Code received, validating state");
  
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
  
  console.log("Exchanging code for tokens");
  
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
      timeout: 15000, // 15 second timeout
    });

    if (tokenResponse.error) {
      console.error("Token exchange failed:", tokenResponse);
      return sendRedirect(event, "/login?error=token_failed");
    }

    console.log("Token received, getting user info");

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
      },
      timeout: 10000, // 10 second timeout
    });

    console.log("User info received:", userInfo.sub);

    // Calculate when the token expires
    const expiresAt = Date.now() + (tokenResponse.expires_in * 1000);
    
    // Set user session with Zitadel tokens and expiry info
    const sessionData = {
      user: {
        id: userInfo.sub,
        sub: userInfo.sub,
        email: userInfo.email,
        name: userInfo.name || `${userInfo.given_name || ''} ${userInfo.family_name || ''}`.trim(),
        accessToken: tokenResponse.access_token,
        refreshToken: tokenResponse.refresh_token,
        expiresAt: expiresAt,
      },
      loggedInAt: Date.now(),
    };
    
    console.log("Setting session with data:", JSON.stringify({...sessionData, user: {...sessionData.user, accessToken: '[REDACTED]'}}));
    await setUserSession(event, sessionData);

    console.log("Zitadel OAuth success for user:", userInfo.sub, "expires in:", tokenResponse.expires_in, "seconds");
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
