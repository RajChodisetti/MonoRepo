import assert from "node:assert/strict";
import { afterEach, test } from "node:test";

import {
  getActiveTemplate,
  getNextTemplate,
  getTemplateSwitchCopy,
} from "./templateConfig.ts";

const originalTemplate = process.env.TEMPLATE;

afterEach(() => {
  if (originalTemplate === undefined) {
    delete process.env.TEMPLATE;
    return;
  }
  process.env.TEMPLATE = originalTemplate;
});

test("Elysian is the default template", () => {
  delete process.env.TEMPLATE;

  assert.equal(getActiveTemplate(), "3");
});

test("invalid or empty template configuration fails safely to Elysian", () => {
  for (const value of ["", "invalid", "4"]) {
    process.env.TEMPLATE = value;
    assert.equal(getActiveTemplate(), "3");
  }

  process.env.TEMPLATE = "1";
  assert.equal(getActiveTemplate(), "1");
});

test("template previews cycle from Elysian to Aurora to Cinematic", () => {
  assert.equal(getNextTemplate("3"), "2");
  assert.equal(getNextTemplate("2"), "1");
  assert.equal(getNextTemplate("1"), "3");

  assert.equal(getTemplateSwitchCopy("3").targetLabel, "Aurora");
  assert.equal(getTemplateSwitchCopy("2").targetLabel, "Cinematic");
  assert.equal(getTemplateSwitchCopy("1").targetLabel, "Elysian");
});
