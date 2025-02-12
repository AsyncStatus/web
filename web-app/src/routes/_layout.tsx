import { getUserProjectsOptions } from "@asyncstatus/sdk/@tanstack/react-query.gen";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

import { authQueryOptions } from "../auth";
import { SidebarSkeleton } from "../components/sidebar-skeleton";

export const Route = createFileRoute("/_layout")({
  component: RouteComponent,
  pendingComponent: SidebarSkeleton,
  beforeLoad: async ({ context: { queryClient } }) => {
    const user = await queryClient
      .ensureQueryData(authQueryOptions())
      .catch(() => {});
    if (!user) {
      throw redirect({ to: "/login" });
    }

    const projects = await queryClient.ensureQueryData(
      getUserProjectsOptions({
        path: { userId: user!.id },
        query: { limit: 100 },
      })
    );
    if (projects.length === 0) {
      throw redirect({ to: "/create-project" });
    }

    throw redirect({
      to: `/$projectSlug`,
      params: { projectSlug: projects[0].slug },
    });
  },
});

function RouteComponent() {
  return <Outlet />;
}
