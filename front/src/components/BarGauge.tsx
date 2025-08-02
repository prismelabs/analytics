export default function BarGauge(
  { data }: { data: Array<{ label: string; value: number }> },
) {
  data.sort((a, b) => b.value - a.value);

  const min = data[data.length - 1].value;
  const max = data[0].value;

  return (
    <ul class="flex flex-col gap-2 text-system-fg">
      {data.map((d, i) => (
        <li
          key={i}
          class="rounded overflow-hidden hover:bg-page-bg cursor-pointer"
        >
          <div class="flex justify-between py-1 px-2">
            <span>{d.label}</span>
            <span>{d.value}</span>
          </div>
          <Bar percentage={(d.value - min) / max * 100} />
        </li>
      ))}
    </ul>
  );
}

function Bar({ percentage }: { percentage: number }) {
  return (
    <div class="h-0.5">
      <div
        style={`width: ${percentage}%`}
        class="h-0.5 bg-indigo-400"
      >
      </div>
    </div>
  );
}
