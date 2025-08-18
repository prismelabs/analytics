import { JSX } from "preact";

export default function Card(
  { title, class: className = "", size: size = "big", children }: {
    title?: string;
    class?: string;
    size: "small" | "big";
    children: JSX.Element | JSX.Element[];
  },
) {
  return (
    <section class={`bg-trend-background rounded p-4 ${className}`}>
      {!title ? null : size === "big"
        ? (
          <h2 class="text-md font-semibold mb-2">
            {title}
          </h2>
        )
        : (
          <h2 class="text-center font-semibold tracking-wide whitespace-nowrap">
            {title}
          </h2>
        )}
      {children}
    </section>
  );
}
