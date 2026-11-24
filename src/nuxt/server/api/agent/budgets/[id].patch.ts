import { createError, readBody } from "h3";

// En budget ändras - till exempel pausas när gränsen är nådd.
export default defineEventHandler(async (event) => {
  await requireSession(event);

  const id = getRouterParam(event, "id");
  const body = await readBody(event).catch(() => ({}));

  try {
    return await forwardToGo(event, `agent-budgets/${id}`, {
      method: "PATCH",
      body,
    });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Plan API:t svarar inte",
      data: error?.message,
    });
  }
});
