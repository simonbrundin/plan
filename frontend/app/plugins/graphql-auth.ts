export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig()
  nuxtApp.hook('gql:auth:init', async ({ client, token }) => {
    // Always use admin secret in development for easier testing
    if (process.dev) {
      client.setHeaders({
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      })
      return
    }

     // In production, use JWT token from Authentik with user-specific headers
     // The JWT includes x-hasura-user-sub, x-hasura-default-role, etc
     try {
       const { session } = useUserSession()

       // Validate session structure to prevent hydration errors
       if (session.value && typeof session.value === 'object' && session.value.accessToken && session.value.user?.id) {
         // Ensure the token is a valid string
         const accessToken = session.value.accessToken
         if (typeof accessToken === 'string' && accessToken.length > 0) {
           token.value = accessToken
           // Set user-specific headers for Hasura RLS
           client.setHeaders({
             'x-hasura-user-id': session.value.user.id.toString(),
             'x-hasura-role': 'user'
           })
           console.log('GraphQL auth: Using JWT token with user-specific headers for user', session.value.user.id)
         } else {
           console.warn('GraphQL auth: Invalid access token format, falling back to admin secret')
           // Fallback to admin secret if token is invalid
           client.setHeaders({
             'x-hasura-admin-secret': process.env.HASURA_GRAPHQL_ADMIN_SECRET || 'fallback-admin-secret'
           })
         }
       } else {
         console.warn('GraphQL auth: No valid session found - user must authenticate to access goals')
         // Don't set any headers for anonymous users - they won't be able to access user-specific data
       }
    } catch (error) {
      console.error('GraphQL auth: Critical error getting user session:', error)
      // Fallback to admin secret on any error to prevent 500 errors
      try {
        client.setHeaders({
          'x-hasura-admin-secret': process.env.HASURA_GRAPHQL_ADMIN_SECRET || 'fallback-admin-secret'
        })
        console.log('GraphQL auth: Successfully fell back to admin secret after error')
      } catch (fallbackError) {
        console.error('GraphQL auth: Even fallback failed:', fallbackError)
      }
    }
  })
})
