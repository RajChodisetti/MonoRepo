export type PropertySummary = {
  id: string;
  title: string;
  type?: string;
  status?: string;
  city?: string;
  locality?: string;
  size?: string | null;
  price?: string;
  price_inr?: number | null;
  features?: string[];
  nearby?: string[];
  possession?: string;
  furnishing?: string;
};

export type PropertyFormValues = {
  id?: string;
  title: string;
  type: string;
  status: string;
  city: string;
  locality: string;
  state: string;
  pincode: string;
  bhk: string;
  size_sqft: string;
  size_sqyd: string;
  price_inr: string;
  description: string;
  amenities: string;
  nearby: string;
  furnishing: string;
  possession: string;
  negotiable: boolean;
};

export type PropertyRecord = {
  id: string;
  title?: string;
  type?: string;
  status?: string;
  description?: string;
  location?: {
    city?: string;
    locality?: string;
    state?: string;
    pincode?: string;
  };
  specs?: {
    bhk?: number | null;
    size_sqft?: number | null;
    size_sqyd?: number | null;
    furnishing?: string;
    possession?: string;
  };
  pricing?: {
    price_inr?: number | null;
    negotiable?: boolean;
  };
  amenities?: string[];
  nearby?: string[];
  city?: string;
  locality?: string;
  bhk?: number | null;
  size_sqft?: number | null;
  size_sqyd?: number | null;
  price_inr?: number | null;
};

export const EMPTY_PROPERTY_FORM: PropertyFormValues = {
  title: "",
  type: "apartment",
  status: "sale",
  city: "Hyderabad",
  locality: "",
  state: "Telangana",
  pincode: "",
  bhk: "",
  size_sqft: "",
  size_sqyd: "",
  price_inr: "",
  description: "",
  amenities: "",
  nearby: "",
  furnishing: "",
  possession: "ready_to_move",
  negotiable: true,
};

export function recordToForm(p: PropertyRecord): PropertyFormValues {
  return {
    id: p.id,
    title: p.title || "",
    type: p.type || "apartment",
    status: p.status || "sale",
    city: p.location?.city || p.city || "",
    locality: p.location?.locality || p.locality || "",
    state: p.location?.state || "",
    pincode: p.location?.pincode || "",
    bhk: p.specs?.bhk != null ? String(p.specs.bhk) : p.bhk != null ? String(p.bhk) : "",
    size_sqft:
      p.specs?.size_sqft != null
        ? String(p.specs.size_sqft)
        : p.size_sqft != null
          ? String(p.size_sqft)
          : "",
    size_sqyd:
      p.specs?.size_sqyd != null
        ? String(p.specs.size_sqyd)
        : p.size_sqyd != null
          ? String(p.size_sqyd)
          : "",
    price_inr:
      p.pricing?.price_inr != null
        ? String(p.pricing.price_inr)
        : p.price_inr != null
          ? String(p.price_inr)
          : "",
    description: p.description || "",
    amenities: (p.amenities || []).join(", "),
    nearby: (p.nearby || []).join(", "),
    furnishing: p.specs?.furnishing || "",
    possession: p.specs?.possession || "",
    negotiable: p.pricing?.negotiable ?? true,
  };
}

export function formToPayload(form: PropertyFormValues) {
  return {
    id: form.id?.trim() || undefined,
    title: form.title.trim(),
    type: form.type,
    status: form.status,
    city: form.city.trim(),
    locality: form.locality.trim(),
    state: form.state.trim(),
    pincode: form.pincode.trim(),
    bhk: form.bhk ? Number(form.bhk) : null,
    size_sqft: form.size_sqft ? Number(form.size_sqft) : null,
    size_sqyd: form.size_sqyd ? Number(form.size_sqyd) : null,
    price_inr: form.price_inr ? Number(form.price_inr) : null,
    description: form.description.trim(),
    amenities: form.amenities,
    nearby: form.nearby,
    furnishing: form.furnishing.trim(),
    possession: form.possession.trim(),
    negotiable: form.negotiable,
  };
}
