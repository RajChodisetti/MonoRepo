export type WebsitePresentationEvidence = {
  websiteScreenshot?: string;
  websiteMobileScreenshot?: string;
  websiteReview?: string;
};

export function hasWebsitePresentationEvidence(
  evidence: WebsitePresentationEvidence,
): boolean {
  return Boolean(
    evidence.websiteScreenshot?.trim() ||
      evidence.websiteMobileScreenshot?.trim() ||
      evidence.websiteReview?.trim(),
  );
}
