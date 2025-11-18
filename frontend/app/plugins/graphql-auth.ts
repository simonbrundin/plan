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
       const { user } = useUserSession()

       // Validate session structure to prevent hydration errors
       if (user.value && typeof user.value === 'object' && user.value.id) {
         // Set user-specific headers for Hasura RLS
         client.setHeaders({
           'x-hasura-user-id': user.value.id.toString(),
           'x-hasura-role': 'user'
         })
         console.log('GraphQL auth: Using session for user', user.value.id)
       } else {
         console.warn('GraphQL auth: No valid session found - user must authenticate to access goals')
       }
    } catch (error) {
      console.error('GraphQL auth: Critical error getting user session:', error)
    }
  })
})
