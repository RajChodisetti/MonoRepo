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
  calendar_link?: string;
  message: string;
};

type BookConflictResponse = {
  status: "conflict";
  message: string;
  alternatives?: string[];
};

type ErrorResponse = {
  message?: string;
};

async function readJson<T>(response: Response): Promise<T> {
  return (await response.json()) as T;
}

export async function fetchAvailability(days = 14): Promise<AvailabilityResponse> {
  const response = await fetch(`/api/consultations/availability?days=${days}`, {
    cache: "no-store",
  });
  const data = await readJson<AvailabilityResponse & ErrorResponse>(response);
  if (!response.ok) {
    throw new Error(data.message || "Could not load available times.");
  }
  return data;
}

export async function bookConsultation(payload: {
  date: string;
  time: string;
  prospect_name: string;
  prospect_email: string;
  prospect_phone: string;
  source?: string;
}): Promise<BookSuccessResponse> {
  const response = await fetch("/api/consultations", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...payload, source: payload.source ?? "web" }),
  });
  const data = await readJson<BookSuccessResponse | BookConflictResponse | ErrorResponse>(response);

  if (response.status === 409) {
    const conflict = data as BookConflictResponse;
    throw new Error(conflict.message || "That time was just booked. Choose another time.");
  }
  if (!response.ok) {
    throw new Error((data as ErrorResponse).message || "Booking failed. Please try again.");
  }
  return data as BookSuccessResponse;
}
