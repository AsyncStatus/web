import { Button } from "@asyncstatus/ui/components/button.js";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/$projectSlug/_layout/")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="p-4">
      <h1 className="text-lg">Home</h1>
      <Button asChild>
        <a href="https://buy.stripe.com/test_dR6bL20rL1aZaPucMM">
          Join private beta
        </a>
      </Button>
    </div>
  );
}
