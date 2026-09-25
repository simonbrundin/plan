import { eventHandler } from "h3";

export default eventHandler(async (event) => {
  const session = await getUserSession(event);
  
  return {
    session,
    hasUser: !!session.user,
    userKeys: session.user ? Object.keys(session.user) : [],
    userAccessToken: session.user?.accessToken || 'NO_TOKEN',
  };
});
