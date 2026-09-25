import { getCookie } from "h3";

// Proxy API calls to Go API, attaching session cookie
export default defineEventHandler(async (event) => {
  const sessionCookie = getCookie(event, 'plan_session')
  
  if (!sessionCookie) {
    throw createError({
      statusCode: 401,
      message: 'Not authenticated'
    })
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
  
  // Forward the request with session cookie
  const response = await $fetch(`${goApiUrl}/${path}`, {
    method,
    headers: {
      // Pass session cookie as header (Go API expects X-Session-Token)
      'X-Session-Token': sessionCookie,
      'Content-Type': 'application/json',
    },
    query,
    body,
  })
  
  return response
})
