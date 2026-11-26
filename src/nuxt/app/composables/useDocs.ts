import type { DocMeta, DocPage } from "~/utils/docs";
import { buildDocPage } from "~/utils/docs";

/**
 * Laddar dokumentationssidorna från markdown-filerna i
 * app/content/docs. Filerna importeras som råtext och byggs till
 * sidor med metadata + renderad HTML.
 *
 * Importen sker statiskt (via import.meta.glob) så att filerna ingår
 * i bygget och att en ny undersida bara kräver att man lägger till
 * en fil - ingen lista att underhålla.
 */
const rawModules = import.meta.glob("../content/docs/*.md", {
  eager: true,
  query: "?raw",
  import: "default",
}) as Record<string, string>;

function pages(): DocPage[] {
  return Object.entries(rawModules)
    .map(([path, raw]) => {
      const slug = path.split("/").pop()?.replace(/\.md$/, "") ?? path;
      return buildDocPage(slug, raw);
    })
    .sort((a, b) => a.order - b.order);
}

export function useDocs() {
  const all = pages();

  const list = computed<DocMeta[]>(() =>
    all.map(({ slug, title, description, order, icon, updated, group }) => ({
      slug,
      title,
      description,
      order,
      icon,
      updated,
      group,
    })),
  );

  const groups = computed(() => {
    const map = new Map<string, DocMeta[]>();
    for (const doc of list.value) {
      const bucket = map.get(doc.group) ?? [];
      bucket.push(doc);
      map.set(doc.group, bucket);
    }
    return [...map.entries()];
  });

  const find = (slug: string): DocPage | undefined =>
    all.find((page) => page.slug === slug);

  return { list, groups, find };
}
