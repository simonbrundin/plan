import { eventHandler, sendRedirect } from "h3";

export default eventHandler(async (event) => {
  console.log("Test auth endpoint called");
  return {
    message: "Auth test endpoint works",
    timestamp: new Date().toISOString()
  };
});
