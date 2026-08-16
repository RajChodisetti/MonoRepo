import type { OutreachEmailSignature, OutreachSequenceStep } from "@/lib/types";

export const WEBSITE_TOKEN = "{{website_url}}";
export const DEFAULT_EMAIL_SIGNATURE: OutreachEmailSignature = {
  name: "Praveen Maurya",
  title: "Business Development Manager",
  additional_details: "",
};

export const PLACEHOLDERS = [
  { token: "[GREETING]", description: "Complete verified-fact greeting; use it once in the first enabled message body only." },
  { token: "[FIRST_NAME]", description: "Owner first name; falls back to the restaurant name plus “team”." },
  { token: "[RESTAURANT_NAME]", description: "Saved restaurant name." },
  { token: "[CUISINE]", description: "First verified Google cuisine, or N/A when unavailable." },
  { token: "[CITY]", description: "Verified Google city, or N/A when unavailable." },
  { token: "[RATING]", description: "Verified Google rating to one decimal place, or N/A." },
  { token: "[TOTAL_REVIEWS]", description: "Verified Google review count, or N/A." },
  { token: "[WEBSITE_URL]", description: "Tuvi Solutions website URL." },
] as const;

export const LEGACY_PLACEHOLDERS = [
  { token: "{{greeting}}", description: "Simple Hi greeting retained for existing follow-ups." },
  { token: "{{greeting01}}", description: "Legacy alias for [GREETING]." },
  { token: "{{restaurant_name}}", description: "Legacy alias for [RESTAURANT_NAME]." },
  { token: "{{website_url}}", description: "Legacy alias for [WEBSITE_URL]." },
] as const;

export const ALLOWED_MERGE_TAGS = new Set([
  "greeting",
  "greeting01",
  "restaurant_name",
  "website_url",
]);
export const ALLOWED_BRACKET_PLACEHOLDERS = new Set<string>(PLACEHOLDERS.map(({ token }) => token));

const MERGE_TAG_PATTERN = /{{([^{}\r\n]+)}}/g;
const BRACKET_PLACEHOLDER_PATTERN = /\[[A-Za-z][A-Za-z0-9_ ]*\]/g;
const HTML_TAG_PATTERN = /<\s*\/?\s*[a-z][^>]*>/i;

export function validateSequenceStep(
  step: OutreachSequenceStep,
  index: number,
): string[] {
  const issues: string[] = [];
  const subject = step.subject_template.trim();
  const body = step.body_text_template.trim();

  if (!subject) issues.push("Add a subject.");
  if (subject.length > 200) issues.push("Keep the subject at 200 characters or fewer.");
  if (!body) issues.push("Add a plain-text message.");
  if (HTML_TAG_PATTERN.test(subject) || HTML_TAG_PATTERN.test(body)) {
    issues.push("HTML is not allowed; use plain text only.");
  }

  const mergeTags = Array.from(`${subject}\n${body}`.matchAll(MERGE_TAG_PATTERN));
  const unknownTags = mergeTags
    .map((match) => match[1])
    .filter((tag) => !ALLOWED_MERGE_TAGS.has(tag));
  if (unknownTags.length > 0) {
    issues.push(`Unknown merge tag: {{${Array.from(new Set(unknownTags)).join("}}, {{")}}}.`);
  }

  const bracketTags = Array.from(`${subject}\n${body}`.matchAll(BRACKET_PLACEHOLDER_PATTERN));
  const unknownBracketTags = bracketTags
    .map((match) => match[0])
    .filter((tag) => !ALLOWED_BRACKET_PLACEHOLDERS.has(tag));
  if (unknownBracketTags.length > 0) {
    issues.push(`Unknown placeholder: ${Array.from(new Set(unknownBracketTags)).join(", ")}.`);
  }

  const greeting01SubjectCount = subject.split("{{greeting01}}").length - 1 + subject.split("[GREETING]").length - 1;
  const greeting01BodyCount = body.split("{{greeting01}}").length - 1 + body.split("[GREETING]").length - 1;
  if (greeting01SubjectCount > 0) {
    issues.push("Use [GREETING] only in the message body.");
  }
  if (greeting01BodyCount > 0 && index !== 0) {
    issues.push("Use [GREETING] only in the first enabled email.");
  }
  if (greeting01BodyCount > 1) {
    issues.push("Use the complete greeting exactly once.");
  }
  if (greeting01BodyCount === 1 && body.includes("{{greeting}}")) {
    issues.push("Replace {{greeting}} when using the complete greeting.");
  }

  if (!Number.isInteger(step.delay_hours) || step.delay_hours < 0) {
    issues.push("Delay must be zero or a positive whole number of hours.");
  }
  if (index === 0 && step.delay_hours !== 0) {
    issues.push("The first enabled email must have no delay.");
  }

  return Array.from(new Set(issues));
}

