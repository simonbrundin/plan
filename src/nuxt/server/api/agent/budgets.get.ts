import { createError } from "h3";

// Listar budgetgränser för agentkostnad.
export default defineEventHandler(async (event) => {
  await requireSession(event);

  try {
    return await forwardToGo(event, "agent-budgets", { method: "GET" });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Plan API:t svarar inte",
      data: error?.message,
    });
  }
});
