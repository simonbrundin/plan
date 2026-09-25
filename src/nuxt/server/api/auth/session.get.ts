import { eventHandler, getCookie } from "h3";

interface SessionData {
  userId: string;
  email?: string;
  name?: string;
  sessionToken: string;
  loggedInAt: number;
}

// Simple base64url encoding/decoding
function base64urlEncode(str: string): string {
  return Buffer.from(str).toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

function base64urlDecode(str: string): string {
  str = str.replace(/-/g, '+').replace(/_/g, '/');
  while (str.length % 4) str += '=';
  return Buffer.from(str, 'base64').toString();
}

function getSession(event: any): SessionData | null {
  const sessionCookie = getCookie(event, 'plan_session');
  if (!sessionCookie) return null;
  
  try {
    const [dataB64, signature] = sessionCookie.split('.');
    if (!dataB64 || !signature) return null;
    
    const secret = process.env.NUXT_SESSION_PASSWORD || 'default-secret';
    const expectedSig = base64urlEncode(
      Buffer.from(secret + dataB64).toString('base64')
    ).slice(0, 32);
    
    if (signature !== expectedSig) {
      return null;
    }
    
    return JSON.parse(base64urlDecode(dataB64));
  } catch (e) {
    return null;
  }
}

export default eventHandler(async (event) => {
  const session = getSession(event);
  
  if (!session) {
    throw createError({
      statusCode: 401,
      message: "Not authenticated"
    });
  }
  
  return session;
});
