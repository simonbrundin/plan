import { eventHandler, getCookie } from "h3";
import { base64url } from "ufo";

interface SessionData {
  userId: string;
  email?: string;
  name?: string;
  sessionToken: string;
  loggedInAt: number;
}

function getSession(event: any): SessionData | null {
  const sessionCookie = getCookie(event, 'plan_session');
  if (!sessionCookie) return null;
  
  try {
    const [dataB64, signature] = sessionCookie.split('.');
    if (!dataB64 || !signature) return null;
    
    const secret = process.env.NUXT_SESSION_PASSWORD || 'default-secret';
    const expectedSig = base64url.encode(
      Buffer.from(secret + dataB64).toString('base64')
    ).slice(0, 32);
    
    if (signature !== expectedSig) {
      return null;
    }
    
    return JSON.parse(Buffer.from(dataB64, 'base64').toString());
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
