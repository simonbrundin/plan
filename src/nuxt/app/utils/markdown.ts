import { marked, Renderer } from "marked";

/**
 * Renderar markdown till HTML för chattmeddelanden.
 *
 * Agenten svarar i markdown (rubriker, listor, kodblock, länkar) och utan
 * rendering blir svaret svårläst - asterisker och brödtext blandar sig.
 *
 * Säkerhet: innehållet sätts in med v-html, så all rå HTML escapas och
 * länkar filtreras. Kodblock och inline-kod escapas av marked i sig.
 */

/** Escape av HTML-tecken så att rå HTML inte kan passera igenom. */
export function escapeHtml(value: string): string {
	return value
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#39;");
}

/**
 * Endast http(s), mailto, tel och interna länkar (som börjar med / eller #)
 * får användas. Annars skrivs länken ut som text.
 */
export function isSafeHref(href: string): boolean {
	const trimmed = href.trim();
	if (trimmed.startsWith('/') || trimmed.startsWith('#')) return true;
	return /^(https?:|mailto:|tel:)/i.test(trimmed);
}

// Cacharade renderingar - chatten ritar om sig när ett svar streamas in,
// och samma meddelande renderas om många gånger.
const cache = new Map<string, string>();
const CACHE_LIMIT = 200;

/** Renderar markdown till säker HTML. */
export function renderMarkdown(source: string): string {
	if (!source) return "";

	const cached = cache.get(source);
	if (cached !== undefined) return cached;

	const html = render(source);

	if (cache.size >= CACHE_LIMIT) {
		const oldest = cache.keys().next().value;
		if (oldest !== undefined) cache.delete(oldest);
	}
	cache.set(source, html);
	return html;
}

function render(source: string): string {
	const renderer = new Renderer();

	// Rå HTML i källan skrivs ut som text i stället för att tolkas.
	renderer.html = ({ text }) => escapeHtml(text);

	renderer.link = function ({ href, title, tokens }) {
		const text = this.parser.parseInline(tokens);
		if (!isSafeHref(href)) return text;
		const titleAttr = title ? ` title="${escapeHtml(title)}"` : "";
		return `<a href="${escapeHtml(href.trim())}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
	};

	renderer.image = ({ href, title, text }) => {
		if (!isSafeHref(href)) return escapeHtml(text);
		const titleAttr = title ? ` title="${escapeHtml(title)}"` : "";
		return `<img src="${escapeHtml(href.trim())}" alt="${escapeHtml(text)}"${titleAttr}>`;
	};

	return marked.parse(source, {
		renderer,
		gfm: true,
		breaks: true,
	}) as string;
}
