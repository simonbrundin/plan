import { db } from "../../utils/db";
import { users } from "../../../drizzle/schema";
import { eq } from "drizzle-orm";

export default defineOAuthZitadelEventHandler({
  config: {
    redirectURL: process.env.NUXT_OAUTH_ZITADEL_REDIRECT_URL,
  },
  async onSuccess(event, { user, tokens: _tokens }) {
    try {
      if (!user?.sub || !user?.email) {
        console.error("Zitadel OAuth: Missing required user data", { user });
        return sendRedirect(event, "/?error=auth_data_missing");
      }

      let dbUser = await db.query.users.findFirst({
        where: eq(users.sub, user.sub),
      });

      if (!dbUser) {
        dbUser = await db.query.users.findFirst({
          where: eq(users.email, user.email),
        });
        if (dbUser) {
          await db.update(users).set({ sub: user.sub }).where(eq(users.id, dbUser.id));
          dbUser.sub = user.sub;
        }
      }

      if (!dbUser) {
        dbUser = await db.insert(users).values({
          sub: user.sub,
          email: user.email,
        }).returning().then(([newUser]) => newUser);
      }

      if (!dbUser) {
        console.error("Zitadel OAuth: Failed to find or create user");
        return sendRedirect(event, "/?error=user_creation_failed");
      }

      await setUserSession(event, {
        user: {
          id: Number(dbUser.id),
          sub: String(dbUser.sub),
          email: String(dbUser.email),
        },
        loggedInAt: Number(Date.now()),
      });

      console.log(`Zitadel OAuth: Successfully authenticated user ${dbUser.email}`);
      return sendRedirect(event, "/");
    } catch (error) {
      console.error("Zitadel OAuth success handler error:", error);
      return sendRedirect(event, "/?error=auth_internal_error");
    }
  },
  onError(event, error) {
    console.error("Zitadel OAuth error:", {
      message: error?.message,
      statusCode: error?.statusCode,
      url: event.node.req.url,
    });
    return sendRedirect(event, "/?error=auth_failed");
  },
});