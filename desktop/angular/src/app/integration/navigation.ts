export type NavigationTarget = { kind: 'internal' | 'blocked' } | { kind: 'external'; url: string };

/** Classify links before crossing into a native URI opener. */
export function navigationTarget(href: string, base: string): NavigationTarget {
  try {
    const url = new URL(href, base);
    const origin = new URL(base).origin;
    if (url.username || url.password) {
      return { kind: 'blocked' };
    }
    if (url.protocol === 'blob:') {
      return { kind: url.origin === origin ? 'internal' : 'blocked' };
    }
    if (url.protocol !== 'https:' && url.protocol !== 'http:') {
      return { kind: 'blocked' };
    }
    return url.origin === origin ? { kind: 'internal' } : { kind: 'external', url: url.href };
  } catch {
    return { kind: 'blocked' };
  }
}
