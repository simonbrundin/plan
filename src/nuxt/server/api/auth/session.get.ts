import { useSession } from "h3";

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

export default defineEventHandler(async (event) => {
  const session = await useSession<SessionData>(event, {
    password: process.env.NUXT_SESSION_PASSWORD || 'default-secret-change-in-production',
    name: 'plan-session',
  });

  const data = await session.data;
  
  if (!data?.user) {
    throw createError({
      statusCode: 401,
      message: 'Not authenticated'
    });
  }
  
  // Return only public data (no access token)
  return {
    user: data.user,
    loggedInAt: data.loggedInAt,
  };
});
