import {
  createProjectMutation,
  getUserProjectsQueryKey,
} from "@asyncstatus/sdk/@tanstack/react-query.gen";
import { CreateProjectBody, Project } from "@asyncstatus/sdk/types.gen";
import { zCreateProjectBody } from "@asyncstatus/sdk/zod.gen";
import { Button } from "@asyncstatus/ui/components/button.js";
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
          path: { user_id: user.data!.id },
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
      <form
        onSubmit={createProjectForm.handleSubmit((values) => {
          createProject.mutate({ body: values });
        })}
      >
        <div>
          <label htmlFor="name">Name</label>
          <input
            type="text"
            id="name"
            autoCapitalize="none"
            autoComplete="name"
            autoCorrect="off"
            {...createProjectForm.register("name")}
          />
        </div>
        <Button type="submit">Create Project</Button>
      </form>
    </SimpleLayout>
  );
}
