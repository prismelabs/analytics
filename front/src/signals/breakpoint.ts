import { signal } from "@preact/signals";

// Tailwind CSS breakpoints.
export type Breakpoint = "sm" | "md" | "lg" | "xl" | "2xl";

export const breakpoint = signal<Breakpoint>(findBreakPoint());

function findBreakPoint(): Breakpoint {
  if (globalThis.innerWidth <= 640) {
    return "sm";
  } else if (globalThis.innerWidth <= 768) {
    return "md";
  } else if (globalThis.innerWidth <= 1024) {
    return "lg";
  } else if (globalThis.innerWidth <= 1280) {
    return "xl";
  } else {
    return "2xl";
  }
}

globalThis.addEventListener("resize", () => {
  breakpoint.value = findBreakPoint();
});
