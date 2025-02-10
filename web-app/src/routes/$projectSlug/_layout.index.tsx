import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/$projectSlug/_layout/")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="p-4">
      <h1 className="text-lg">Welcome to AsyncStatus</h1>
    </div>
  );
}
