import { defaultPlugins, defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../api/docs/swagger.json",
  output: {
    path: "gen",
    format: "prettier",
  },
  plugins: [
    ...defaultPlugins,
    "zod",
    {
      name: "@hey-api/client-fetch",
      runtimeConfigPath: "./lib/client.ts",
      throwOnError: true,
    },
    "@tanstack/react-query",
    { name: "@hey-api/sdk", auth: false },
  ],
});
