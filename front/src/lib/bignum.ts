export function format(num: number): string {
    const absNum = Math.abs(num);
    if (absNum < 1100) return num.toString();

    const suffixes = ['K', 'M', 'B', 'T'];
    const suffixIndex = Math.floor(Math.log10(absNum) / 3) - 1;
    const shortNum = (num / Math.pow(1000, suffixIndex + 1)).toFixed(1);

    return `${shortNum}${suffixes[suffixIndex]}`;
}
