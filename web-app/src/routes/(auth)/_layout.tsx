import { SimpleHeader } from "@asyncstatus/ui/components/simple-header.js";
import {
  createFileRoute,
  isRedirect,
  Outlet,
  redirect,
} from "@tanstack/react-router";
import { z } from "zod";

import { authQueryOptions } from "../../auth";

export const Route = createFileRoute("/(auth)/_layout")({
  validateSearch: z.object({ redirect: z.string().optional().catch("") }),
  beforeLoad: ({ context: { queryClient } }) => {
    queryClient
      .ensureQueryData(authQueryOptions())
      .then(() => {
        throw redirect({ to: "/" });
      })
      .catch((error) => {
        if (isRedirect(error)) {
          throw error;
        }
      });
  },
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <>
      <SimpleHeader href={import.meta.env.VITE_WEB_MARKETING_URL} />
      <Outlet />
    </>
  );
}
