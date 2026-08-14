"use client";

import { useEffect, useId, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import type { Restaurant } from "@/lib/types";

type Props = {
  label: string;
  selected: Restaurant | null;
  onSelect: (restaurant: Restaurant | null) => void;
  help?: string;
};

export function RestaurantSearch({ label, selected, onSelect, help }: Props) {
  const inputId = useId();
  const listId = `${inputId}-results`;
  const [query, setQuery] = useState(selected?.name || "");
  const [results, setResults] = useState<Restaurant[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    setQuery(selected?.name || "");
    setResults([]);
    setOpen(false);
  }, [selected?.id, selected?.name]);

  useEffect(() => {
    const term = query.trim();
    if (selected || term.length < 2) {
      setLoading(false);
      setResults([]);
      setError(null);
      return;
    }

    let cancelled = false;
    const timer = window.setTimeout(async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await adminFetch<{ items: Restaurant[] }>("restaurants", {
          query: { restaurant: term },
        });
        if (!cancelled) {
          setResults((response.items || []).slice(0, 8));
          setOpen(true);
        }
      } catch (reason) {
        if (!cancelled) {
          setResults([]);
          setError(reason instanceof Error ? reason.message : "Restaurant search failed");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }, 250);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [query, selected]);

  return (
    <div className="restaurant-search">
      <label className="field-label" htmlFor={inputId}>
        {label}
        <div className="restaurant-search-input-row">
          <input
            id={inputId}
            className="input"
            role="combobox"
            aria-autocomplete="list"
            aria-controls={listId}
            aria-expanded={open && !selected}
            value={query}
            onChange={(event) => {
              if (selected) onSelect(null);
              setQuery(event.target.value);
              setOpen(true);
            }}
            onFocus={() => setOpen(true)}
            placeholder="Search by restaurant name…"
            autoComplete="off"
          />
          {selected ? (
            <button
              className="btn btn-secondary btn-compact"
              type="button"
              onClick={() => onSelect(null)}
            >
              Clear
            </button>
          ) : null}
        </div>
      </label>
      {help ? <span className="field-help">{help}</span> : null}
      {loading ? <span className="field-help" role="status">Searching…</span> : null}
      {error ? <span className="restaurant-search-error" role="alert">{error}</span> : null}
      {open && !selected && results.length > 0 ? (
        <div className="restaurant-search-results" id={listId} role="listbox">
          {results.map((restaurant) => (
            <button
              type="button"
              role="option"
              aria-selected="false"
              key={restaurant.id}
              onClick={() => onSelect(restaurant)}
            >
              <strong>{restaurant.name}</strong>
              <span>{restaurant.email || restaurant.status}</span>
            </button>
          ))}
        </div>
      ) : null}
      {selected ? (
        <div className="restaurant-search-selected" role="status">
          Using authoritative listing facts for <strong>{selected.name}</strong>.
        </div>
      ) : null}
    </div>
  );
}
