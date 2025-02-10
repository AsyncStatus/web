import { PropsWithChildren } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarProvider,
  SidebarRail,
} from "@asyncstatus/ui/components/sidebar.js";
import { cn } from "@asyncstatus/ui/lib/utils.js";

import { ProjectSelectSkeleton } from "./project-select";
import { SidebarMainLinks } from "./sidebar-main-links";
import { SidebarTeamsSkeleton } from "./sidebar-teams";
import { SidebarUserSkeleton } from "./sidebar-user";

export function SidebarSkeleton(props: PropsWithChildren) {
  return (
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

      {props.children}
    </SidebarProvider>
  );
}
