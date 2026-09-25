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
  // Get session
  const session = await useSession<SessionData>(event, {
    password: process.env.NUXT_SESSION_PASSWORD || 'default-secret-change-in-production',
    name: 'plan-session',
  });

  const user = await session.data;
  
  if (!user?.secure?.accessToken) {
    throw createError({
      statusCode: 401,
      message: 'Not authenticated'
    });
  }
  
  // Get the path from the URL
  const path = getRouterParam(event, 'path')
  if (!path) {
    throw createError({
      statusCode: 400,
      message: 'Missing path'
    })
  }
  
  const config = useRuntimeConfig()
  const goApiUrl = config.public.goApiUrl || 'http://localhost:8080'
  
  // Get query params
  const query = getQuery(event)
  
  // Get method and body
  const method = getMethod(event)
  const body = method !== 'GET' ? await readBody(event) : undefined
  
  // Forward the request with access token
  const response = await $fetch(`${goApiUrl}/${path}`, {
    method,
    headers: {
      Authorization: `Bearer ${user.secure.accessToken}`,
      'Content-Type': 'application/json',
    },
    query,
    body,
  })
  
  return response
})
