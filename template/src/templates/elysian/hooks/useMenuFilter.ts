"use client";

import { useCallback, useMemo, useState } from "react";
import type { ElysianMenuItem } from "../lib/mapContent";

export function useMenuFilter(items: ElysianMenuItem[], tabs: { id: string; label: string }[]) {
  const [query, setQuery] = useState("");
  const [activeFilter, setActiveFilter] = useState("all");

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();
    return items.filter((item) => {
      const matchesCat = activeFilter === "all" || item.categories.includes(activeFilter);
      const matchesSearch =
        !term ||
        item.name.toLowerCase().includes(term) ||
        item.description.toLowerCase().includes(term);
      return matchesCat && matchesSearch;
    });
  }, [items, query, activeFilter]);

  const setFilter = useCallback((id: string) => setActiveFilter(id), []);

  return { query, setQuery, activeFilter, setFilter, filtered, tabs };
}
