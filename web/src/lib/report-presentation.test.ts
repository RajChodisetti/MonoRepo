import assert from "node:assert/strict";
import test from "node:test";
import { hasWebsitePresentationEvidence } from "./report-presentation.ts";

test("website evidence renders when only a mobile capture is available", () => {
  assert.equal(
    hasWebsitePresentationEvidence({ websiteMobileScreenshot: "data:image/jpeg;base64,mobile" }),
    true,
  );
});

test("website evidence requires a capture or review", () => {
  assert.equal(hasWebsitePresentationEvidence({ websiteScreenshot: "desktop" }), true);
  assert.equal(hasWebsitePresentationEvidence({ websiteReview: "observations" }), true);
  assert.equal(hasWebsitePresentationEvidence({ websiteMobileScreenshot: "  " }), false);
  assert.equal(hasWebsitePresentationEvidence({}), false);
});
