import { createError, readBody } from "h3";

// Skapar en budgetgräns för agentkostnad.
export default defineEventHandler(async (event) => {
  await requireSession(event);

  const body = await readBody(event).catch(() => ({}));

  try {
    return await forwardToGo(event, "agent-budgets", {
      method: "POST",
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
