import { defineConfig } from "vitest/config";
import { defineVitestProject } from "@nuxt/test-utils/config";

export default defineConfig(async () => ({
  test: {
    projects: [
      {
        name: "unit",
        test: {
          include: ["test/{e2e,unit}/*.{test,spec}.ts"],
          environment: "node",
          setupFiles: ["./test/setup.ts"],
        },
      },
      await defineVitestProject({
        test: {
          name: "nuxt",
          include: ["test/nuxt/*.{test,spec}.ts"],
          environment: "nuxt",
        },
      }),
    ],
  },
}));
