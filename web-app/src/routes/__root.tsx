import { lazy, Suspense } from "react";
import {
  getCurrentUserOptions,
  getSessionOptions,
} from "@asyncstatus/sdk/@tanstack/react-query.gen";
import { QueryClient, useSuspenseQueries } from "@tanstack/react-query";
import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()(
  { component: Component }
);

function Component() {
  useSuspenseQueries({
    queries: [
      { ...getSessionOptions(), staleTime: Infinity },
      { ...getCurrentUserOptions(), staleTime: Infinity },
    ],
  });

  return (
    <>
      <Outlet />
      <Suspense>
        <TanStackRouterDevtools position="bottom-right" />
      </Suspense>
      <Suspense>
        <ReactQueryDevtoolsProduction />
      </Suspense>
    </>
  );
}

const TanStackRouterDevtools =
  process.env.NODE_ENV === "production"
    ? () => null
    : lazy(() =>
        import("@tanstack/router-devtools").then((res) => ({
          default: res.TanStackRouterDevtools,
        }))
      );

const ReactQueryDevtoolsProduction =
  process.env.NODE_ENV === "production"
    ? () => null
    : lazy(() =>
        import("@tanstack/react-query-devtools/production").then((d) => ({
          default: d.ReactQueryDevtools,
        }))
      );
