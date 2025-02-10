import { PropsWithChildren } from "react";

import { SimpleHeader } from "./simple-header";

export function SimpleLayout(props: PropsWithChildren<{ href: string }>) {
  return (
    <>
      <SimpleHeader href={props.href} />
      <main className="container mx-auto flex flex-col items-center justify-center pt-32">
        {props.children}
      </main>
    </>
  );
}
