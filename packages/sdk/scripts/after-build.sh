#!/bin/bash

echo "✨ Adding ASError interceptor to client.gen.ts"

cat << EOF >> gen/client.gen.ts

client.interceptors.error.use((error) => {
  const asError = getASError(error);
  if (asError) {
    return asError;
  }

  return error;
});

export class ASError {
  details: string;
  errors_map?: { [key: string]: string };
  internal_status: string;
  message: string;
  constructor(
    details: string,
    message: string,
    internal_status: string,
    errors_map?: { [key: string]: string },
  ) {
    this.details = details;
    this.message = message;
    this.internal_status = internal_status;
    this.errors_map = errors_map;
  }
}

export function getASError(err: any): ASError | undefined {
  if (!err || typeof err !== "object") {
    return;
  }
  if (err.details === undefined || err.message === undefined || err.internal_status === undefined) {
    return;
  }

  return new ASError(err.details, err.message, err.internal_status, err.errors_map);
}
EOF

echo "🧹 Formatting client.gen.ts"

./node_modules/.bin/prettier --write ./gen/client.gen.ts

echo "🚀 Done"
