import { db } from "../../utils/db";
import { users } from "../../../drizzle/schema";
import { eq } from "drizzle-orm";

export default defineOAuthAuthentikEventHandler({
  config: {
    emailRequired: true,
  },
  async onSuccess(event, { user, tokens }) {
    try {
      // Validate input data
      if (!user?.sub || !user?.email || !tokens?.access_token) {
        console.error("Authentik OAuth: Missing required user or token data", { user, hasTokens: !!tokens });
        return sendRedirect(event, "/?error=auth_data_missing");
      }

      // Find or create user in database
      let dbUser = await db.query.users.findFirst({
        where: eq(users.sub, user.sub),
      });

      if (!dbUser) {
        // Create new user
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

      // Set user session with validated data
      await setUserSession(event, {
        user: {
          id: dbUser.id,
          sub: dbUser.sub,
          email: dbUser.email,
        },
        session: {
          loginTime: Date.now(),
          accessToken: tokens.access_token,
        },
      });

      console.log(`Authentik OAuth: Successfully authenticated user ${dbUser.email}`);
      return sendRedirect(event, "/");
    } catch (error) {
      console.error("Authentik OAuth success handler error:", error);
      return sendRedirect(event, "/?error=auth_internal_error");
    }
  },
  // Enhanced error handling to prevent 500 errors
  onError(event, error) {
    console.error("Authentik OAuth error:", {
      error: error?.message || error,
      code: error?.code,
      statusCode: error?.statusCode,
      url: event.node.req.url,
    });

    // Don't throw errors, redirect gracefully
    try {
      return sendRedirect(event, "/?error=auth_failed");
    } catch (redirectError) {
      console.error("Authentik OAuth: Even redirect failed:", redirectError);
      // Last resort: return a basic response
      return {
        statusCode: 302,
        headers: { Location: "/?error=auth_redirect_failed" }
      };
    }
  },
});
