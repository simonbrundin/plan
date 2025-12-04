export default defineNuxtRouteMiddleware((to) => {
  if (process.env.NODE_ENV === 'development') return

  const { user } = useUserSession()

  // Allow access to login and signup pages
  if (to.path === '/login' || to.path === '/signup' || to.path === '/api/health') {
    return
  }

  // Redirect to login if not authenticated
  if (!user.value) {
    return navigateTo('/login')
  }
})