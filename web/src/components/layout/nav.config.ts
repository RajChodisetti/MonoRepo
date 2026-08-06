import type { NavLink } from "@/components/layout/NavLinks";

/** Shared nav config — keep labels/routes in one place */
export const primaryNavLinks: NavLink[] = [
  { type: "mega", label: "Product" },
  { type: "link", label: "Pricing", href: "/pricing" },
  { type: "link", label: "How it works", href: "/how-it-works" },
];
