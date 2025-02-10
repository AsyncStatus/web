import Image, { getImageProps, ImageProps } from "next/image";
import Link from "next/link";
import { AsyncStatusLogo } from "@asyncstatus/ui/components/async-status-logo.tsx";
import { Button } from "@asyncstatus/ui/components/button.tsx";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@asyncstatus/ui/components/dialog.tsx";
import { MagnifyingGlassPlus } from "@phosphor-icons/react/dist/ssr";

export default function NewPage() {
  const {
    props: { srcSet: light, ...rest },
  } = getImageProps({ ...common, src: "/hero-light.webp" });
  const {
    props: { srcSet: lightLg },
  } = getImageProps({ ...commonLg, src: "/hero-light-lg.webp" });
  return (
    <div className="mx-auto mt-24 w-full max-w-3xl p-4 pb-24 max-sm:mt-0">
      <AsyncStatusLogo className="h-3.5 w-auto" />
      <h1 className="mt-3.5 text-lg font-semibold">
        Async status updates for remote startups
      </h1>
      <h2 className="text-sm">
        Made for high-agency teams that value their time.
      </h2>

      <Dialog>
        <DialogTrigger asChild>
          <div className="group relative mt-4 w-full max-w-3xl rounded-md hover:opacity-60">
            <div className="absolute inset-0 z-20 flex items-center justify-center opacity-0 group-hover:opacity-100">
              <MagnifyingGlassPlus className="text-foreground bg-background/50 pointer-events-none size-8 rounded-full p-1 backdrop-blur-md" />
            </div>

            <picture className="relative z-10 rounded-[inherit]">
              <source media="(min-width: 1300px)" srcSet={lightLg} />
              <source media="(min-width: 768px)" srcSet={light} />
              <img
                {...rest}
                className="border-border h-auto w-full select-none rounded-[inherit] border object-contain"
              />
            </picture>
          </div>
        </DialogTrigger>

        <DialogContent
          hideCloseButton
          className="max-w-7xl border-0 bg-transparent p-0 p-4 outline-none sm:max-w-7xl"
        >
          <DialogTitle hidden aria-hidden>
            Demo image
          </DialogTitle>
          <DialogDescription hidden aria-hidden>
            See the platform
          </DialogDescription>

          <div className="relative w-full overflow-hidden rounded-md bg-transparent">
            <img
              src="/hero-light-lg.webp"
              sizes="100vw"
              alt="Demo image"
              className="h-auto w-full rounded-md object-contain"
            />
          </div>
        </DialogContent>
      </Dialog>

      <div className="mt-4 flex items-center justify-end gap-2">
        <Button asChild>
          <a href={process.env.NEXT_PUBLIC_STRIPE_LINK} target="_blank">
            Join private beta
          </a>
        </Button>
      </div>

      <h3 className="mb-1.5 mt-8 text-base font-semibold">How it works?</h3>
      <ol className="list-inside list-decimal text-sm">
        <li>
          You connect your tools (GitHub, Slack){" "}
          <span className="text-muted-foreground">
            which takes less than 5 minutes
          </span>
        </li>
        <li>
          We listen for updates from your tools and generate status updates from
          your team's activity.
        </li>
        <li>Your team can optionally adjust their status updates.</li>
      </ol>

      <h3 className="mb-1.5 mt-8 text-base font-semibold">Use cases</h3>
      <ul className="list-inside list-disc text-sm">
        <li>
          <span className="italic">Developers</span> - generate status based on
          your activity.
        </li>
        <li>
          <span className="italic">Product owners</span> - see summary of your
          team's work, spot blockers and mood changes.
        </li>
        <li>
          <span className="italic">Founders</span> - get insights from your team
          without unnecessary questions.
        </li>
      </ul>

      <h3 className="mb-1.5 mt-8 text-base font-semibold">Features</h3>
      <ul className="list-inside list-disc text-sm">
        <li>
          Integrations with GitHub and Slack.{" "}
          <span
            role="button"
            className="text-muted-foreground underline underline-offset-2"
          >
            Examples
          </span>
        </li>
        <li>
          Blockers and mood changes.{" "}
          <span
            role="button"
            className="text-muted-foreground underline underline-offset-2"
          >
            Examples
          </span>
        </li>
        <li>
          User timezones.{" "}
          <span
            role="button"
            className="text-muted-foreground underline underline-offset-2"
          >
            Examples
          </span>
        </li>
        <li>
          Slack bot that asks for scheduled updates and posts summaries.{" "}
          <span
            role="button"
            className="text-muted-foreground underline underline-offset-2"
          >
            Examples
          </span>
        </li>
        <li>
          Manual updates.{" "}
          <span
            role="button"
            className="text-muted-foreground underline underline-offset-2"
          >
            Examples
          </span>
        </li>
        <li>
          <Link
            target="_blank"
            href="https://github.com/asyncstatus/web"
            className="text-muted-foreground underline underline-offset-2"
          >
            Open source
          </Link>
        </li>
      </ul>
      {/* <ul className="text-sm list-disc list-inside">
        <li>
          Integrations with GitHub, Slack. We listen for updates from your tools and generate easy
          to read summary.{" "}
          <span role="button" className="text-muted-foreground underline">
            Click to see examples
          </span>
          .
        </li>
        <li>
          <span role="button" className="text-muted-foreground underline">
            Open source
          </span>
          .
        </li>
      </ul> */}
      {/* <h1 className="font-medium text-lg mt-1">Async status updates for remote startups</h1>
      <h2 className="text-sm">Built for high-agency teams working globally.</h2> */}
    </div>
  );
}

const common = {
  alt: "AsyncStatus App",
  width: 1300,
  height: 800,
  unoptimized: false,
  sizes: "100vw",
} satisfies Omit<ImageProps, "src">;
const commonLg = { ...common, width: 2666, height: 1500 } satisfies Omit<
  ImageProps,
  "src"
>;
