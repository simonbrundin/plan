import { eventHandler } from "h3";

export default eventHandler(() => {
	return { status: "healthy", timestamp: new Date().toISOString() };
});
