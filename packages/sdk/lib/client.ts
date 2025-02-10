import type { CreateClientConfig } from "../gen/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: `${(import.meta as any).env.VITE_WEB_API_URL}`,
  credentials: "include",
});
