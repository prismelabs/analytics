import { signal } from "@preact/signals";

export type ColorScheme = "dark" | "light";

export const preferred = signal<ColorScheme>(
  globalThis.matchMedia &&
    globalThis.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light",
);

globalThis.matchMedia("(prefers-color-scheme: dark)").addEventListener(
  "change",
  (ev) => {
    preferred.value = ev.matches ? "dark" : "light";
  },
);
