import { createError } from "h3";

// En budget tas bort.
export default defineEventHandler(async (event) => {
  await requireSession(event);

  const id = getRouterParam(event, "id");

  try {
    return await forwardToGo(event, `agent-budgets/${id}`, { method: "DELETE" });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Plan API:t svarar inte",
      data: error?.message,
    });
  }
});
