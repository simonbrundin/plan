import { describe, expect, it } from "bun:test";
import { escapeHtml, isSafeHref, renderMarkdown } from "../../app/utils/markdown";

describe("renderMarkdown", () => {
	it("renderar rubriker som HTML", () => {
		expect(renderMarkdown("## Hej")).toContain("<h2>Hej</h2>");
	});

	it("renderar listor", () => {
		const html = renderMarkdown("- ett\n- två");
		expect(html).toContain("<ul>");
		expect(html).toContain("<li>ett</li>");
	});

	it("renderar kodblock utan att tolka innehållet", () => {
		const html = renderMarkdown("```bash\necho <hej>\n```");
		expect(html).toContain("<pre>");
		expect(html).toContain("echo &lt;hej&gt;");
	});

	it("renderar fet och kursiv stil", () => {
		expect(renderMarkdown("**fet**")).toContain("<strong>fet</strong>");
		expect(renderMarkdown("*kursiv*")).toContain("<em>kursiv</em>");
	});

	it("behåller nya rader i brödtext", () => {
		expect(renderMarkdown("rad ett\nrad två")).toContain("<br>");
	});

	it("escapar rå HTML i källan", () => {
		const html = renderMarkdown('<script>alert(1)</script>');
		expect(html).not.toContain("<script>");
		expect(html).toContain("&lt;script&gt;");
	});

	it("sätter säkra attribut på länkar", () => {
		const html = renderMarkdown("[Plan](https://plan.simonbrundin.com)");
		expect(html).toContain('href="https://plan.simonbrundin.com"');
		expect(html).toContain('rel="noopener noreferrer"');
	});

	it("skriver ut osäkra länkar som text", () => {
		const html = renderMarkdown("[farligt](javascript:alert(1))");
		expect(html).not.toContain("javascript:");
		expect(html).not.toContain("<a ");
	});

	it("returnerar tom sträng för tom indata", () => {
		expect(renderMarkdown("")).toBe("");
	});
});

describe("isSafeHref", () => {
	it("tillåter http, https, mailto och tel", () => {
		expect(isSafeHref("https://example.com")).toBe(true);
		expect(isSafeHref("http://example.com")).toBe(true);
		expect(isSafeHref("mailto:a@example.com")).toBe(true);
		expect(isSafeHref("tel:+46123456")).toBe(true);
	});

	it("tillåter interna länkar", () => {
		expect(isSafeHref("/goal/12")).toBe(true);
		expect(isSafeHref("#avsnitt")).toBe(true);
	});

	it("vägrar javascript och data", () => {
		expect(isSafeHref("javascript:alert(1)")).toBe(false);
		expect(isSafeHref("data:text/html,<script>")).toBe(false);
	});
});

describe("escapeHtml", () => {
	it("escapar HTML-tecken", () => {
		expect(escapeHtml(`<a href="x">&'`)).toBe(
			"&lt;a href=&quot;x&quot;&gt;&amp;&#39;",
		);
	});
});
