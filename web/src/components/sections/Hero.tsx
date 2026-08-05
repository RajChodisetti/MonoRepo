"use client";

import { FormEvent, KeyboardEvent, useEffect, useId, useRef, useState } from "react";
import { useRouter } from "next/navigation";

const rotatingLines = ["grow first-party sales.", "grow online discovery."] as const;

type SearchHit = {
  placeId: string;
  name: string;
  address: string;
  latitude?: number;
  longitude?: number;
  rating?: number;
  source: "monorepo" | "places";
};

function StarIcon() {
  return (
    <svg viewBox="0 0 16 16" className="h-4 w-4 fill-current" aria-hidden="true">
      <path d="M8 1.2 9.7 5.8 14.6 6.1 10.8 9.3 12 14.2 8 11.7 4 14.2 5.2 9.3 1.4 6.1 6.3 5.8 8 1.2Z" />
    </svg>
  );
}

function AnimatedPhrase({
  text,
  show,
  mode,
}: {
  text: string;
  show: boolean;
  mode: "enter" | "exit";
}) {
  const words = text.split(" ");

  return (
    <span className="mx-auto flex w-full flex-wrap items-center justify-center gap-x-[0.28em]">
      {words.map((word, i) => {
        const delay = show
          ? mode === "enter"
            ? i * 140
            : (words.length - 1 - i) * 90
          : 0;

        const hidden =
          mode === "enter"
            ? "translate-y-[110%] opacity-0"
            : "-translate-y-[110%] opacity-0";

        return (
          <span key={`${text}-${word}-${i}`} className="inline-block overflow-hidden pb-[0.08em]">
            <span
              className={`inline-block transition-all ease-[cubic-bezier(0.22,1,0.36,1)] ${
                show ? "translate-y-0 opacity-100" : hidden
              }`}
              style={{
                transitionDelay: `${delay}ms`,
                transitionDuration: "420ms",
              }}
            >
              {word}
            </span>
          </span>
        );
      })}
    </span>
  );
}

