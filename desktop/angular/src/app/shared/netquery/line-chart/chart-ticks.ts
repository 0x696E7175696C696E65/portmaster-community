/** Explicit ticks avoid D3's approximate tick count overcrowding narrow charts. */
export function chartTimeTicks(from: Date, to: Date, width: number): Date[] {
  const count = Math.max(2, Math.min(6, Math.floor(width / 110)));
  const start = from.getTime();
  const span = to.getTime() - start;
  return Array.from({ length: count }, (_, index) => new Date(start + span * index / (count - 1)));
}

export function bandwidthTimeLabel(date: Date, now = Date.now()): string {
  const minutes = Math.floor(Math.max(0, now - date.getTime()) / 60000);
  return minutes < 1 ? '<1m ago' : `${minutes}m ago`;
}
