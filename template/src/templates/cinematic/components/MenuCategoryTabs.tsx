"use client";

import type { MenuCategory } from "@/data/types/menu";
import { shortenCategoryLabel } from "../lib/menuImages";

interface MenuCategoryTabsProps {
  categories: MenuCategory[];
  active: string;
  onChange: (name: string) => void;
}

export default function MenuCategoryTabs({
  categories,
  active,
  onChange,
}: MenuCategoryTabsProps) {
  return (
    <div className="sticky top-[72px] z-30 -mx-6 mb-8 overflow-x-auto border-b border-cream/10 bg-[#141210]/98 px-6 py-4 backdrop-blur-md">
      <div className="flex min-w-max max-w-full gap-2 pb-1">
        <button
          type="button"
          onClick={() => onChange("all")}
          className={`shrink-0 rounded-full border px-4 py-2 text-[10px] uppercase tracking-wider transition ${
            active === "all"
              ? "border-brass bg-brass/10 text-brass"
              : "border-cream/15 text-cream/50 hover:border-brass/40"
          }`}
        >
          All
        </button>
        {categories.map((cat) => {
          const label = shortenCategoryLabel(cat.name);
          return (
            <button
              key={cat.name}
              type="button"
              title={cat.name}
              onClick={() => onChange(cat.name)}
              className={`max-w-[160px] shrink-0 truncate rounded-full border px-4 py-2 text-[10px] uppercase tracking-wider transition ${
                active === cat.name
                  ? "border-brass bg-brass/10 text-brass"
                  : "border-cream/15 text-cream/50 hover:border-brass/40"
              }`}
            >
              {label}
            </button>
          );
        })}
      </div>
    </div>
  );
}
