import { eventHandler } from "h3";

export default eventHandler(() => {
	return { status: "ok", timestamp: new Date().toISOString() };
});
