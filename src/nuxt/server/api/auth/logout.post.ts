import { deleteCookie } from "h3";

// Clear session cookies
export default defineEventHandler(async (event) => {
  deleteCookie(event, 'plan_session', {
    path: '/',
  })
  deleteCookie(event, 'plan_user', {
    path: '/',
  })
  
  return { success: true }
})
