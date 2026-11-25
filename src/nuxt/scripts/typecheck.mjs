#!/usr/bin/env node
// Typecheck-gate för Plan-gränssnittet.
//
// Gränssnittet har ärvt teknisk skuld: typechecken ger ett par hundra
// fel i arvskoden. Att skuldbelägga allt på en gång är ett eget projekt -
// och det är inte det här skriptet ska göra.
//
// Gateen fångar regressioner i stället: antalet fel per fil får inte öka.
// Ett nytt fel i en fil som redan har fel faller igenom, ett nytt fel i en
// ren fil faller igenom, och en rättning sänker antalet.
//
// Kör:
//   node scripts/typecheck.mjs            # gate - failar på nya fel
//   node scripts/typecheck.mjs --update   # skriv om baslinjen
//
// Baslinjen ligger i typecheck-baseline.json.

import { spawnSync } from "node:child_process";
import { readFileSync, writeFileSync, existsSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const baselinePath = join(root, "typecheck-baseline.json");
const update = process.argv.includes("--update");

// Konfigurationerna som typcheckas. Server-konfigurationen har en egen
// uppsättning fel - de räknas separat så att en rättning i den ena inte
// döljer en ny i den andra.
const projects = [
  { name: "app", config: ".nuxt/tsconfig.app.json" },
  { name: "server", config: ".nuxt/tsconfig.server.json" },
];

if (!existsSync(join(root, ".nuxt/tsconfig.app.json"))) {
  console.error(
    "Typen är inte genererad. Kör 'bun run postinstall' (nuxt prepare) först.",
  );
  process.exit(2);
}

const vueTsc = join(root, "node_modules/.bin/vue-tsc");

/**
 * Kör vue-tsc för en konfiguration och räknar fel per fil.
 *
 * @param {string} config
 * @returns {{counts: Record<string, number>, total: number}}
 */
function typecheck(config) {
  const result = spawnSync(vueTsc, ["--noEmit", "-p", config], {
    cwd: root,
    encoding: "utf8",
    maxBuffer: 64 * 1024 * 1024,
  });

  const output = `${result.stdout ?? ""}${result.stderr ?? ""}`;
  const counts = {};

  // Raderna ser ut så här:
  //   app/pages/goals/index.vue(12,3): error TS2345: ...
  const pattern = /^(.+?)\(\d+,\d+\): error TS\d+:/gm;
  let match;
  while ((match = pattern.exec(output)) !== null) {
    const file = match[1].trim();
    counts[file] = (counts[file] ?? 0) + 1;
  }

  const total = Object.values(counts).reduce((sum, n) => sum + n, 0);
  return { counts, total };
}

const current = {};

for (const project of projects) {
  const { counts, total } = typecheck(project.config);
  current[project.name] = counts;
  console.log(
    `${project.name}: ${total} fel i ${Object.keys(counts).length} filer`,
  );
}

// Baslinjen är per projekt, så att en ny fil i den ena projektet inte
// påverkar den andra.
const baseline = existsSync(baselinePath)
  ? JSON.parse(readFileSync(baselinePath, "utf8"))
  : { app: {}, server: {} };

if (update) {
  writeFileSync(baselinePath, `${JSON.stringify(current, null, 2)}\n`);
  console.log("Skrev baslinjen till typecheck-baseline.json");
  process.exit(0);
}

let nyaFel = 0;

for (const project of projects) {
  const bas = baseline[project.name] ?? {};
  const nu = current[project.name] ?? {};

  for (const [file, count] of Object.entries(nu)) {
    const before = bas[file] ?? 0;
    if (count > before) {
      nyaFel += count - before;
      console.error(
        `NYTT FEL: ${file} har ${count} fel (baslinjen: ${before})`,
      );
    }
  }

  for (const [file, count] of Object.entries(bas)) {
    if (!(file in nu)) {
      console.log(`Rättat: ${file} (${count} fel borta) - kör --update för att skriva om baslinjen`);
    }
  }
}

if (nyaFel > 0) {
  console.error(
    `\nTypecheck-gaten föll: ${nyaFel} nya fel. Fixa dem eller kör 'node scripts/typecheck.mjs --update' om de är avsiktliga.`,
  );
  process.exit(1);
}

console.log("\nTypecheck-gaten godkänd - inga nya fel.");
