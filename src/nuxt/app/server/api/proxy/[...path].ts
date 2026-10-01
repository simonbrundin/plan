import { defineEventHandler, getRequestHeader, createError } from "h3";

// Proxy API calls through the server to attach session cookies
export default defineEventHandler(async (event) => {
  const path = getRouterParam(event, 'path')
  
  if (!path) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing path parameter'
    })
  }

  // Get session cookie from incoming request
  const sessionCookie = getRequestHeader(event, 'cookie') || ''
  
  // Build the upstream API URL
  const apiUrl = `http://plan-api-service.plan:8080/api/v1/${path}`
  
  // Get query string
  const queryString = event.searchParams?.toString() || ''
  const fullUrl = queryString ? `${apiUrl}?${queryString}` : apiUrl
  
  // Get the request method and body
  const method = event.method
  let body: any = undefined
  
  if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
    body = await readBody(event)
  }
  
  try {
    // Forward the request to the Go API with session cookie
    const response = await $fetch(fullUrl, {
      method,
      headers: {
        'Cookie': sessionCookie,
        'Content-Type': 'application/json',
      },
      body,
    })
    
    return response
  } catch (error: any) {
    // Pass through the error from the API
    if (error.statusCode) {
      throw createError({
        statusCode: error.statusCode,
        statusMessage: error.statusMessage || 'API Error',
        data: error.data,
      })
    }
    
    throw createError({
      statusCode: 500,
      statusMessage: 'Proxy error',
      data: error.message,
    })
  }
})
