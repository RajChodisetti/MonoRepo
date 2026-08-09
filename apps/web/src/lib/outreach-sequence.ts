import type { OutreachSequenceStep } from "@/lib/types";

export const WEBSITE_TOKEN = "{{website_url}}";
export const UNSUBSCRIBE_TOKEN = "{{unsubscribe_url}}";
export const ALLOWED_MERGE_TAGS = new Set([
  "greeting",
  "restaurant_name",
  "website_url",
  "unsubscribe_url",
]);

const MERGE_TAG_PATTERN = /{{\s*([a-z_]+)\s*}}/gi;
const HTML_TAG_PATTERN = /<\s*\/?\s*[a-z][^>]*>/i;
const RAW_LINK_PATTERN = /(?:https?:\/\/|www\.)[^\s<]+|\b(?:[a-z0-9-]+\.)+[a-z]{2,}(?:\/[^\s<]*)?/i;

function occurrences(value: string, token: string): number {
  return value.split(token).length - 1;
}

export function countTemplateLinks(step: OutreachSequenceStep): number {
  const combined = `${step.subject_template}\n${step.body_text_template}`;
  const mergeLinks =
    occurrences(combined, WEBSITE_TOKEN) +
    occurrences(combined, UNSUBSCRIBE_TOKEN);
  const rawLinks =
    combined.match(new RegExp(RAW_LINK_PATTERN.source, "gi"))?.length ?? 0;
  return mergeLinks + rawLinks;
}

export function validateSequenceStep(
  step: OutreachSequenceStep,
  index: number,
): string[] {
  if (!step.enabled) return [];

  const issues: string[] = [];
  const subject = step.subject_template.trim();
  const body = step.body_text_template.trim();

  if (!subject) issues.push("Add a subject.");
  if (subject.length > 200) issues.push("Keep the subject at 200 characters or fewer.");
  if (!body) issues.push("Add a plain-text message.");
  if (HTML_TAG_PATTERN.test(subject) || HTML_TAG_PATTERN.test(body)) {
    issues.push("HTML is not allowed; use plain text only.");
  }

  const unknownTags = Array.from(`${subject}\n${body}`.matchAll(MERGE_TAG_PATTERN))
    .map((match) => match[1].toLowerCase())
    .filter((tag) => !ALLOWED_MERGE_TAGS.has(tag));
  if (unknownTags.length > 0) {
    issues.push(`Unknown merge tag: {{${Array.from(new Set(unknownTags)).join("}}, {{")}}}.`);
  }

  if (occurrences(body, WEBSITE_TOKEN) !== 1) {
    issues.push(`Include ${WEBSITE_TOKEN} exactly once in the message.`);
  }
  if (occurrences(body, UNSUBSCRIBE_TOKEN) !== 1) {
    issues.push(`Include ${UNSUBSCRIBE_TOKEN} exactly once in the message.`);
  }
  if (subject.includes(WEBSITE_TOKEN) || subject.includes(UNSUBSCRIBE_TOKEN)) {
    issues.push("Keep both links in the message body, not the subject.");
  }
  if (RAW_LINK_PATTERN.test(`${subject}\n${body}`)) {
    issues.push("Remove typed URLs; use only the website and unsubscribe merge tags.");
  }
  if (countTemplateLinks(step) !== 2) {
    issues.push("Each enabled email must render exactly two links.");
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

  return value
    .replaceAll("{{greeting}}", greeting)
    .replaceAll("{{restaurant_name}}", restaurant)
    .replaceAll(WEBSITE_TOKEN, "https://tuvisolutions.com")
    .replaceAll(
      UNSUBSCRIBE_TOKEN,
      "https://tuvisolutions.com/unsubscribe/sample-token",
    );
}

export function createBlankStep(position: number): OutreachSequenceStep {
  return {
    position,
    enabled: true,
    delay_hours: position === 1 ? 0 : 72,
    subject_template: "A practical idea for {{restaurant_name}}",
    body_text_template: `{{greeting}}

Write the next short message here.

See how we help restaurants:
{{website_url}}

Best,
The Tuvi Solutions team

Business outreach from Tuvi Solutions
Opt out: {{unsubscribe_url}}`,
  };
}
