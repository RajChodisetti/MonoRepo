export type ConsultationSlot = {
  date: string;
  time: string;
  iso: string;
  available: boolean;
};

export type AvailabilityResponse = {
  status: string;
  slots: ConsultationSlot[];
};

export type BookSuccessResponse = {
  status: "success";
  confirmation_code: string;
  prospect_name: string;
  prospect_email: string;
  prospect_phone?: string;
  slot: string;
  booking_date: string;
  booking_time: string;
  calendar_link: string;
  message: string;
};

export type BookConflictResponse = {
  status: "conflict";
  message: string;
  alternatives: string[];
};

export async function fetchAvailability(days = 7): Promise<AvailabilityResponse> {
  const res = await fetch(`/api/consultations/availability?days=${days}`, {
    cache: "no-store",
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.message || "Could not load available slots.");
  }
  return data;
}

export async function bookConsultation(payload: {
  date: string;
  time: string;
  slot?: string;
  prospect_name: string;
  prospect_email: string;
  prospect_phone: string;
  source?: string;
}): Promise<BookSuccessResponse> {
  const res = await fetch("/api/consultations", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...payload, source: payload.source ?? "web" }),
  });
  const data = await res.json();
  if (res.status === 409) {
    const conflict = data as BookConflictResponse;
    throw new Error(
      conflict.alternatives?.length
        ? `${conflict.message} Try: ${conflict.alternatives.slice(0, 3).join(", ")}`
        : conflict.message || "That slot was just taken.",
    );
  }
  if (!res.ok) {
    throw new Error(data.message || "Booking failed. Please try again.");
  }
  return data as BookSuccessResponse;
}
