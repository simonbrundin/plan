import { eventHandler, getQuery, sendRedirect, useSession } from "h3";
import { withQuery } from "ufo";

// Session type
interface SessionData {
  user: {
    id: string;
    email?: string;
    name?: string;
  };
  secure: {
    accessToken: string;
    refreshToken?: string;
    expiresAt: number;
  };
  loggedInAt: number;
}

// Zitadel OAuth handler using PKCE
export default eventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const zitadelConfig = config.oauth?.zitadel;
  
  if (!zitadelConfig?.clientId || !zitadelConfig?.domain) {
    console.error("Zitadel OAuth not configured");
    return sendRedirect(event, "/login?error=missing_config");
  }

  const query = getQuery(event);
  
  // If no code, redirect to Zitadel login
  if (!query.code) {
    const state = crypto.randomUUID();
    const verifier = crypto.randomUUID().replace(/-/g, '') + crypto.randomUUID().replace(/-/g, '');
    
    setCookie(event, 'oauth_zitadel_state', state, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 600,
      path: '/'
    });
    setCookie(event, 'oauth_zitadel_verifier', verifier, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 600,
      path: '/'
    });

    // Calculate PKCE code challenge
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);
    const digest = await crypto.subtle.digest('SHA-256', data);
    const codeChallenge = btoa(String.fromCharCode(...new Uint8Array(digest)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
    
    const authorizationURL = `https://${zitadelConfig.domain}/oauth/v2/authorize`;
    const redirectUrl = zitadelConfig.redirectUrl || `${config.public.appUrl}/auth/zitadel`;
    
    return sendRedirect(event, withQuery(authorizationURL, {
      response_type: 'code',
      client_id: zitadelConfig.clientId,
      redirect_uri: redirectUrl,
      scope: 'openid email profile offline_access',
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

  deleteCookie(event, 'oauth_zitadel_state');
  deleteCookie(event, 'oauth_zitadel_verifier');

  const tokenURL = `https://${zitadelConfig.domain}/oauth/v2/token`;
  const redirectUrl = zitadelConfig.redirectUrl || `${config.public.appUrl}/auth/zitadel`;
  
  try {
    // Exchange code for tokens
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
      timeout: 15000,
    });

    if (tokenResponse.error) {
      console.error("Token exchange failed:", tokenResponse);
      return sendRedirect(event, "/login?error=token_failed");
    }

    // Get user info
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
      timeout: 10000,
    });

    // Set session using H3 useSession
    const session = await useSession<SessionData>(event, {
      password: process.env.NUXT_SESSION_PASSWORD || 'default-secret-change-in-production',
      name: 'plan-session',
      maxAge: 60 * 60 * 24 * 7, // 7 days
    });

    await session.update({
      user: {
        id: userInfo.sub,
        email: userInfo.email,
        name: userInfo.name || `${userInfo.given_name || ''} ${userInfo.family_name || ''}`.trim(),
      },
      secure: {
        accessToken: tokenResponse.access_token,
        refreshToken: tokenResponse.refresh_token,
        expiresAt: Date.now() + (tokenResponse.expires_in * 1000),
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
