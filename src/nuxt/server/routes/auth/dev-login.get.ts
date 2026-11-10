// Local development login without Zitadel. The Go API creates the session
// and sets the cookies; we only forward the request and its response.
export default defineEventHandler(async (event) => {
  return await forwardToGo(event, "auth/dev-login", { method: "POST" });
});
