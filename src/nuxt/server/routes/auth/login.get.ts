import { sendRedirect, setCookie } from "h3";

// Redirect to Go API for OAuth login.
// Sets a `wt` cookie identifying this worktree so the dev gateway (Caddy)
// can route the OAuth callback to the correct worktree's Go API.
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const goApiUrl = config.public.goApiUrl || 'http://localhost:8080'
  const worktree = process.env.TILT_WORKTREE || process.env.WORKTREE_NAME || ''

  if (worktree) {
    // Cookie scoped to root so Caddy can match it on /oauth/callback.
    // Short max-age - the OAuth flow completes within seconds.
    setCookie(event, 'wt', worktree, {
      httpOnly: false,
      sameSite: 'lax',
      maxAge: 600,
      path: '/',
    })
  }

  return sendRedirect(event, `${goApiUrl}/api/v1/auth/login`)
})