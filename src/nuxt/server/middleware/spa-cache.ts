import { setResponseHeader } from "h3";

/**
 * SPA-skalet (index.html) får aldrig cachas.
 *
 * Utan cache-headers kan webbläsaren återanvända en gammal index.html och
 * därmed ladda gamla JS-bundles (hashade asset-namn ändras inte). En sådan
 * gammal bundle kan t.ex. ha en äldre ikon-provider, vilket ger en app som
 * "ser ut att fungera" men saknar funktioner - exempelvis kunde ikonväljaren
 * söka (den routen är server-side) men inga ikoner syntes.
 */
export default defineEventHandler((event) => {
  const path = event.path.split("?")[0] ?? "";

  // API-svar och statiska assets har egna cache-regler.
  if (path.startsWith("/api/") || path.startsWith("/_nuxt/")) {
    return;
  }

  setResponseHeader(event, "cache-control", "no-store, must-revalidate");
});
