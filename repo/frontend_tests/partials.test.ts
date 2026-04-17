import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import * as cheerio from "cheerio";

const here = dirname(fileURLToPath(import.meta.url));
const partialsDir = resolve(here, "..", "views", "partials");
const raw = (name: string) => readFileSync(resolve(partialsDir, name), "utf8");
const load = (name: string) => cheerio.load(raw(name));

describe("budgets_list partial", () => {
  const src = raw("budgets_list.html");
  const $ = load("budgets_list.html");

  it("renders a table with the expected budget columns", () => {
    const headers = $("thead th").map((_, el) => $(el).text().trim()).get();
    for (const col of ["ID", "Acct/Campus/Project", "Period", "Amount", "Spent", "Execution", "Remaining", "Alert"]) {
      expect(headers).toContain(col);
    }
  });

  it("shows an empty-state row when no budgets are present", () => {
    expect(src).toContain("No budgets yet.");
    expect(src).toContain("{{else}}");
  });
});

describe("budget_changes partial", () => {
  const src = raw("budget_changes.html");

  it("iterates over .Changes with a fallback empty state", () => {
    expect(src).toContain("{{range .Changes}}");
    expect(src).toContain("{{else}}");
  });
});

describe("budget_projection_result partial", () => {
  const src = raw("budget_projection_result.html");

  it("surfaces the projected end balance label", () => {
    expect(src).toContain("Projected end balance");
  });

  it("renders tone classes from the handler-provided context", () => {
    expect(src).toContain("{{.ToneBorder}}");
    expect(src).toContain("{{.ToneBG}}");
    expect(src).toContain("{{.ToneText}}");
  });
});

describe("reviews_list partial", () => {
  const src = raw("reviews_list.html");

  it("iterates over .Reviews with empty-state branch", () => {
    expect(src).toContain("{{range .Reviews}}");
    expect(src).toContain("{{else}}");
  });
});

describe("fulfilled_orders_options partial", () => {
  const src = raw("fulfilled_orders_options.html");

  it("emits <option> entries from .Orders", () => {
    expect(src).toContain("{{range .Orders}}");
    expect(src).toMatch(/<option[^>]*value="\{\{\.ID\}\}"/);
  });
});
