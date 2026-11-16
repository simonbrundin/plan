import { getQuery } from "h3";
import translate from "translate";

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
  const shouldTranslate = query.translate === "true";

  // Sök inte med för få tecken - Iconify ger inga relevanta träffar då.
  if (searchQuery.length < 2) {
    return {
      results: [],
      message: "Enter at least 2 characters to search",
    };
  }

  try {
    // Nya mål skrivs ofta på svenska, medan Iconify-katalogen huvudsakligen
    // använder engelska sökord. Översätt bara den automatiska målikonsökningen;
    // IconPicker ska söka exakt på det användaren skriver.
    let iconQuery = searchQuery;
    if (shouldTranslate) {
      try {
        iconQuery = await translate(searchQuery, { from: "sv", to: "en" });
      } catch (error) {
        console.warn("Could not translate goal title for icon search:", error);
      }
    }

    const terms = [
      iconQuery,
      ...iconQuery.split(/\s+/).filter((term) => term.length >= 2),
    ];

    for (const term of [...new Set(terms)]) {
      const response = await fetch(
        `${ICONIFY_SEARCH_URL}?query=${encodeURIComponent(term)}&limit=${RESULT_LIMIT}`,
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

      if (results.length > 0) {
        return {
          results,
          total: data.total ?? results.length,
        };
      }
    }

    return { results: [], total: 0 };
  } catch (error) {
    console.error("Failed to search icons:", error);
    throw createError({
      statusCode: 502,
      statusMessage: "Icon search failed",
    });
  }
});
