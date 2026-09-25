import { eventHandler, getQuery, sendRedirect } from "h3";

// Callback from Go API after successful OAuth
// Receives session token and user info, stores in cookie
export default eventHandler(async (event) => {
  const query = getQuery(event)
  
  const session = query.session as string
  const sub = query.sub as string
  const email = query.email as string

  if (!session) {
    console.error("No session token in callback")
    return sendRedirect(event, "/login?error=no_session")
  }

  // Store session in cookie
  setCookie(event, 'plan_session', session, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: '/',
  })

  // Also store user info in a separate cookie (readable by client)
  const userData = JSON.stringify({ sub, email })
  setCookie(event, 'plan_user', Buffer.from(userData).toString('base64'), {
    httpOnly: false, // Readable by client
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: '/',
  })

  console.log("Session stored for user:", sub)
  return sendRedirect(event, "/")
})
