export function getBookCallUrl(): string {
  return "/book";
}

export function getContactEmail(): string {
  return process.env.NEXT_PUBLIC_CONTACT_EMAIL ?? "contact@tuvisolutions.com";
}

export function getLinkedInUrl(): string {
  return process.env.NEXT_PUBLIC_LINKEDIN_URL ?? "https://linkedin.com/company/tuvi-solutions";
}
