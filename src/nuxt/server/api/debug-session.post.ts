import { eventHandler } from "h3";

export default eventHandler(async (event) => {
  // Set a test session
  await setUserSession(event, {
    user: {
      id: "test-user-123",
      sub: "test-user-123",
      email: "test@example.com",
      accessToken: "test-access-token",
    },
    loggedInAt: Date.now(),
  });

  return { message: "Session set", success: true };
});
