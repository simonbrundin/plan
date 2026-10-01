import { defineEventHandler } from "h3";

// Keep this endpoint independent from the database and authentication layers so
// Kubernetes can use it for liveness and readiness probes.
export default defineEventHandler(() => {
	const start = Date.now();
	console.log(`[TIMING] /api/health start at ${start}`);
	const result = { status: "ok", timestamp: new Date().toISOString() };
	console.log(`[TIMING] /api/health done in ${Date.now() - start}ms`);
	return result;
});
