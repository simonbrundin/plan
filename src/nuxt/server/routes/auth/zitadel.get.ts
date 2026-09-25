import { defineOAuthZitadelEventHandler } from 'nuxt-auth-utils'

export default defineOAuthZitadelEventHandler({
  onSuccess: async (event, { user, tokens }) => {
    await setUserSession(event, {
      user: {
        id: user.sub,
        email: user.email,
        name: user.name,
      },
      secure: {
        accessToken: tokens.access_token,
        refreshToken: tokens.refresh_token,
        expiresAt: Date.now() + (tokens.expires_in * 1000),
      },
      loggedInAt: Date.now(),
    })
    return sendRedirect(event, '/')
  },
  onError: (event, error) => {
    console.error('Zitadel OAuth error:', error)
    return sendRedirect(event, '/login?error=auth_failed')
  },
})
