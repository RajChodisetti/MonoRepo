"use client";

import { FormEvent, useState } from "react";
import { EMPTY_PROPERTY_FORM, type PropertyFormValues } from "@/lib/types";

export function PropertyForm({
  initial,
  submitLabel,
  onSubmit,
  onCancel,
}: {
  initial?: PropertyFormValues;
  submitLabel: string;
  onSubmit: (values: PropertyFormValues) => Promise<void>;
  onCancel: () => void;
}) {
  const [form, setForm] = useState<PropertyFormValues>(initial || EMPTY_PROPERTY_FORM);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  function update<K extends keyof PropertyFormValues>(key: K, value: PropertyFormValues[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      if (!form.title.trim() || !form.city.trim()) {
        throw new Error("Title and city are required");
      }
      await onSubmit(form);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && <p className="text-sm text-[var(--bad)]">{error}</p>}
      <div className="grid gap-4 md:grid-cols-2">
        <div className="field md:col-span-2">
          <label>Title</label>
          <input value={form.title} onChange={(e) => update("title", e.target.value)} required />
        </div>
        <div className="field">
          <label>Type</label>
          <select value={form.type} onChange={(e) => update("type", e.target.value)}>
            <option value="apartment">Apartment</option>
            <option value="villa">Villa</option>
            <option value="plot">Plot</option>
          </select>
        </div>
        <div className="field">
          <label>Status</label>
          <select value={form.status} onChange={(e) => update("status", e.target.value)}>
            <option value="sale">Sale</option>
            <option value="rent">Rent</option>
          </select>
        </div>
        <div className="field">
          <label>City</label>
          <input value={form.city} onChange={(e) => update("city", e.target.value)} required />
        </div>
        <div className="field">
          <label>Locality</label>
          <input value={form.locality} onChange={(e) => update("locality", e.target.value)} />
        </div>
        <div className="field">
          <label>State</label>
          <input value={form.state} onChange={(e) => update("state", e.target.value)} />
        </div>
        <div className="field">
          <label>Pincode</label>
          <input value={form.pincode} onChange={(e) => update("pincode", e.target.value)} />
        </div>
        <div className="field">
          <label>BHK</label>
          <input
            inputMode="numeric"
            value={form.bhk}
            onChange={(e) => update("bhk", e.target.value)}
          />
        </div>
        <div className="field">
          <label>Size (sq ft)</label>
          <input
            inputMode="numeric"
            value={form.size_sqft}
            onChange={(e) => update("size_sqft", e.target.value)}
          />
        </div>
        <div className="field">
          <label>Size (sq yd)</label>
          <input
            inputMode="numeric"
            value={form.size_sqyd}
            onChange={(e) => update("size_sqyd", e.target.value)}
          />
        </div>
        <div className="field">
          <label>Price (INR)</label>
          <input
            inputMode="numeric"
            value={form.price_inr}
            onChange={(e) => update("price_inr", e.target.value)}
          />
        </div>
        <div className="field">
          <label>Furnishing</label>
          <input value={form.furnishing} onChange={(e) => update("furnishing", e.target.value)} />
        </div>
        <div className="field">
          <label>Possession</label>
          <input value={form.possession} onChange={(e) => update("possession", e.target.value)} />
        </div>
        <div className="field md:col-span-2">
          <label>Amenities (comma separated)</label>
          <input value={form.amenities} onChange={(e) => update("amenities", e.target.value)} />
        </div>
        <div className="field md:col-span-2">
          <label>Nearby (comma separated)</label>
          <input value={form.nearby} onChange={(e) => update("nearby", e.target.value)} />
        </div>
        <div className="field md:col-span-2">
          <label>Description</label>
          <textarea
            rows={3}
            value={form.description}
            onChange={(e) => update("description", e.target.value)}
          />
        </div>
        <label className="flex items-center gap-2 text-sm text-muted md:col-span-2">
          <input
            type="checkbox"
            checked={form.negotiable}
            onChange={(e) => update("negotiable", e.target.checked)}
          />
          Negotiable
        </label>
      </div>
      <div className="flex gap-3 justify-end pt-2">
        <button type="button" className="btn btn-ghost" onClick={onCancel} disabled={loading}>
          Cancel
        </button>
        <button type="submit" className="btn btn-primary" disabled={loading}>
          {loading ? "Saving…" : submitLabel}
        </button>
      </div>
    </form>
  );
}
