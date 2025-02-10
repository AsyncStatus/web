import { Suspense } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarProvider,
  SidebarRail,
} from "@asyncstatus/ui/components/sidebar.js";
import { createFileRoute, Link, Outlet } from "@tanstack/react-router";

import {
  ProjectSelect,
  ProjectSelectSkeleton,
} from "../../components/project-select";
import { SidebarMainLinks } from "../../components/sidebar-main-links";
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
