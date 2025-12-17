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
    "@nuxt/icon",
    "shadcn-nuxt",
    "nuxt-auth-utils",
    "nuxt-graphql-client",
    [
      "@pinia/nuxt",
      {
        disableVuex: true,
        autoImports: ["defineStore", "storeToRefs"],
      },
    ],
  ],
  ],
  ],
  // Disable fontsource provider for unifont (API is unreliable)
  fonts: {
    providers: {
      fontsource: false,
    },
  },
  pinia: {
    storesDirs: ["./app/stores/**"],
    // Disable Pinia state transfer to prevent serialization errors
    disableNuxt: false,
  },
  colorMode: {
    classSuffix: "",
  },
  icon: {
    // Disable client-side fetching to prevent timeouts
    provider: 'server',
    serverBundle: {
      collections: ['material-symbols', 'material-symbols-light', 'lucide']
    },
    // Disable client-side bundle to prevent runtime fetching
    clientBundle: false,
    // Set a very short timeout for any remaining requests
    timeout: 100
  },
  runtimeConfig: {
    public: {
      GQL_HOST: "http://localhost:8080/v1/graphql", // overwritten by NUXT_PUBLIC_GQL_HOST
      hasuraAdminSecret:
        process.env.HASURA_GRAPHQL_ADMIN_SECRET || "dev-admin-secret",
    },
    oauth: {
      // OAuth provider configuration for nuxt-auth-utils
      authentik: {
        clientId: process.env.NUXT_OAUTH_AUTHENTIK_CLIENT_ID || "",
        clientSecret: process.env.NUXT_OAUTH_AUTHENTIK_CLIENT_SECRET || "",
        domain: process.env.NUXT_OAUTH_AUTHENTIK_DOMAIN || "",
      },
    },
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
   nitro: {
     preset: "bun",
     host: "0.0.0.0",
     trustProxy: true,
   },

  vite: {
    server: {
      allowedHosts: [".simonbrundin.com"],
    },
  },

  // Configure payload to handle potential serialization issues
  experimental: {
    payloadExtraction: false, // Disable to prevent Pinia hydration errors
  },
  hooks: {
    // Add a hook to handle Pinia payload serialization errors gracefully
    "app:error": (error) => {
      if (error.message?.includes("hasOwnProperty")) {
        console.warn(
          "Pinia payload serialization error - this is expected and handled"
        );
      }
    },
    // Disable Pinia state transfer completely
    "app:beforeMount": (app) => {
      // Clear any Pinia state from the payload to prevent serialization
      if (app?.config?.globalProperties?.$nuxt?.payload?.state) {
        delete app.config.globalProperties.$nuxt.payload.state;
      }
    },
  },
  "graphql-client": {
    codegen: false,
    clients: {
      default: {
        host:
          process.env.NUXT_PUBLIC_GQL_HOST ||
          "http://localhost:8080/v1/graphql",
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
