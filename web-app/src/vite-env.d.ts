/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly WEB_API_URL: string;
  readonly VITE_WEB_MARKETING_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
