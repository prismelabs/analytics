import { JSX } from "preact";
import UplotReact from "uplot-react";
import "uplot/dist/uPlot.min.css";

export default function (props: {
  options: uPlot.Options;
  data: uPlot.AlignedData;
  target?: HTMLElement;
  onDelete?: (chart: uPlot) => void;
  onCreate?: (chart: uPlot) => void;
  resetScales?: boolean;
  className?: string;
}) {
  // @ts-ignore: ...
  return <UplotReact {...props} /> as unknown as JSX.Element;
}
