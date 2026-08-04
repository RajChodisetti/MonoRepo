"use client";

import { useCallback, useEffect, useState } from "react";
import { PropertyForm } from "@/components/PropertyForm";
import { apiFetch } from "@/lib/client-api";
import {
  formToPayload,
  recordToForm,
  type PropertyFormValues,
  type PropertyRecord,
  type PropertySummary,
} from "@/lib/types";

type Mode = { kind: "closed" } | { kind: "create" } | { kind: "edit"; id: string };

export function PropertiesClient() {
  const [items, setItems] = useState<PropertySummary[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [mode, setMode] = useState<Mode>({ kind: "closed" });
  const [editForm, setEditForm] = useState<PropertyFormValues | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await apiFetch<{ results: PropertySummary[] }>("/api/properties");
      setItems(data.results || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load properties");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function openEdit(id: string) {
    setError("");
    try {
      const data = await apiFetch<{ property: PropertyRecord }>(`/api/properties/${id}`);
      setEditForm(recordToForm(data.property));
      setMode({ kind: "edit", id });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load property");
    }
  }

  async function createProperty(values: PropertyFormValues) {
    await apiFetch("/api/properties", {
      method: "POST",
      body: formToPayload(values),
    });
    setMode({ kind: "closed" });
    await load();
  }

  async function updateProperty(values: PropertyFormValues) {
    if (mode.kind !== "edit") return;
    await apiFetch(`/api/properties/${mode.id}`, {
      method: "PUT",
      body: formToPayload(values),
    });
    setMode({ kind: "closed" });
    setEditForm(null);
    await load();
  }

  async function removeProperty(id: string, title: string) {
    if (!window.confirm(`Delete “${title}”? This cannot be undone.`)) return;
    try {
      await apiFetch(`/api/properties/${id}`, { method: "DELETE" });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Delete failed");
    }
  }

  return (
    <div className="space-y-5">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1
            className="text-3xl tracking-tight"
            style={{ fontFamily: "var(--font-fraunces), Georgia, serif" }}
          >
            Properties
          </h1>
          <p className="text-sm text-muted mt-1">
            Manage the voice agent mock inventory (`properties.json`).
          </p>
        </div>
        <button type="button" className="btn btn-primary" onClick={() => setMode({ kind: "create" })}>
          Add listing
        </button>
      </header>

      {error && <p className="text-sm text-[var(--bad)]">{error}</p>}

      {mode.kind !== "closed" && (
        <section className="rounded-2xl border border-line bg-panel p-5">
          <h2 className="text-lg font-semibold mb-4">
            {mode.kind === "create" ? "New listing" : "Edit listing"}
          </h2>
          <PropertyForm
            initial={mode.kind === "edit" ? editForm || undefined : undefined}
            submitLabel={mode.kind === "create" ? "Create" : "Save changes"}
            onCancel={() => {
              setMode({ kind: "closed" });
              setEditForm(null);
            }}
            onSubmit={mode.kind === "create" ? createProperty : updateProperty}
          />
        </section>
      )}

      <section className="rounded-2xl border border-line bg-panel overflow-x-auto">
        {loading ? (
          <p className="p-5 text-sm text-muted">Loading listings…</p>
        ) : items.length === 0 ? (
          <p className="p-5 text-sm text-muted">No listings yet. Add the first one.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Title</th>
                <th>Type</th>
                <th>Location</th>
                <th>Status</th>
                <th>Price</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id}>
                  <td>
                    <div className="font-semibold">{item.title}</div>
                    <div className="text-xs text-muted">{item.id}</div>
                  </td>
                  <td className="capitalize">{item.type}</td>
                  <td>
                    {[item.locality, item.city].filter(Boolean).join(", ")}
                  </td>
                  <td>
                    <span className="chip capitalize">{item.status}</span>
                  </td>
                  <td>{item.price}</td>
                  <td>
                    <div className="flex gap-2 justify-end">
                      <button type="button" className="btn btn-ghost" onClick={() => openEdit(item.id)}>
                        Edit
                      </button>
                      <button
                        type="button"
                        className="btn btn-danger"
                        onClick={() => removeProperty(item.id, item.title)}
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}
