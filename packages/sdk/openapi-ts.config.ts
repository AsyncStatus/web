import { defaultPlugins, defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../apps/web-api-v3/docs/swagger.json",
  output: {
    path: "gen",
    format: "prettier",
    lint: "eslint",
  },
  plugins: [
    ...defaultPlugins,
    "zod",
    { name: "@hey-api/client-fetch", runtimeConfigPath: "./lib/client.ts", throwOnError: true },
    "@tanstack/react-query",
    { name: "@hey-api/sdk", auth: false },
  ],
});