export function validateEmailSignature(signature: OutreachEmailSignature): string[] {
  const issues: string[] = [];
  const name = signature.name.trim();
  const title = signature.title.trim();
  const details = signature.additional_details.trim();
  if (!name) issues.push("Add a signature name.");
  if (name.length > 120 || /[\r\n]/.test(signature.name)) issues.push("Keep the signature name on one line and at 120 characters or fewer.");
  if (title.length > 160 || /[\r\n]/.test(signature.title)) issues.push("Keep the signature title on one line and at 160 characters or fewer.");
  if (details.length > 1000) issues.push("Keep additional signature details at 1,000 characters or fewer.");
  if (HTML_TAG_PATTERN.test(details)) issues.push("HTML is not allowed in signature details.");
  return Array.from(new Set(issues));
}

export function validateSequence(steps: OutreachSequenceStep[]): string[] {
  const enabledSteps = steps.filter((step) => step.enabled);
  if (enabledSteps.length === 0) return ["Enable at least one email before approval."];

  let enabledIndex = 0;
  return steps.flatMap((step, index) => {
    const stepEnabledIndex = step.enabled ? enabledIndex++ : -1;
    return validateSequenceStep(step, stepEnabledIndex).map(
      (issue) => `Email ${index + 1}: ${issue}`,
    );
  });
}

export function renderLocalTemplate(
  value: string,
  restaurantName: string,
  ownerFirstName: string,
): string {
  const restaurant = restaurantName.trim() || "Sample Restaurant";
  const owner = ownerFirstName.trim();
  const greeting = owner ? `Hi ${owner},` : `Hi ${restaurant} team,`;
  const greeting01 = `Morning ${owner || `${restaurant} team`},\n\nI noticed ${restaurant} has been building a local following.`;

  return value
    .replaceAll("{{greeting}}", greeting)
    .replaceAll("{{greeting01}}", greeting01)
    .replaceAll("{{restaurant_name}}", restaurant)
    .replaceAll(WEBSITE_TOKEN, "https://tuvisolutions.com")
    .replaceAll("[GREETING]", greeting01)
    .replaceAll("[FIRST_NAME]", owner || `${restaurant} team`)
    .replaceAll("[RESTAURANT_NAME]", restaurant)
    .replaceAll("[CUISINE]", "N/A")
    .replaceAll("[CITY]", "N/A")
    .replaceAll("[RATING]", "N/A")
    .replaceAll("[TOTAL_REVIEWS]", "N/A")
    .replaceAll("[WEBSITE_URL]", "https://tuvisolutions.com");
}

export function createBlankStep(position: number): OutreachSequenceStep {
  return {
    position,
    enabled: true,
    delay_hours: position === 1 ? 0 : 72,
    subject_template: "A practical idea for [RESTAURANT_NAME]",
    body_text_template: `${position === 1 ? "[GREETING]" : "{{greeting}}"}

Write the next short message here.
`,
  };
}
