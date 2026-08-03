export function getCallInNumber(): string {
  return (process.env.NEXT_PUBLIC_CALL_IN_NUMBER || "").trim();
}

export function getCallInDisplay(): string {
  const n = getCallInNumber();
  if (!n) return "";
  // Soft format for AU-looking numbers; otherwise return as-is.
  const digits = n.replace(/\D/g, "");
  if (digits.startsWith("61") && digits.length >= 11) {
    return `+61 ${digits.slice(2, 3)} ${digits.slice(3, 7)} ${digits.slice(7)}`;
  }
  return n;
}

export function getCallInTelHref(): string | null {
  const n = getCallInNumber();
  if (!n) return null;
  const digits = n.replace(/[^\d+]/g, "");
  return digits ? `tel:${digits}` : null;
}
