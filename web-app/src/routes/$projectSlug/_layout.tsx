import { Suspense } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarProvider,
  SidebarRail,
} from "@asyncstatus/ui/components/sidebar.js";
import {
  createFileRoute,
  Link,
  Outlet,
  redirect,
} from "@tanstack/react-router";

import { authQueryOptions } from "../../auth";
import {
  ProjectSelect,
  ProjectSelectSkeleton,
} from "../../components/project-select";
import { SidebarMainLinks } from "../../components/sidebar-main-links";
import { SidebarSkeleton } from "../../components/sidebar-skeleton";
import {
  SidebarTeams,
  SidebarTeamsSkeleton,
} from "../../components/sidebar-teams";
import {
  SidebarUser,
  SidebarUserSkeleton,
} from "../../components/sidebar-user";

export const Route = createFileRoute("/$projectSlug/_layout")({
  component: RouteComponent,
  beforeLoad: async ({ location, context: { queryClient } }) => {
    const user = await queryClient
      .ensureQueryData(authQueryOptions())
      .catch(() => {});
    if (!user) {
      throw redirect({ to: "/login", search: { redirect: location.pathname } });
    }
  },
  pendingComponent: SidebarSkeleton,
});

function RouteComponent() {
  const params = Route.useParams();

  return (
    <>
      <SidebarProvider>
        <Sidebar collapsible="icon" className="z-20">
          <SidebarHeader>
            <Suspense fallback={<ProjectSelectSkeleton />}>
              <ProjectSelect projectSlug={params.projectSlug} />
            </Suspense>
          </SidebarHeader>

          <SidebarContent>
            <SidebarMainLinks _Link={Link} projectSlug={params.projectSlug} />
            <Suspense fallback={<SidebarTeamsSkeleton />}>
              <SidebarTeams projectSlug={params.projectSlug} />
            </Suspense>
          </SidebarContent>

          <SidebarFooter>
            <Suspense fallback={<SidebarUserSkeleton />}>
              <SidebarUser projectSlug={params.projectSlug} />
            </Suspense>
          </SidebarFooter>

          <SidebarRail />
        </Sidebar>

        <Outlet />
      </SidebarProvider>
    </>
  );
}
