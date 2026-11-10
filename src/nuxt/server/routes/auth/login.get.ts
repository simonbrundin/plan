// Start the OAuth login. The Go API owns the flow: it builds the Zitadel
// URL, sets the `oauth_state` cookie and redirects the browser to Zitadel.
export default defineEventHandler(async (event) => {
  return await forwardToGo(event, "auth/login");
});
