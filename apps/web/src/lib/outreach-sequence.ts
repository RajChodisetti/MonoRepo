import type { OutreachSequenceStep } from "@/lib/types";

export const WEBSITE_TOKEN = "{{website_url}}";
export const ALLOWED_MERGE_TAGS = new Set([
  "greeting",
  "greeting01",
  "restaurant_name",
  "website_url",
]);

const MERGE_TAG_PATTERN = /{{([^{}\r\n]+)}}/g;
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

  const greeting01SubjectCount = subject.split("{{greeting01}}").length - 1;
  const greeting01BodyCount = body.split("{{greeting01}}").length - 1;
  if (greeting01SubjectCount > 0) {
    issues.push("Use {{greeting01}} only in the message body.");
  }
  if (greeting01BodyCount > 0 && index !== 0) {
    issues.push("Use {{greeting01}} only in the first enabled email.");
  }
  if (greeting01BodyCount > 1) {
    issues.push("Use {{greeting01}} exactly once.");
  }
  if (greeting01BodyCount === 1 && body.includes("{{greeting}}")) {
    issues.push("Replace {{greeting}} when using {{greeting01}}.");
  }

  if (!Number.isInteger(step.delay_hours) || step.delay_hours < 0) {
    issues.push("Delay must be zero or a positive whole number of hours.");
  }
  if (index === 0 && step.delay_hours !== 0) {
    issues.push("The first enabled email must have no delay.");
  }

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
  const greeting01 = `${greeting}\n\nI came across ${restaurant} while looking at local restaurants.\nI thought it was worth reaching out directly to your team.`;

  return value
    .replaceAll("{{greeting}}", greeting)
    .replaceAll("{{greeting01}}", greeting01)
    .replaceAll("{{restaurant_name}}", restaurant)
    .replaceAll(WEBSITE_TOKEN, "");
}

export function createBlankStep(position: number): OutreachSequenceStep {
  return {
    position,
    enabled: true,
    delay_hours: position === 1 ? 0 : 72,
    subject_template: "A practical idea for {{restaurant_name}}",
    body_text_template: `{{greeting}}

Write the next short message here.
`,
  };
}
