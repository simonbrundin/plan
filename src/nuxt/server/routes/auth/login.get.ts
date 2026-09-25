import { sendRedirect } from "h3";

// Redirect to Go API for OAuth login
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const goApiUrl = config.public.goApiUrl || 'http://localhost:8080'
  
  // Redirect to Go API login endpoint
  return sendRedirect(event, `${goApiUrl}/auth/login`)
})
