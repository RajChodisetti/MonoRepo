import assert from "node:assert/strict";
import test from "node:test";
import { parsePreviewCoordinates } from "./report-preview.ts";

test("missing coordinate params do not become the Gulf of Guinea", () => {
  assert.deepEqual(parsePreviewCoordinates(null, null), {});
  assert.deepEqual(parsePreviewCoordinates("", ""), {});
  assert.deepEqual(parsePreviewCoordinates("-33.86", null), {});
});

test("coordinate previews require a valid pair and geographic ranges", () => {
  assert.deepEqual(parsePreviewCoordinates("-33.8688", "151.2093"), {
    latitude: -33.8688,
    longitude: 151.2093,
  });
  assert.deepEqual(parsePreviewCoordinates("91", "151"), {});
  assert.deepEqual(parsePreviewCoordinates("-33", "181"), {});
  assert.deepEqual(parsePreviewCoordinates("not-a-number", "151"), {});
});
