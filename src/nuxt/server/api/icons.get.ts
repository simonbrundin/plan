import { getQuery } from "h3";

interface IconResult {
  name: string;
  collection: string;
  full: string;
}

// Iconifys sökmotor - samma katalog som ikonerna renderas från.
const ICONIFY_SEARCH_URL = "https://api.iconify.design/search";

// Max antal träffar som returneras till klienten.
const RESULT_LIMIT = 64;

/**
 * Sök ikoner i Iconify-katalogen.
 *
 * Routen låg i Nuxt fram tills commit ea47707 ("remove Nuxt server routes,
 * use Go API instead") tog bort den - IconPicker.vue fortsatte dock att
 * anropa /api/icons, vilket gjorde att sökningen alltid returnerade en
 * tom lista ("Ingen ikon matchade ...").
 *
 * Sökningen är ett rent UI-behov och har ingen affärslogik, så den
 * ligger kvar i Nuxt framför att läggas till i Go API:t.
 */
export default defineEventHandler(async (event) => {
  const query = getQuery(event);
  const searchQuery = typeof query.q === "string" ? query.q.trim() : "";

  // Sök inte med för få tecken - Iconify ger inga relevanta träffar då.
  if (searchQuery.length < 2) {
    return {
      results: [],
      message: "Enter at least 2 characters to search",
    };
  }

  try {
    const response = await fetch(
      `${ICONIFY_SEARCH_URL}?query=${encodeURIComponent(searchQuery)}&limit=${RESULT_LIMIT}`,
    );

    if (!response.ok) {
      throw new Error(`Iconify API error: ${response.statusText}`);
    }

    const data = await response.json();

    // Iconify returnerar en array med "collection:name", t.ex. "mdi:test-tube".
    const results: IconResult[] = (data.icons ?? []).map((iconString: string) => {
      const parts = iconString.split(":");
      return {
        collection: parts[0] || "unknown",
        name: parts.slice(1).join(":") || iconString,
        full: iconString,
      };
    });

    return {
      results,
      total: data.total ?? results.length,
    };
  } catch (error) {
    console.error("Failed to search icons:", error);
    throw createError({
      statusCode: 502,
      statusMessage: "Icon search failed",
    });
  }
});
