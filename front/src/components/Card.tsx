import { JSX } from "preact";

export default function Card(
  { title, class: className = "", size: size = "big", children, onClick }: {
    title?: string;
    class?: string;
    size: "small" | "big";
    children: JSX.Element | JSX.Element[];
    onClick?: (_: MouseEvent) => void;
  },
) {
  return (
    <section
      class={`bg-trend-background rounded-2xl p-4 ${className}`}
      onClick={onClick}
    >
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
