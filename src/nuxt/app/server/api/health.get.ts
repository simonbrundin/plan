import { defineEventHandler } from "h3";

// Keep this endpoint independent from the database and authentication layers so
// Kubernetes can use it for liveness and readiness probes.
export default defineEventHandler(() => {
	return { status: "healthy", timestamp: new Date().toISOString() };
});
