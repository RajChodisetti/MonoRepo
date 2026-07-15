"use client";

import { useMenuFilter } from "../hooks/useMenuFilter";
import type { ElysianContent } from "../lib/mapContent";
import ElysianImage from "./ElysianImage";

export default function ElysianMenu({
  menuItems,
  menuTabs,
}: {
  menuItems: ElysianContent["menuItems"];
  menuTabs: ElysianContent["menuTabs"];
}) {
  const { query, setQuery, activeFilter, setFilter, filtered, tabs } = useMenuFilter(
    menuItems,
    menuTabs,
  );

  if (!menuItems.length) return null;

  return (
    <section className="menu section" id="menu">
      <div className="container">
        <div className="section-head reveal fade-up">
          <p className="eyebrow">The Menu</p>
          <h2 className="section-title">
            A Journey Through <span className="gold-text">Flavor</span>
          </h2>
          <p className="section-sub">Search and filter our full menu, curated across land, sea and season.</p>
        </div>

        <div className="menu-controls reveal fade-up">
          <div className="menu-search">
            <svg viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="1.6" />
              <path d="m20 20-3.5-3.5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
            </svg>
            <input
              type="text"
              id="menuSearch"
              placeholder="Search dishes..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>
          <div className="menu-tabs" id="menuTabs">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                type="button"
                className={`menu-tab${activeFilter === tab.id ? " active" : ""}`}
                data-filter={tab.id}
                onClick={() => setFilter(tab.id)}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        <div className="menu-grid" id="menuGrid">
          {filtered.map((item, i) => (
            <div key={`${item.name}-${i}`} className="menu-item" style={{ animationDelay: `${i * 0.05}s` }}>
              {item.image ? (
                <ElysianImage
                  src={item.image}
                  alt={item.name}
                  width={92}
                  height={92}
                  sizes="92px"
                />
              ) : null}
              <div className="menu-item-info">
                <div className="menu-item-top">
                  <h4>{item.name}</h4>
                  {item.price ? <span className="menu-item-price">{item.price}</span> : null}
                </div>
                {item.description ? <p>{item.description}</p> : null}
                <span className="menu-item-cat">{item.categories.filter((c) => c !== "all").join(" · ")}</span>
              </div>
            </div>
          ))}
        </div>
        <p className="menu-empty" id="menuEmpty" hidden={filtered.length !== 0}>
          No dishes match your search — try another word.
        </p>
      </div>
    </section>
  );
}
