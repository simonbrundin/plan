import { eventHandler } from "h3";

export default eventHandler(async (event) => {
  console.log("Debug session: Setting session");
  
  const sessionData = {
    user: {
      id: "test-user-123",
      sub: "test-user-123",
      email: "test@example.com",
      accessToken: "test-access-token",
    },
    loggedInAt: Date.now(),
  };
  
  console.log("Debug session: Data to set", JSON.stringify(sessionData));
  
  await setUserSession(event, sessionData);
  
  const savedSession = await getUserSession(event);
  console.log("Debug session: Saved session", JSON.stringify(savedSession));

  return { message: "Session set", success: true, savedSession };
});
