import type { CreateClientConfig } from "../gen/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: `${process.env.NEXT_PUBLIC_API_URL}${process.env.NEXT_PUBLIC_API_URL_PREFIX}`,
  credentials: "include",
});
