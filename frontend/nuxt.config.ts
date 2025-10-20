// https://nuxt.com/docs/api/configuration/nuxt-config
import { defineNuxtConfig } from "nuxt/config";

export default defineNuxtConfig({
  // extends: [
  //   // "github:simonbrundin/nuxt-base-layer",
  //   // ["github:simonbrundin/nuxt-base-layer", { install: true }],
  //   "../../nuxt-base-layer/",
  //   // eller med specifik branch/tag:
  //   // 'github:username/my-template-layer#v1.0.0'
  // ],
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  css: ["@/assets/css/main.css"],
  ui: {
    theme: {
      colors: [
        "primary",
        "secondary",
        "tertiary",
        "info",
        "success",
        "warning",
        "error",
      ],
    },
  },

  modules: [
    "@nuxt/image",
    "@nuxtjs/color-mode",
    "@nuxtjs/tailwindcss",
    "@nuxt/ui",
    "shadcn-nuxt",
    "nuxt-auth-utils",
    "nuxt-graphql-client",
    "@pinia/nuxt",
  ],
  colorMode: {
    classSuffix: "",
  },
  runtimeConfig: {
    public: {
      authBaseUrl: process.env.BETTER_AUTH_URL || "http://localhost:3000",
      GQL_HOST: "http://localhost:8080/v1/graphql", // overwritten by NUXT_PUBLIC_GQL_HOST
      hasuraAdminSecret: process.env.HASURA_GRAPHQL_ADMIN_SECRET || 'dev-admin-secret',
    },
    // oauth: {
    //   // provider in lowercase (github, google, etc.)
    //   authentik: {
    //     clientId: "...",
    //     clientSecret: "...",
    //   },
    // },
  },
  shadcn: {
    /**
     * Prefix for all the imported component
     */
    prefix: "",
    /**
     * Directory that the component lives in.
     * @default "./components/ui"
     */
    componentDir: "./app/components/ui",
  },
  // nitro: {
  //   preset: "bun",
  // },
  
  // Configure payload to handle potential serialization issues
  experimental: {
    payloadExtraction: true,
  },
  "graphql-client": {
    codegen: false,
    clients: {
      default: {
        host: process.env.NUXT_PUBLIC_GQL_HOST || "http://localhost:8080/v1/graphql",
        token: {
          type: "Bearer",
          name: "Authorization",
        },
        // Admin secret is used for schema introspection during codegen
        // The graphql-auth.ts plugin will override this with JWT tokens at runtime
        headers: process.env.HASURA_GRAPHQL_ADMIN_SECRET
          ? {
              "x-hasura-admin-secret": process.env.HASURA_GRAPHQL_ADMIN_SECRET,
            }
          : {},
      },
    },
  },
});
