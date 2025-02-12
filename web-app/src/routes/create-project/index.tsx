import {
  createProjectMutation,
  getUserProjectsQueryKey,
} from "@asyncstatus/sdk/@tanstack/react-query.gen";
import { CreateProjectBody, Project } from "@asyncstatus/sdk/types.gen";
import { zCreateProjectBody } from "@asyncstatus/sdk/zod.gen";
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
import { SimpleLayout } from "@asyncstatus/ui/components/simple-layout.js";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useForm } from "react-hook-form";

import { authQueryOptions } from "../../auth";
import { router } from "../../router";

export const Route = createFileRoute("/create-project/")({
  component: RouteComponent,
  pendingComponent: () => <p>Loading...</p>,
});

function RouteComponent() {
  const user = useSuspenseQuery(authQueryOptions());
  const queryClient = useQueryClient();
  const createProjectForm = useForm<CreateProjectBody>({
    resolver: zodResolver(zCreateProjectBody),
    defaultValues: { name: "" },
  });
  const createProject = useMutation({
    ...createProjectMutation(),
    onSuccess: (data) => {
      queryClient.setQueryData(
        getUserProjectsQueryKey({
          path: { userId: user.data!.id },
          query: { limit: 100 },
        }),
        (projects: Project[]) => [...(projects ?? []), data]
      );
      router.navigate({
        to: `/$projectSlug`,
        params: { projectSlug: data.slug },
      });
    },
  });

  return (
    <SimpleLayout href={import.meta.env.VITE_WEB_MARKETING_URL}>
      <Form {...createProjectForm}>
        <form
          className="mx-auto flex max-w-[230px] flex-col items-center gap-4"
          onSubmit={createProjectForm.handleSubmit((values) => {
            createProject.mutate({ body: values });
          })}
        >
          <h1 className="mb-6 text-xl font-medium">Create project</h1>

          <FormField
            control={createProjectForm.control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Project name</FormLabel>
                <FormControl>
                  <Input
                    type="text"
                    autoCapitalize="none"
                    autoComplete="name"
                    autoCorrect="off"
                    placeholder="Apple Inc."
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button type="submit" className="w-full">
            Create Project
          </Button>
        </form>
      </Form>
    </SimpleLayout>
  );
}
