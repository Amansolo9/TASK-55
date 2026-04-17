import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import * as cheerio from "cheerio";

const here = dirname(fileURLToPath(import.meta.url));
const viewsDir = resolve(here, "..", "views");
const load = (name: string) =>
  cheerio.load(readFileSync(resolve(viewsDir, name), "utf8"));

describe("login.html", () => {
  const $ = load("login.html");

  it("renders a login form posting to /login with username and password inputs", () => {
    const form = $('form[action="/login"]');
    expect(form.length).toBe(1);
    expect(form.attr("method")).toBe("post");
    expect(form.attr("hx-post")).toBe("/login");
    expect(form.find('input[name="username"][required]').length).toBe(1);
    expect(form.find('input[name="password"][type="password"][required]').length).toBe(1);
  });

  it("renders a registration form posting to /register", () => {
    const form = $('form[action="/register"]');
    expect(form.length).toBe(1);
    expect(form.find('input[name="username"][required]').length).toBe(1);
    expect(form.find('input[name="password"][type="password"][required]').length).toBe(1);
  });

  it("provides HTMX swap targets for inline results", () => {
    expect($("#login-result").length).toBe(1);
    expect($("#register-result").length).toBe(1);
  });
});

describe("change_password.html", () => {
  const $ = load("change_password.html");

  it("posts to the change-password API via HTMX and enforces min length 12", () => {
    const form = $('form[hx-post="/api/auth/change-password"]');
    expect(form.length).toBe(1);
    const newPw = form.find('input[name="new_password"]');
    expect(newPw.attr("type")).toBe("password");
    expect(newPw.attr("minlength")).toBe("12");
    expect(newPw.attr("required")).toBeDefined();
  });
});

describe("index.html (dashboard)", () => {
  const $ = load("index.html");

  it("shows sign-in prompt when no user present", () => {
    expect($('a[href="/login"]').length).toBeGreaterThan(0);
  });

  it("contains conditional welcome branch for authenticated users", () => {
    const raw = readFileSync(resolve(viewsDir, "index.html"), "utf8");
    expect(raw).toContain("{{if .User}}");
  });
});

describe("members.html", () => {
  const $ = load("members.html");

  it("has create-member form posting to /api/members with required PII fields", () => {
    const form = $('form[hx-post="/api/members"]');
    expect(form.length).toBe(1);
    for (const field of ["full_name", "email", "phone"]) {
      expect(form.find(`input[name="${field}"][required]`).length).toBe(1);
    }
  });

  it("has import form with multipart encoding", () => {
    const form = $('form[hx-post="/api/members/import"]');
    expect(form.length).toBe(1);
    expect(form.attr("hx-encoding")).toBe("multipart/form-data");
    expect(form.find('input[type="file"][name="file"][required]').length).toBe(1);
  });

  it("has export form using GET to /api/members/export", () => {
    const form = $('form[action="/api/members/export"][method="get"]');
    expect(form.length).toBe(1);
  });
});

describe("layouts/main.html", () => {
  const raw = readFileSync(resolve(viewsDir, "layouts", "main.html"), "utf8");
  const $ = cheerio.load(raw);

  it("loads htmx and tailwind vendor bundles locally (offline-capable)", () => {
    expect($('script[src="/static/vendor/htmx.min.js"]').length).toBe(1);
    expect($('script[src="/static/vendor/tailwindcss.js"]').length).toBe(1);
  });

  it("wires a logout form with a data-csrf-field hidden input", () => {
    const form = $('form[action="/logout"][method="post"]');
    expect(form.length).toBe(1);
    expect(form.find("input[data-csrf-field]").length).toBe(1);
  });

  it("renders role-gated nav links via template conditionals", () => {
    for (const needle of [
      '/budgets', '/reviews', '/credits', '/members', '/clubs', '/regions', '/mdm', '/users', '/flags',
    ]) {
      expect(raw).toContain(`href="${needle}"`);
    }
    expect(raw).toContain('eq .User.Role "admin"');
    expect(raw).toContain('eq .User.Role "organizer"');
    expect(raw).toContain('eq .User.Role "team_lead"');
  });

  it("embeds a client-side script that attaches X-CSRF-Token on htmx requests", () => {
    expect(raw).toContain("htmx:configRequest");
    expect(raw).toContain("X-CSRF-Token");
    expect(raw).toContain("readCookie('csrf_token')");
  });
});
