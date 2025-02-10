import { signInEmailMutation } from "@asyncstatus/sdk/@tanstack/react-query.gen";
import { SignInEmailBody } from "@asyncstatus/sdk/types.gen";
import { zSignInEmailBody } from "@asyncstatus/sdk/zod.gen";
import { Button } from "@asyncstatus/ui/components/button.js";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { router } from "../../router";

export const Route = createFileRoute("/(auth)/_layout/login")({
  validateSearch: z.object({ redirect: z.string().optional().catch("") }),
  component: RouteComponent,
});

function RouteComponent() {
  const queryClient = useQueryClient();
  const loginForm = useForm<SignInEmailBody>({
    resolver: zodResolver(zSignInEmailBody),
    defaultValues: { email: "", password: "" },
  });
  const search = Route.useSearch();
  const loginEmail = useMutation({
    ...signInEmailMutation(),
    onSuccess: async () => {
      await router.invalidate();
      await queryClient.resetQueries();
      await router.navigate({ to: search.redirect ?? "/" });
    },
  });

  return (
    <div>
      <form
        onSubmit={loginForm.handleSubmit((values) =>
          loginEmail.mutate({ body: values })
        )}
        className="flex flex-col items-center gap-2 [&>input]:max-w-64 [&>input]:border"
      >
        <input
          className="rounded-md p-1 px-2.5"
          type="email"
          placeholder="name@example.com"
          autoCapitalize="none"
          autoComplete="email"
          autoCorrect="off"
          {...loginForm.register("email")}
        />
        <input
          className="rounded-md p-1 px-2.5"
          type="password"
          autoComplete="current-password"
          placeholder="**********"
          {...loginForm.register("password")}
        />

        <Button type="submit" disabled={loginEmail.isPending}>
          Login
        </Button>
      </form>

      {loginEmail.error && <p>{loginEmail.error.message}</p>}
    </div>
  );
}
