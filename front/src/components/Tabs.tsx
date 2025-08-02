import { JSX } from "preact";
import { useState } from "preact/hooks";

export type Tab = {
  name: string;
  children: JSX.Element;
};

export default function Tabs({ tabs }: { tabs: Tab[] }) {
  if (tabs.length === 0) return;

  const [tab, setTab] = useState(tabs[0]);

  return (
    <div class="overflow-x-auto">
      <div class="flex gap-4 mb-2 whitespace-nowrap">
        {tabs.map((t) => (
          <h2
            key={t.name}
            onClick={() => setTab(t)}
            class={`select-none ${
              t === tab
                ? "font-semibold text-system-fg"
                : "hover:cursor-pointer"
            }`}
          >
            {t.name}
          </h2>
        ))}
      </div>
      {tab.children}
    </div>
  );
}
