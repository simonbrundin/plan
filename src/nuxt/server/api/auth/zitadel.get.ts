import { eventHandler, getQuery, sendRedirect, getCookie, setCookie, deleteCookie } from "h3";
import { withQuery, base64url } from "ufo";
import { useRuntimeConfig } from "#imports";
import type { H3Event } from "h3";

// Session data stored in a signed cookie
interface SessionData {
  userId: string;
  email?: string;
  name?: string;
  sessionToken: string;
  loggedInAt: number;
}

// Simple session management using signed cookies
function getSession(event: H3Event): SessionData | null {
  const sessionCookie = getCookie(event, 'plan_session');
  if (!sessionCookie) return null;
  
  try {
    // Cookie format: base64(JSON data).signature
    const [dataB64, signature] = sessionCookie.split('.');
    if (!dataB64 || !signature) return null;
    
    // Verify signature (simple HMAC)
    const secret = process.env.NUXT_SESSION_PASSWORD || 'default-secret';
    const expectedSig = base64url.encode(
      Buffer.from(secret + dataB64).toString('base64')
    ).slice(0, 32);
    
    if (signature !== expectedSig) {
      console.warn('Invalid session signature');
      return null;
    }
    
    return JSON.parse(Buffer.from(dataB64, 'base64').toString());
  } catch (e) {
    console.error('Failed to parse session:', e);
    return null;
  }
}

function setSession(event: H3Event, data: SessionData): void {
  const secret = process.env.NUXT_SESSION_PASSWORD || 'default-secret';
  const dataB64 = Buffer.from(JSON.stringify(data)).toString('base64');
  const signature = base64url.encode(
    Buffer.from(secret + dataB64).toString('base64')
  ).slice(0, 32);
  
  setCookie(event, 'plan_session', `${dataB64}.${signature}`, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: '/',
  });
}

// Zitadel OAuth handler using PKCE
export default eventHandler(async (event: H3Event) => {
  const config = useRuntimeConfig(event);
  const zitadelConfig = config.oauth?.zitadel;
  const goApiUrl = config.public.goApiUrl || "http://localhost:8080";
  
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
  
  console.log("Code received, validating state");
  
  if (!state || state !== savedState) {
    console.error("Invalid OAuth state");
    return sendRedirect(event, "/login?error=invalid_state");
  }

  deleteCookie(event, 'oauth_zitadel_state');
  deleteCookie(event, 'oauth_zitadel_verifier');

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
      timeout: 15000,
    });

    if (tokenResponse.error) {
      console.error("Token exchange failed:", tokenResponse);
      return sendRedirect(event, "/login?error=token_failed");
    }

    console.log("Token received, getting user info");

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

    console.log("User info received:", userInfo.sub);

    // Get session token from Go API
    let sessionToken = tokenResponse.access_token;
    try {
      console.log("Getting session from Go API");
      const sessionResponse = await $fetch<{
        session: string;
        user_id: string;
      }>(`${goApiUrl}/auth/session`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${tokenResponse.access_token}`,
        },
        timeout: 10000,
      });
      sessionToken = sessionResponse.session;
      console.log("Got session token from Go API");
    } catch (apiError) {
      console.warn("Failed to get session from Go API, using Zitadel token:", apiError);
    }

    // Set session using our custom session management
    const sessionData: SessionData = {
      userId: userInfo.sub,
      email: userInfo.email,
      name: userInfo.name || `${userInfo.given_name || ''} ${userInfo.family_name || ''}`.trim(),
      sessionToken: sessionToken,
      loggedInAt: Date.now(),
    };
    
    console.log("Setting session for user:", sessionData.userId);
    setSession(event, sessionData);

    console.log("Zitadel OAuth success for user:", userInfo.sub);
    return sendRedirect(event, "/");

  } catch (error) {
    console.error("Zitadel OAuth error:", error);
    return sendRedirect(event, "/login?error=auth_failed");
  }
});

async function calculateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}
