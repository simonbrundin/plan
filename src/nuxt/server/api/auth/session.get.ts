// The current user, according to the Go API (which owns the session).
export default defineEventHandler(async (event) => {
  return await forwardToGo(event, "auth/session");
});
