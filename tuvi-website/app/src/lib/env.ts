export function getBookCallUrl(): string {
  return "/book";
}

export function getContactEmail(): string {
  return process.env.NEXT_PUBLIC_CONTACT_EMAIL ?? "contact@tuvisolutions.com";
}

export function getLinkedInUrl(): string {
  return process.env.NEXT_PUBLIC_LINKEDIN_URL ?? "https://linkedin.com/company/tuvi-solutions";
}

/** Public Twilio number for display (never expose Twilio auth). */
export function getCallInNumber(): string {
  return (process.env.NEXT_PUBLIC_CALL_IN_NUMBER || "").trim();
}

export function getCallInDisplay(): string {
  const n = getCallInNumber();
  if (!n) return "";
  const digits = n.replace(/\D/g, "");
  // AU mobile: +61 4XX XXX XXX
  if (digits.startsWith("614") && digits.length === 11) {
    const local = digits.slice(2);
    return `+61 ${local.slice(0, 3)} ${local.slice(3, 6)} ${local.slice(6)}`;
  }
  return n.startsWith("+") ? n : `+${digits}`;
}

export function getCallInTelHref(): string | null {
  const n = getCallInNumber();
  if (!n) return null;
  const e164 = n.startsWith("+") ? n.replace(/\s/g, "") : `+${n.replace(/\D/g, "")}`;
  return `tel:${e164}`;
}
