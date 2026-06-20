import { db } from "../../utils/db";
import { users } from "../../../drizzle/schema";
import { eq } from "drizzle-orm";

export default defineOAuthAuthentikEventHandler({
  async onSuccess(event, { user, tokens }) {
    try {
      if (!user?.sub || !user?.email || !tokens?.access_token) {
        console.error("Authentik OAuth: Missing required user or token data", {
          user,
          hasTokens: !!tokens,
        });
        return sendRedirect(event, "/?error=auth_data_missing");
      }

      let dbUser = await db.query.users.findFirst({
        where: eq(users.sub, user.sub),
      });

      if (!dbUser) {
        const [newUser] = await db
          .insert(users)
          .values({
            sub: user.sub,
            email: user.email,
          })
          .returning();
        dbUser = newUser;
      }

      if (!dbUser) {
        console.error("Authentik OAuth: Failed to find or create user");
        return sendRedirect(event, "/?error=user_creation_failed");
      }

      const sessionData = JSON.parse(
        JSON.stringify({
          user: {
            id: Number(dbUser.id),
            sub: String(dbUser.sub),
            email: String(dbUser.email),
          },
          loggedInAt: Number(Date.now()),
        }),
      );

      await setUserSession(event, sessionData);

      console.log(
        `Authentik OAuth: Successfully authenticated user ${dbUser.email}`,
      );
      return sendRedirect(event, "/");
    } catch (error) {
      console.error("Authentik OAuth success handler error:", error);
      return sendRedirect(event, "/?error=auth_internal_error");
    }
  },
  onError(event, error) {
    console.error("Authentik OAuth error:", {
      message: error?.message,
      statusCode: error?.statusCode,
      url: event.node.req.url,
    });
    return sendRedirect(event, "/?error=auth_failed");
  },
});
