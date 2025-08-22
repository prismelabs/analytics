const canvas = new OffscreenCanvas(512, 128);
const ctx = canvas.getContext("2d")!;

export function measureText(
  text: string,
  fontSize: number,
  fontWeight = 400,
): TextMetrics {
  const fontStyle = `${fontWeight} ${fontSize}px 'Arial'`;

  ctx.font = fontStyle;

  return ctx.measureText(text);
}
