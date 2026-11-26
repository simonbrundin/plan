import { marked } from "marked";

/**
 * Hjälpare för dokumentationssidorna.
 *
 * Dokumentationen ligger som markdown-filer i app/content/docs. De
 * importeras som råtext (se useDocs) och renderas här till HTML med
 * marked. Rubriker får id:n så att innehållsförteckningen kan länka
 * till dem.
 */

export interface DocHeading {
  id: string;
  text: string;
  depth: number;
}

export interface DocMeta {
  /** Filnamnet utan .md - blir URL:en /docs/<slug> */
  slug: string;
  title: string;
  description: string;
  order: number;
  icon: string;
  updated: string;
  group: string;
}

export interface DocPage extends DocMeta {
  html: string;
  headings: DocHeading[];
}

const FRONTMATTER = /^---\r?\n([\s\S]*?)\r?\n---\r?\n?/;

/**
 * Delar upp frontmatter och brödtext. Frontmatter är en enkel
 * nyckel-värde-lista - ingen YAML-parser behövs för formatet vi använder.
 */
export function parseFrontmatter(raw: string): {
  data: Record<string, string>;
  body: string;
} {
  const match = FRONTMATTER.exec(raw);
  if (!match) {
    return { data: {}, body: raw };
  }

  const frontmatter = match[1] ?? "";
  const data: Record<string, string> = {};
  for (const line of frontmatter.split(/\r?\n/)) {
    const separator = line.indexOf(":");
    if (separator === -1) continue;

    const key = line.slice(0, separator).trim();
    const value = line
      .slice(separator + 1)
      .trim()
      .replace(/^["']/, "")
      .replace(/["']$/, "")
      .trim();

    if (key) data[key] = value;
  }

  return { data, body: raw.slice(match[0].length) };
}

/**
 * Gör en rubrik till ett ankarlänk-id. Svenska bokstägg behålls - de är
 * vanliga i rubrikerna och gör länkarna läsbara. Webbläsare hanterar
 * icke-ASCII i ankarlänkar utan problem.
 */
export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9åäöéèü]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

/**
 * Tar bort HTML-taggar - används för att få ren text ur en rubrik.
 */
function stripTags(html: string): string {
  return html.replace(/<[^>]*>/g, "").trim();
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

/**
 * Renderar markdown till HTML och samlar samtidigt in rubrikerna till
 * en innehållsförteckning.
 *
 * Rå HTML i markdownen.escape:as - innehållet kommer från repot, men det
 * ska aldrig kunna köra kod i webbläsaren.
 */
export function renderMarkdown(body: string): {
  html: string;
  headings: DocHeading[];
} {
  const headings: DocHeading[] = [];
  const usedIds = new Map<string, number>();

  const renderer = new marked.Renderer();

  renderer.heading = function ({ tokens, depth }) {
    const inner = this.parser.parseInline(tokens);
    const text = stripTags(inner);

    if (depth !== 2 && depth !== 3) {
      return `<h${depth}>${inner}</h${depth}>`;
    }

    const base = slugify(text) || "avsnitt";
    const count = usedIds.get(base) ?? 0;
    usedIds.set(base, count + 1);
    const id = count === 0 ? base : `${base}-${count + 1}`;

    headings.push({ id, text, depth });
    return `<h${depth} id="${id}">${inner}</h${depth}>`;
  };

  renderer.html = ({ text }) => escapeHtml(text);

  const html = marked.parse(body, {
    renderer,
    gfm: true,
    async: false,
  }) as string;

  return { html, headings };
}

/**
 * Bygger en docs-sida av en markdown-fil.
 */
export function buildDocPage(slug: string, raw: string): DocPage {
  const { data, body } = parseFrontmatter(raw);
  const { html, headings } = renderMarkdown(body);

  return {
    slug,
    title: data.title ?? slug,
    description: data.description ?? "",
    order: Number(data.order ?? 99),
    icon: data.icon ?? "i-lucide-file-text",
    updated: data.updated ?? "",
    group: data.group ?? "Dokumentation",
    html,
    headings,
  };
}
