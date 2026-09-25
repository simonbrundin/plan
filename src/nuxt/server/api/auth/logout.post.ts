import { eventHandler, deleteCookie } from "h3";

export default eventHandler((event) => {
  deleteCookie(event, 'plan_session', {
    path: '/',
  });
  
  return { success: true };
});
