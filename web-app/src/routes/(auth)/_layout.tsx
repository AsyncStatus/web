import { AsyncStatusLogo } from "@asyncstatus/ui/components/async-status-logo.js";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/(auth)/_layout")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <>
      <header className="flex items-center justify-center p-6">
        <nav>
          <a
            className="flex items-center gap-0.5"
            href="http://localhost:3000"
            aria-label="AsyncStatus Home"
          >
            <AsyncStatusLogo className="h-4 w-auto" />
          </a>
        </nav>
      </header>

      <Outlet />
    </>
  );
}
