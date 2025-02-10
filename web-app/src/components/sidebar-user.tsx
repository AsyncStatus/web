"use client";

import {
  getCurrentUserOptions,
  logoutMutation,
} from "@asyncstatus/sdk/@tanstack/react-query.gen";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@asyncstatus/ui/components/avatar.js";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@asyncstatus/ui/components/dropdown-menu.js";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@asyncstatus/ui/components/sidebar.js";
import { Skeleton } from "@asyncstatus/ui/components/skeleton.js";
import { getInitials } from "@asyncstatus/ui/lib/utils.js";
import { User } from "@phosphor-icons/react";
import { Bell } from "@phosphor-icons/react/Bell";
import { CaretUpDown } from "@phosphor-icons/react/CaretUpDown";
import { CreditCard } from "@phosphor-icons/react/CreditCard";
import { SignOut } from "@phosphor-icons/react/SignOut";
import { Sparkle } from "@phosphor-icons/react/Sparkle";
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

export function SidebarUser(props: { projectSlug: string }) {
  const { isMobile } = useSidebar();
  const user = useSuspenseQuery(getCurrentUserOptions());

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="size-8">
                <AvatarImage
                  src={
                    user.data.avatar_url
                      ? `https://cdn.asyncstatus.com/${user.data.avatar_url}`
                      : undefined
                  }
                  alt={user.data.name}
                />
                <AvatarFallback>{getInitials(user.data.name)}</AvatarFallback>
              </Avatar>

              <div className="grid flex-1 pt-1 text-left text-sm leading-3">
                <span className="truncate font-semibold">{user.data.name}</span>
                <span className="truncate text-xs text-muted-foreground">
                  {user.data.email}
                </span>
              </div>

              <CaretUpDown className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>

          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="size-8">
                  <AvatarImage
                    src={
                      user.data.avatar_url
                        ? `https://cdn.asyncstatus.com/${user.data.avatar_url}`
                        : undefined
                    }
                    alt={user.data.name}
                  />
                  <AvatarFallback>{getInitials(user.data.name)}</AvatarFallback>
                </Avatar>

                <div className="grid flex-1 pt-1 text-left text-sm leading-3">
                  <span className="truncate font-semibold">
                    {user.data.name}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {user.data.email}
                  </span>
                </div>
              </div>
            </DropdownMenuLabel>

            <DropdownMenuSeparator />

            <UserDropdownItems
              userId={user.data.id}
              projectSlug={props.projectSlug}
            />
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

function UserDropdownItems(props: { projectSlug: string; userId: string }) {
  const queryClient = useQueryClient();
  const logout = useMutation({
    ...logoutMutation(),
    onSuccess: () => {
      queryClient.resetQueries();
    },
  });
  return (
    <>
      <DropdownMenuGroup>
        <DropdownMenuItem>
          <Sparkle />
          Upgrade plan
        </DropdownMenuItem>
      </DropdownMenuGroup>

      <DropdownMenuSeparator />

      <DropdownMenuGroup>
        <DropdownMenuItem asChild>
          {/* @ts-ignore */}
          <Link to={`/${props.projectSlug}/users/${props.userId}`}>
            <User />
            Account
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem>
          <CreditCard />
          Billing
        </DropdownMenuItem>

        <DropdownMenuItem>
          <Bell />
          Notifications
        </DropdownMenuItem>
      </DropdownMenuGroup>

      <DropdownMenuSeparator />

      <DropdownMenuItem
        onClick={() => logout.mutate({})}
        disabled={logout.isPending}
      >
        <SignOut />
        Sign out
      </DropdownMenuItem>
    </>
  );
}

export function SidebarUserSkeleton() {
  return (
    <div className="flex items-center gap-1 p-2">
      <Skeleton className="size-8 rounded-full" />
      <div className="flex flex-col gap-1">
        <Skeleton className="h-3 w-20 rounded-sm" />
        <Skeleton className="h-3 w-24 rounded-sm" />
      </div>
      <CaretUpDown className="ml-auto" />
    </div>
  );
}
