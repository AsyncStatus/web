import { useMemo } from "react";
import {
  getCurrentUserOptions,
  getUserProjectsOptions,
} from "@asyncstatus/sdk/@tanstack/react-query.gen";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@asyncstatus/ui/components/avatar.js";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@asyncstatus/ui/components/dropdown-menu.js";
import { ScrollArea } from "@asyncstatus/ui/components/scroll-area.js";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@asyncstatus/ui/components/sidebar.js";
import { Skeleton } from "@asyncstatus/ui/components/skeleton.js";
import { getInitials, upperFirst } from "@asyncstatus/ui/lib/utils.js";
import { CaretUpDown, Plus } from "@phosphor-icons/react";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

export function ProjectSelect(props: { projectSlug: string }) {
  const { isMobile } = useSidebar();
  const user = useSuspenseQuery(getCurrentUserOptions());
  const projects = useSuspenseQuery(
    getUserProjectsOptions({
      query: { limit: 100 },
      path: { userId: user.data!.id },
    })
  );
  const activeProject = useMemo(
    () => projects.data.find((p) => p.slug === props.projectSlug),
    [projects.data, props.projectSlug]
  );

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground h-auto p-1"
            >
              <Avatar className="size-8">
                <AvatarImage
                  src={
                    activeProject?.avatar_url
                      ? `https://cdn.asyncstatus.com/${activeProject.avatar_url}`
                      : undefined
                  }
                  alt={activeProject?.name}
                />
                <AvatarFallback>
                  {getInitials(activeProject?.name ?? "")}
                </AvatarFallback>
              </Avatar>

              <div className="grid flex-1 text-left text-sm">
                <span className="truncate">{activeProject?.name}</span>
              </div>

              <CaretUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>

          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56"
            align="start"
            side={isMobile ? "bottom" : "right"}
            sideOffset={4}
          >
            <DropdownMenuLabel className="text-muted-foreground text-xs">
              Projects
            </DropdownMenuLabel>

            <ScrollArea className="h-36 w-full">
              {projects.data.map((project) => (
                <DropdownMenuItem
                  asChild
                  key={project.name}
                  className="gap-2 p-2"
                >
                  <Link
                    to={`/$projectSlug`}
                    params={{ projectSlug: project.slug }}
                  >
                    <Avatar className="size-8">
                      <AvatarImage
                        src={
                          project.avatar_url
                            ? `https://cdn.asyncstatus.com/${project.avatar_url}`
                            : undefined
                        }
                        alt={project.name}
                      />
                      <AvatarFallback>
                        {getInitials(project.name)}
                      </AvatarFallback>
                    </Avatar>

                    <div className="grid flex-1 pt-1 text-left text-sm leading-3">
                      <span className="truncate font-semibold">
                        {project.name}
                      </span>
                      <span className="text-muted-foreground truncate text-xs">
                        {upperFirst(project.plan)}
                      </span>
                    </div>
                  </Link>
                </DropdownMenuItem>
              ))}
            </ScrollArea>

            <DropdownMenuSeparator />

            <DropdownMenuItem className="gap-2 p-2">
              <div className="bg-background flex size-6 items-center justify-center rounded-md border">
                <Plus className="size-4" />
              </div>

              <div className="text-muted-foreground font-medium">
                Create project
              </div>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

export function ProjectSelectSkeleton() {
  return (
    <div className="flex items-center p-1">
      <Skeleton className="size-8 rounded-full" />
      <CaretUpDown className="ml-auto" />
    </div>
  );
}
