import { useState } from "preact/hooks";

export type Theme = "dark" | "light";

export default function useTheme() {
  const darkTheme = globalThis.matchMedia &&
    globalThis.matchMedia("(prefers-color-scheme: dark)").matches;
  const [theme, setTheme] = useState<Theme>(
    darkTheme ? "dark" : "light",
  );

  globalThis.matchMedia("(prefers-color-scheme: dark)").addEventListener(
    "change",
    (event) => {
      setTheme(event.matches ? "dark" : "light");
    },
  );

  return theme;
}
