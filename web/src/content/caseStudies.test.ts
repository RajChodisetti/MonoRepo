import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { caseStudies, caseStudyPath, getCaseStudy } from "./caseStudies.ts";

const publicDirectory = fileURLToPath(new URL("../../public/", import.meta.url));

test("case studies have complete, unique, locally backed content", () => {
  assert.equal(caseStudies.length, 7);
  assert.equal(new Set(caseStudies.map((study) => study.slug)).size, caseStudies.length);
  assert.equal(new Set(caseStudies.map((study) => study.imageUrl)).size, caseStudies.length);

  for (const study of caseStudies) {
    const requiredStrings = [
      study.slug,
      study.name,
      study.restaurant,
      study.role,
      study.metricValue,
      study.metricDescription,
      study.location,
      study.summary,
      study.challenge,
      study.quote,
    ];

    assert.ok(requiredStrings.every((value) => value.trim().length > 0));
    assert.equal(study.approach.length, 3);
    assert.equal(study.results.length, 3);
    assert.match(study.imageUrl, /^\/owners\/case-studies\/[a-z0-9-]+\.jpg$/);
    assert.ok(existsSync(`${publicDirectory}${study.imageUrl}`), study.imageUrl);
    assert.equal(caseStudyPath(study.slug), `/resources/case-studies/${study.slug}`);
    assert.equal(getCaseStudy(study.slug), study);
  }
});
