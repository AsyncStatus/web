import { getASError } from "@asyncstatus/sdk/client.gen";
import { AsyncStatusLogo } from "@asyncstatus/ui/components/async-status-logo.js";
import { SimpleLayout } from "@asyncstatus/ui/components/simple-layout.js";
import { QueryClient } from "@tanstack/react-query";
import { createRouter } from "@tanstack/react-router";

import { ErrorFallback } from "./components/error-fallback";
import { NotFound } from "./components/not-found";
import { routeTree } from "./routeTree.gen";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      throwOnError: true,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      retry(failureCount, error) {
        const asError = getASError(error);
        if (asError && asError.internal_status.startsWith("ASE-4")) {
          return false;
        }

        return failureCount < 3;
      },
    },
  },
});

export const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: "intent",
  defaultPreloadStaleTime: 0,
  scrollRestoration: true,
  defaultPendingComponent: () => {
    return (
      <div className="flex h-screen w-screen items-center justify-center">
        <AsyncStatusLogo className="h-4 w-auto animate-pulse duration-1250" />
      </div>
    );
  },
  defaultNotFoundComponent: NotFound,
  defaultErrorComponent: (props) => (
    <SimpleLayout href={import.meta.env.VITE_WEB_MARKETING_URL}>
      <ErrorFallback {...props} />
    </SimpleLayout>
  ),
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
