import { requireUserSession } from 'nuxt-auth-utils'

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  
  const accessToken = session.secure?.accessToken as string | undefined
  if (!accessToken) {
    throw createError({
      statusCode: 401,
      message: 'No access token'
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
  
  // Forward the request
  const response = await $fetch(`${goApiUrl}/${path}`, {
    method,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    query,
    body,
  })
  
  return response
})
