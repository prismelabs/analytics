import { JSX } from "preact";
import { useState } from "preact/hooks";

export type Tab = {
  name: string;
  children: JSX.Element;
};

export default function Tabs({ tabs }: { tabs: Tab[] }) {
  const [active, setTab] = useState(0);

  return (
    <div class="overflow-x-auto overflow-y-hidden h-full max-h-full">
      <div class="flex gap-4 mb-2 whitespace-nowrap">
        {tabs.map((t, i) => (
          <h2
            key={t.name}
            onClick={() => setTab(i)}
            class={`select-none ${
              t.name === tabs[active].name
                ? "font-semibold text-system-fg"
                : "hover:cursor-pointer"
            }`}
          >
            {t.name}
          </h2>
        ))}
      </div>
      {tabs[active]?.children}
    </div>
  );
}
