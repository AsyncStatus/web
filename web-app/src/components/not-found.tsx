import { AsyncStatusLogo } from "@asyncstatus/ui/components/async-status-logo.js";
import { Button } from "@asyncstatus/ui/components/button.js";
import { Link } from "@tanstack/react-router";

export function NotFound() {
  return (
    <>
      <header className="flex items-center justify-center p-6">
        <nav>
          <Link
            className="flex items-center gap-0.5"
            to="/"
            aria-label="AsyncStatus Home"
          >
            <AsyncStatusLogo className="h-4 w-auto" />
          </Link>
        </nav>
      </header>

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
