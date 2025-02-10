"use client";

import { getProjectTeamsOptions } from "@asyncstatus/sdk/@tanstack/react-query.gen";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@asyncstatus/ui/components/dropdown-menu.js";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@asyncstatus/ui/components/sidebar.js";
import { Skeleton } from "@asyncstatus/ui/components/skeleton.js";
import { DotsThree, Pen, Plus } from "@phosphor-icons/react";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

export function SidebarTeams({ projectSlug }: { projectSlug: string }) {
  const { isMobile } = useSidebar();
  const teams = useSuspenseQuery(
    getProjectTeamsOptions({ path: { projectSlug: projectSlug } })
  );

  return (
    <SidebarGroup className="group-data-[collapsible=icon]:hidden">
      <SidebarGroupLabel>
        {/* @ts-ignore */}
        <Link to={`/${projectSlug}/teams`}>
          <span>Teams</span>
        </Link>
      </SidebarGroupLabel>

      <SidebarMenu className="gap-0.5">
        {teams.data.length === 0 && (
          <SidebarMenuItem>
            <SidebarMenuButton className="gap-0.5">
              <Plus />
              <span>New</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        )}

        {teams.data.map((team) => (
          <SidebarMenuItem key={team.slug}>
            <SidebarMenuButton asChild>
              {/* @ts-ignore */}
              <Link to={`/${projectSlug}/teams/${team.slug}`}>
                {team.emoji && <span>{team.emoji}</span>}
                <span>{team.name}</span>
              </Link>
            </SidebarMenuButton>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuAction showOnHover>
                  <DotsThree />
                  <span className="sr-only">More</span>
                </SidebarMenuAction>
              </DropdownMenuTrigger>

              <DropdownMenuContent
                className="w-48"
                side={isMobile ? "bottom" : "right"}
                align={isMobile ? "end" : "start"}
              >
                <DropdownMenuItem asChild>
                  {/* @ts-ignore */}
                  <Link to={`/${projectSlug}/teams/${team.slug}/edit`}>
                    <Pen className="text-muted-foreground" />
                    <span>Edit team</span>
                  </Link>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}

export function SidebarTeamsSkeleton() {
  return (
    <SidebarGroup className="group-data-[collapsible=icon]:hidden">
      <SidebarGroupLabel>
        <span>Teams</span>
      </SidebarGroupLabel>

      <SidebarMenu className="gap-0.5">
        {Array.from({ length: 4 }).map((_, index) => (
          <SidebarMenuItem key={index}>
            <SidebarMenuButton asChild>
              <Skeleton className="h-8 w-full" />
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
