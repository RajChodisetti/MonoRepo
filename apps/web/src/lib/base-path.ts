// Next.js's basePath auto-prefixes next/link and next/router, but not plain
// fetch() calls to root-relative paths. Wrap those call sites with
// withBasePath() so they still work when this app is served under a path
// (e.g. https://api.tuvisolutions.com/admin) instead of at the domain root.
export const BASE_PATH = process.env.NEXT_PUBLIC_BASE_PATH || "";

export function withBasePath(path: string): string {
  return `${BASE_PATH}${path}`;
}
