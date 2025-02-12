import { signUpTokenMutation } from "@asyncstatus/sdk/@tanstack/react-query.gen";
import { zSignUpTokenBody } from "@asyncstatus/sdk/zod.gen";
import { Button } from "@asyncstatus/ui/components/button.js";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@asyncstatus/ui/components/form.js";
import { Input } from "@asyncstatus/ui/components/input.js";
import { Label } from "@asyncstatus/ui/components/label.js";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { router } from "../../router";

export const Route = createFileRoute("/(auth)/_layout/sign-up")({
  errorComponent: () => (
    <main className="container mx-auto flex flex-col items-center justify-center pt-32">
      <section
        aria-labelledby="not-found-title"
        className="mx-auto flex w-full flex-col items-center justify-center gap-2"
      >
        <div className="text-center">
          <h1
            id="not-found-title"
            className="mb-4 mt-4 text-balance text-5xl font-semibold"
          >
            Something went wrong
          </h1>
          <h2 className="text-muted-foreground">
            Please try again later or contact support.
          </h2>
        </div>

        <div className="mt-6">
          <Button asChild variant="secondary">
            <a href="mailto:support@asyncstatus.com">Contact support</a>
          </Button>
        </div>
      </section>
    </main>
  ),
  validateSearch: z.object({ token: z.string(), email: z.string().email() }),
  component: RouteComponent,
});

const schema = zSignUpTokenBody
  .extend({ repeatPassword: z.string() })
  .superRefine((data, ctx) => {
    if (data.password !== data.repeatPassword) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Passwords do not match",
      });
    }

    return data;
  });

function RouteComponent() {
  const queryClient = useQueryClient();
  const search = Route.useSearch();
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      password: "",
      repeatPassword: "",
      token: search.token,
      timezone: new Intl.DateTimeFormat().resolvedOptions().timeZone,
    },
  });
  const loginEmail = useMutation({
    ...signUpTokenMutation(),
    onSuccess: async () => {
      await router.invalidate();
      await queryClient.resetQueries();
      await router.navigate({ to: search.redirect ?? "/" });
    },
  });

  return (
    <>
      <Form {...form}>
        <form
          onSubmit={form.handleSubmit((values) =>
            loginEmail.mutate({ body: values })
          )}
          className="mx-auto flex max-w-[230px] flex-col items-center gap-4"
        >
          <h1 className="mb-6 text-xl font-medium">Create account</h1>

          <div className="w-full">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              placeholder="name@example.com"
              autoCapitalize="none"
              autoComplete="email"
              autoCorrect="off"
              readOnly
              disabled
              value={search.email}
            />
          </div>

          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem className="w-full">
                <FormLabel>Password</FormLabel>
                <FormControl>
                  <Input
                    type="password"
                    autoComplete="new-password"
                    placeholder="**********"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="repeatPassword"
            render={({ field }) => (
              <FormItem className="w-full">
                <FormLabel>Repeat password</FormLabel>
                <FormControl>
                  <Input
                    type="password"
                    autoComplete="new-password"
                    placeholder="**********"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          {loginEmail.error && (
            <p className="text-destructive text-sm font-medium">
              {loginEmail.error.message}
            </p>
          )}

          <Button
            className="w-full self-end"
            type="submit"
            disabled={loginEmail.isPending}
          >
            Create account
          </Button>
        </form>
      </Form>
      <script
        async
        src="https://www.googletagmanager.com/gtag/js?id=AW-16869850476"
      ></script>
      <script
        dangerouslySetInnerHTML={{
          __html: `window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());

gtag('config', 'AW-16869850476');`,
        }}
      />
      <script
        dangerouslySetInnerHTML={{
          __html: `gtag('event', 'conversion', {
  'send_to': 'AW-16869850476/tWzsCPyv6JwaEOz6lew-',
  'value': 3.0,
  'currency': 'USD',
  'transaction_id': ''
`,
        }}
      />
    </>
  );
}