export default function Hero() {
  const router = useRouter();
  const listId = useId();
  const wrapRef = useRef<HTMLDivElement>(null);
  const [index, setIndex] = useState(0);
  const [show, setShow] = useState(true);
  const [mode, setMode] = useState<"enter" | "exit">("enter");

  const [query, setQuery] = useState("");
  const [location, setLocation] = useState("Australia");
  const [results, setResults] = useState<SearchHit[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);
  const [highlight, setHighlight] = useState(-1);
  const [selected, setSelected] = useState<SearchHit | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const HOLD_MS = 2600;
    const EXIT_MS = 700;
    const ENTER_GAP_MS = 80;

    let exitTimer = 0;
    let enterTimer = 0;

    const loop = window.setInterval(() => {
      setMode("exit");
      setShow(false);

      exitTimer = window.setTimeout(() => {
        setIndex((value) => (value + 1) % rotatingLines.length);
        setMode("enter");
        enterTimer = window.setTimeout(() => {
          setShow(true);
        }, ENTER_GAP_MS);
      }, EXIT_MS);
    }, HOLD_MS + EXIT_MS + 900);

    return () => {
      window.clearInterval(loop);
      window.clearTimeout(exitTimer);
      window.clearTimeout(enterTimer);
    };
  }, []);

  useEffect(() => {
    const q = query.trim();
    if (selected && selected.name === query) {
      setResults([]);
      setOpen(false);
      setLoading(false);
      return;
    }
    if (q.length < 2) {
      setResults([]);
      setOpen(false);
      setLoading(false);
      setSearchError(null);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setSearchError(null);
    const timer = window.setTimeout(async () => {
      try {
        const params = new URLSearchParams({
          q,
          location: location.trim() || "Australia",
        });
        const res = await fetch(`/api/restaurants/search?${params.toString()}`);
        const json = await res.json();
        if (cancelled) return;
        if (!res.ok) throw new Error(json.error || "Search failed");
        const hits = (json.results || []) as SearchHit[];
        setResults(hits);
        setOpen(true);
        setHighlight(hits.length ? 0 : -1);
      } catch (e) {
        if (!cancelled) {
          setResults([]);
          setSearchError(e instanceof Error ? e.message : "Search failed");
          setOpen(true);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }, 280);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [query, location, selected]);

  useEffect(() => {
    function onDocClick(event: MouseEvent) {
      if (!wrapRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  function goToReport(hit: SearchHit) {
    setSubmitting(true);
    setSelected(hit);
    setQuery(hit.name);
    setOpen(false);
    const params = new URLSearchParams();
    params.set("name", hit.name);
    if (hit.address) params.set("address", hit.address);
    if (typeof hit.latitude === "number" && typeof hit.longitude === "number") {
      params.set("lat", String(hit.latitude));
      params.set("lng", String(hit.longitude));
    }
    router.push(`/report/${encodeURIComponent(hit.placeId)}?${params.toString()}`);
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    if (submitting) return;

    if (selected && selected.name === query.trim()) {
      goToReport(selected);
      return;
    }

    const fromHighlight = highlight >= 0 ? results[highlight] : undefined;
    if (fromHighlight) {
      goToReport(fromHighlight);
      return;
    }
    if (results[0]) {
      goToReport(results[0]);
      return;
    }

    const q = query.trim();
    if (q.length < 2) return;
    setSubmitting(true);
    try {
      const params = new URLSearchParams({
        q,
        location: location.trim() || "Australia",
      });
      const res = await fetch(`/api/restaurants/search?${params.toString()}`);
      const json = await res.json();
      const hits = (json.results || []) as SearchHit[];
      if (hits[0]) {
        goToReport(hits[0]);
        return;
      }
      setSearchError("No restaurants found. Try a fuller name, city, or postcode.");
      setOpen(true);
    } catch {
      setSearchError("Search failed. Try again.");
      setOpen(true);
    } finally {
      setSubmitting(false);
    }
  }

  function onKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (!open && (event.key === "ArrowDown" || event.key === "ArrowUp")) {
      if (results.length) setOpen(true);
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setHighlight((h) => (results.length ? (h + 1) % results.length : -1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setHighlight((h) => (results.length ? (h <= 0 ? results.length - 1 : h - 1) : -1));
    } else if (event.key === "Escape") {
      setOpen(false);
    }
  }

  return (
    <section className="hero-atmosphere relative px-6 pb-20 pt-16 md:px-10 md:pb-28 md:pt-24 lg:px-12">
      <div
        className="pointer-events-none absolute inset-0 hero-grid opacity-30 [mask-image:radial-gradient(48rem_30rem_at_50%_35%,black,transparent)]"
        aria-hidden="true"
      />

      <div className="relative z-10 mx-auto flex w-full max-w-[1120px] flex-col items-center text-center">
        <p className="inline-flex flex-wrap items-center justify-center gap-1.5 text-[16px] font-medium text-muted">
          <span className="text-ink">4.8</span>
          <span className="inline-flex items-center gap-0.5 text-ink" aria-label="5 stars">
            {Array.from({ length: 5 }).map((_, i) => (
              <StarIcon key={i} />
            ))}
          </span>
          <span>across 1,000+ reviews</span>
        </p>

        <h1 className="mt-7 w-full font-display text-[clamp(1.55rem,calc(4.8vw+0.65rem),5rem)] font-semibold leading-[1.12] tracking-[-0.04em]">
          <span className="block whitespace-nowrap text-center text-muted">
            The AI platform restaurants use to
          </span>

          <span className="relative mx-auto mt-1 grid min-h-[1.25em] w-full place-items-center overflow-hidden text-center text-ink">
            <AnimatedPhrase key={rotatingLines[index]} text={rotatingLines[index]} show={show} mode={mode} />
          </span>
        </h1>

        <div ref={wrapRef} className="relative mt-10 w-full max-w-[560px]">
          <form className="w-full space-y-2 text-left" onSubmit={onSubmit} autoComplete="off">
            <div className="flex w-full items-stretch gap-0 rounded-2xl border border-border bg-bg p-1.5 shadow-[0_10px_40px_rgba(15,39,31,0.1)]">
              <label
                htmlFor="restaurant-location"
                className="flex w-[38%] min-w-[7.5rem] max-w-[11.5rem] shrink-0 cursor-text flex-col justify-center gap-0.5 border-r border-border px-3.5 py-2"
              >
                <span className="block text-left text-[10px] font-semibold uppercase tracking-[0.08em] text-muted">
                  Location
                </span>
                <input
                  id="restaurant-location"
                  name="location"
                  type="text"
                  value={location}
                  onChange={(e) => {
                    setSelected(null);
                    setLocation(e.target.value);
                  }}
                  placeholder="City or postcode"
                  className="w-full min-w-0 bg-transparent text-left text-[15px] leading-tight text-ink outline-none placeholder:text-secondary"
                />
              </label>
              <label
                htmlFor="restaurant-search"
                className="flex min-w-0 flex-1 cursor-text flex-col justify-center gap-0.5 px-3.5 py-2"
              >
                <span className="sr-only">Restaurant name</span>
                <input
                  id="restaurant-search"
                  name="restaurant"
                  type="text"
                  role="combobox"
                  aria-expanded={open}
                  aria-controls={listId}
                  aria-autocomplete="list"
                  aria-activedescendant={highlight >= 0 ? `${listId}-opt-${highlight}` : undefined}
                  value={query}
                  onChange={(e) => {
                    setSelected(null);
                    setQuery(e.target.value);
                  }}
                  onFocus={() => {
                    if (results.length || searchError) setOpen(true);
                  }}
                  onKeyDown={onKeyDown}
                  placeholder="Restaurant name"
                  className="w-full min-w-0 bg-transparent text-left text-[17px] leading-tight text-ink outline-none placeholder:text-secondary"
                />
              </label>
              <button
                type="submit"
                disabled={submitting}
                className="m-0.5 inline-flex shrink-0 cursor-pointer items-center gap-2 self-center rounded-xl bg-primary px-4 py-3 text-[15px] font-semibold text-bg transition-colors hover:bg-primary-dim disabled:opacity-60 sm:px-5"
              >
                {submitting ? "Opening…" : "Get report"}
                <span aria-hidden="true">↑</span>
              </button>
            </div>
            <p className="text-center text-[12px] text-muted">
              Default location is Australia — enter a city or postcode to narrow results.
            </p>
          </form>

          {open ? (
            <div
              id={listId}
              role="listbox"
              className="absolute left-0 right-0 top-[calc(100%+8px)] z-20 overflow-hidden rounded-2xl border border-border bg-bg text-left shadow-[0_16px_48px_rgba(15,39,31,0.14)]"
            >
              {loading ? (
                <p className="px-4 py-3 text-[14px] text-muted">Searching restaurants…</p>
              ) : searchError ? (
                <p className="px-4 py-3 text-[14px] text-[#b42318]">{searchError}</p>
              ) : results.length === 0 ? (
                <p className="px-4 py-3 text-[14px] text-muted">No matches yet — keep typing.</p>
              ) : (
                <ul className="max-h-[280px] overflow-auto py-1">
                  {results.map((hit, i) => (
                    <li key={`${hit.source}-${hit.placeId}`}>
                      <button
                        type="button"
                        id={`${listId}-opt-${i}`}
                        role="option"
                        aria-selected={i === highlight}
                        className={`flex w-full cursor-pointer flex-col gap-0.5 px-4 py-2.5 text-left transition-colors ${
                          i === highlight ? "bg-[#eef4f0]" : "hover:bg-[#f5f7f6]"
                        }`}
                        onMouseEnter={() => setHighlight(i)}
                        onClick={() => goToReport(hit)}
                      >
                        <span className="truncate text-[15px] font-semibold text-ink">{hit.name}</span>
                        <span className="truncate text-[12.5px] text-muted">
                          {hit.address || (hit.source === "monorepo" ? "From your inventory" : "Google Places")}
                          {hit.rating != null ? ` · ${hit.rating.toFixed(1)}★` : ""}
                        </span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          ) : null}
        </div>
      </div>
    </section>
  );
}
