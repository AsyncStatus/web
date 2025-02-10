import { ASError } from "@asyncstatus/sdk/client.gen";
import { Button } from "@asyncstatus/ui/components/button.js";
import { PaperPlaneTilt } from "@phosphor-icons/react";

interface ErrorFallbackProps {
  error: Error;
  reset: () => void;
}

export function ErrorFallback(props: ErrorFallbackProps) {
  const isAsError = "internal_status" in props.error;
  if (isAsError) {
    return <AsErrorFallback {...props} />;
  }

  const isFailedToFetchError =
    props.error instanceof TypeError &&
    props.error.message.includes("Failed to fetch");
  if (isFailedToFetchError) {
    return <FailedToFetchErrorFallback {...props} />;
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-4 rounded-md border-2 border-dashed border-neutral-200 bg-neutral-50 p-2.5 px-3.5">
      <h3 className="text-base font-semibold">Something went wrong</h3>
      <div className="flex items-center gap-1">
        <Button
          size="sm"
          className="max-sm:w-full"
          onClick={() => props.reset()}
        >
          Try again
        </Button>

        <Button
          size="sm"
          variant="ghost"
          className="max-sm:w-full"
          onClick={() => {
            const a = document.createElement("a");
            const errorBodyTemplate = `Could you please take a look at the error above and provide us with feedback?%0D%0A
Project name:
%0D%0A
Any additional information you'd like to provide:
%0D%0A
%0D%0A
%0D%0A
%0D%0A
%0D%0A
--- Error:
%0D%0A
${JSON.stringify(props.error ?? {})}`;
            a.href = `mailto:support@asyncstatus.com?subject=${"Something went wrong"}&body=${errorBodyTemplate}`;
            a.click();
            a.remove();
          }}
        >
          <PaperPlaneTilt />
          Report critical error
        </Button>
      </div>
    </div>
  );
}

function FailedToFetchErrorFallback(props: ErrorFallbackProps) {
  const error = props.error as TypeError;

  return (
    <div className="rounded-md border-2 border-dashed border-neutral-200 bg-neutral-50 p-2.5 px-3.5">
      <h3 className="text-base font-semibold">Fetch failed</h3>
      <p className="mt-0.5 text-sm text-muted-foreground">
        It looks like we're having some issues fetching the data. Please try
        again later. We're working on it.
      </p>

      <div className="mt-2 flex flex-wrap items-center gap-1 max-sm:mt-6">
        <Button
          size="sm"
          className="max-sm:w-full"
          onClick={() => props.reset()}
        >
          Try again
        </Button>

        <Button
          size="sm"
          variant="ghost"
          className="max-sm:w-full"
          onClick={() => {
            const a = document.createElement("a");
            const errorBodyTemplate = `Could you please take a look at the error above and provide us with feedback?%0D%0A
Project name:
%0D%0A
Any additional information you'd like to provide:
%0D%0A
%0D%0A
%0D%0A
%0D%0A
%0D%0A
--- Error:
%0D%0A
${JSON.stringify(error)}`;
            a.href = `mailto:support@asyncstatus.com?subject=${error.message}&body=${errorBodyTemplate}`;
            a.click();
            a.remove();
          }}
        >
          <PaperPlaneTilt />
          Report critical error
        </Button>
      </div>
    </div>
  );
}

function AsErrorFallback(props: ErrorFallbackProps) {
  const error = props.error as unknown as ASError;

  return (
    <div className="rounded-md border-2 border-dashed border-neutral-200 bg-neutral-50 p-2.5 px-3.5">
      <h3 className="text-base font-semibold">
        {error.message}
        <span className="ml-1 rounded-full bg-primary/10 px-1.5 text-xs text-primary">
          {error.internal_status}
        </span>
      </h3>
      {error.internal_status.startsWith("ASE-40") && error.errors_map ? (
        <div className="mt-1 mb-2 flex flex-col overflow-auto rounded-md bg-neutral-200/50 break-all">
          <ObjectComponents data={error.errors_map} />
        </div>
      ) : (
        <p className="mt-0.5 text-sm text-muted-foreground">
          {error.details}
          {error.internal_status.startsWith("ASE-40") &&
          !error.internal_status.startsWith("ASE-401") &&
          !error.internal_status.startsWith("ASE-429")
            ? "Please try again later. We're working on it."
            : ""}
        </p>
      )}

      <div className="mt-2 flex flex-wrap items-center gap-1 max-sm:mt-6">
        <Button
          size="sm"
          className="max-sm:w-full"
          onClick={() => props.reset()}
        >
          Try again
        </Button>

        <Button
          size="sm"
          variant="ghost"
          className="max-sm:w-full"
          onClick={() => {
            const a = document.createElement("a");
            const errorBodyTemplate = `Could you please take a look at the error above and provide us with feedback?%0D%0A
Project name:
%0D%0A
Any additional information you'd like to provide:
%0D%0A
%0D%0A
%0D%0A
%0D%0A
%0D%0A
--- Error:
%0D%0A
${JSON.stringify(error)}`;
            a.href = `mailto:support@asyncstatus.com?subject=${error.message}&body=${errorBodyTemplate}`;
            a.click();
            a.remove();
          }}
        >
          <PaperPlaneTilt />
          Report critical error
        </Button>
      </div>
    </div>
  );
}

export function ObjectComponents({ data }: { data: Record<string, string> }) {
  return Object.entries(data).map(([key, value]) => (
    <div
      key={key}
      className="group flex items-center gap-0.5 border-b border-border p-2 py-1 text-xs"
    >
      <span className="font-medium text-neutral-500 group-hover:text-neutral-800">
        {key}:
      </span>{" "}
      <span className="font-medium text-neutral-800 group-hover:text-neutral-900">
        {value}
      </span>
    </div>
  ));
}
