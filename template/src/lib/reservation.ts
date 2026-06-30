export function gmailComposeUrl(email: string, subject?: string): string {
  const to = email.trim();
  if (!to) return "";
  const params = new URLSearchParams({ view: "cm", fs: "1", to });
  if (subject) params.set("su", subject);
  return `https://mail.google.com/mail/?${params.toString()}`;
}

export function telHref(phone?: string): string {
  if (!phone) return "";
  return `tel:${phone.replace(/\s/g, "")}`;
}

export function mapsHref(
  address: string,
  coords?: { latitude: number; longitude: number }
): string {
  if (coords?.latitude && coords?.longitude) {
    return `https://www.google.com/maps?q=${coords.latitude},${coords.longitude}`;
  }
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(address)}`;
}

export function mapsEmbedSrc(coords: { latitude: number; longitude: number }): string {
  return `https://maps.google.com/maps?q=${coords.latitude},${coords.longitude}&z=15&output=embed`;
}
