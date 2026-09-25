import { eventHandler } from "h3";

export default eventHandler(async (event) => {
  const session = await getUserSession(event);
  console.log("Debug session GET:", JSON.stringify(session));
  return { 
    raw: session,
    user: session.user,
    id: session.id,
    keys: Object.keys(session)
  };
});
