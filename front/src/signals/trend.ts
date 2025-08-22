import { computed, effect, signal } from "@preact/signals";

import { preferred as colorScheme } from "@/signals/color-scheme.ts";

export type Trend = "neutral" | "upward" | "downward";

export const trend = signal<Trend>("neutral");

type Theme = {
  light: {
    page: string;
    background: string;
    foreground: string;
    primary: string;
    "primary-light": string;
  };
  dark: {
    page: string;
    background: string;
    foreground: string;
    primary: string;
    "primary-light": string;
  };
};

const themes: Record<Trend, Theme> = {
  neutral: {
    light: {
      page: "#f5f5f5",
      background: "white",
      foreground: "#707088",
      primary: "#60a5fa",
      "primary-light": "#bfdbfe",
    },
    dark: {
      page: "#05050a",
      background: "#202026",
      foreground: "#b0b0f7",
      primary: "#60a5fa",
      "primary-light": "#1e3a8a",
    },
  },
  upward: {
    light: {
      page: "#f0f5f0",
      background: "white",
      foreground: "#709770",
      primary: "#10b981",
      "primary-light": "#6ee7b7",
    },
    dark: {
      page: "#050a05",
      background: "#202620",
      foreground: "#b0f7b0",
      primary: "#10b981",
      "primary-light": "#047857",
    },
  },
  downward: {
    light: {
      page: "#f5f0f0",
      background: "white",
      foreground: "#977070",
      primary: "#f43f5e",
      "primary-light": "#fda4af",
    },
    dark: {
      page: "#0a0505",
      background: "#262020",
      foreground: "#f7b0b0",
      primary: "#f43f5e",
      "primary-light": "#881337",
    },
  },
};

export const theme = computed(() => themes[trend.value][colorScheme.value]);

effect(() => {
  const doc = globalThis.document.documentElement;
  for (const [name, color] of Object.entries(theme.value)) {
    doc.style.setProperty("--color-trend-" + name, color);
  }
});
