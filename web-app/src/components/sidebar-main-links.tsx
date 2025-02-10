import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@asyncstatus/ui/components/sidebar.js";
import {
  CalendarDots,
  Gear,
  Plugs,
  Pulse,
  Sun,
  Users,
} from "@phosphor-icons/react";

const items = [
  {
    title: "Status updates",
    url: "",
    icon: Sun,
  },
  {
    title: "Live activity",
    url: "/live-activity",
    icon: Pulse,
  },
  {
    title: "Schedules",
    url: "/schedules",
    icon: CalendarDots,
  },
  {
    title: "Users",
    url: "/users",
    icon: Users,
  },
  {
    title: "Integrations",
    url: "/integrations",
    icon: Plugs,
  },
  {
    title: "Settings",
    url: "/settings",
    icon: Gear,
  },
];

export function SidebarMainLinks({
  projectSlug,
  _Link,
}: {
  projectSlug: string;
  _Link: any;
}) {
  return (
    <SidebarGroup>
      <SidebarGroupLabel>Project</SidebarGroupLabel>

      <SidebarMenu className="gap-0.5">
        {items.map((item) => (
          <SidebarMenuItem key={item.title}>
            <SidebarMenuButton asChild tooltip={item.title} className="text-sm">
              {/* @ts-ignore */}
              <_Link to={`/${projectSlug}${item.url}`}>
                {item.icon && <item.icon weight="bold" />}
                <span>{item.title}</span>
              </_Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
