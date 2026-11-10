// Logout: the Go API deletes the session and clears the cookies.
export default defineEventHandler(async (event) => {
  return await forwardToGo(event, "auth/logout", { method: "POST" });
});
