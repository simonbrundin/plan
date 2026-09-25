import { getCookie } from "h3";

// Get current user session info
export default defineEventHandler(async (event) => {
  const sessionCookie = getCookie(event, 'plan_session')
  const userCookie = getCookie(event, 'plan_user')
  
  if (!sessionCookie || !userCookie) {
    throw createError({
      statusCode: 401,
      message: 'Not authenticated'
    })
  }
  
  try {
    const user = JSON.parse(atob(userCookie))
    return {
      sub: user.sub,
      email: user.email,
    }
  } catch {
    throw createError({
      statusCode: 401,
      message: 'Invalid session'
    })
  }
})
