import { createError, getRouterParam, readBody } from "h3";

// Skickar ett meddelande till agenten. Svaret streamas (NDJSON) rakt
// vidare till webbläsaren så att svaret syns medan det skrivs.
export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: "Missing conversation id" });
  }

  const body = await readBody(event).catch(() => ({}));

  await requireSession(event);

  try {
    return await pipeFromAgent(event, `conversations/${id}/messages`, {
      method: "POST",
      body,
    });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Agent-API:t svarar inte",
      data: error?.message,
    });
  }
});
