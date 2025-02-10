import { Button } from "@asyncstatus/ui/components/button.js";
import { SimpleHeader } from "@asyncstatus/ui/components/simple-header.js";

export function NotFound() {
  return (
    <>
      <SimpleHeader href={import.meta.env.VITE_WEB_MARKETING_URL} />

      <main className="container mx-auto flex flex-col items-center justify-center pt-32">
        <section
          aria-labelledby="not-found-title"
          className="mx-auto flex w-full flex-col items-center justify-center gap-2"
        >
          <div className="text-center">
            <h1
              id="not-found-title"
              className="mt-4 mb-4 text-5xl font-semibold text-balance"
            >
              Not found
            </h1>
            <h2 className="text-muted-foreground">
              We couldn't find the page you were looking for.
            </h2>
          </div>

          <div className="mt-6">
            <Button asChild variant="secondary">
              <a href="mailto:support@asyncstatus.com">Contact support</a>
            </Button>
          </div>
        </section>
      </main>
    </>
  );
}
