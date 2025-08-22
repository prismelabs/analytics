import { RefObject } from "preact";
import { useEffect, useState } from "preact/hooks";

export default function (ref: RefObject<Element>) {
  const [size, setSize] = useState({ width: 0, height: 0 });

  useEffect(() => {
    const obs = new ResizeObserver((entries) => {
      for (const ent of entries) {
        const size = ent.contentBoxSize[0];
        setSize({ width: size.inlineSize, height: size.blockSize });
        break;
      }
    });
    if (ref.current) obs.observe(ref.current);

    return () => obs.disconnect();
  }, [ref.current]);

  return size;
}
