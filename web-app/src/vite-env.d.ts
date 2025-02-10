/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly WEB_API_URL: string;
  readonly WEB_API_URL_PREFIX: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
