import { PropsWithChildren, StrictMode } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarProvider,
  SidebarRail,
} from "@asyncstatus/ui/components/sidebar.js";
import { cn } from "@asyncstatus/ui/lib/utils.js";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createRouter, RouterProvider } from "@tanstack/react-router";
import ReactDOM from "react-dom/client";

import { ProjectSelectSkeleton } from "./components/project-select";
import { SidebarMainLinks } from "./components/sidebar-main-links";
import { SidebarTeamsSkeleton } from "./components/sidebar-teams";
import { SidebarUserSkeleton } from "./components/sidebar-user";
import { routeTree } from "./routeTree.gen";

// const areQueryKeysEqual = (
//   a: readonly unknown[],
//   b: readonly unknown[]
// ): boolean => {
//   if (a.length !== b.length) return false;
//   return a.every((val, idx) => {
//     if (Array.isArray(val) && Array.isArray(b[idx])) {
//       return areQueryKeysEqual(val, b[idx]);
//     }
//     return val === b[idx];
//   });
// };

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: (query) => {
        // if (areQueryKeysEqual(getCurrentUserQueryKey(), query.queryKey)) {
        //   return Infinity;
        // }

        // if (areQueryKeysEqual(getSessionQueryKey(), query.queryKey)) {
        //   return Infinity;
        // }

        return 5 * 60 * 1000;
      },
    },
  },
});

const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: "render",
  defaultPreloadStaleTime: 0,
  scrollRestoration: true,
  defaultPendingComponent: () => (
    <SidebarProvider>
      <Sidebar collapsible="icon" className="z-20">
        <SidebarHeader>
          <ProjectSelectSkeleton />
        </SidebarHeader>

        <SidebarContent>
          <SidebarMainLinks
            _Link={(props: PropsWithChildren<any>) => (
              <p {...props} className={cn("cursor-pointer", props.className)} />
            )}
            projectSlug={""}
          />
          <SidebarTeamsSkeleton />
        </SidebarContent>

        <SidebarFooter>
          <SidebarUserSkeleton />
        </SidebarFooter>

        <SidebarRail />
      </Sidebar>
    </SidebarProvider>
  ),
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>
  );
}
