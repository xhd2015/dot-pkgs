---
name: playwright-debug
description: >-
  Run Playwright scripts for browser automation and frontend debugging.
  Use when the user wants to test a web page, debug UI behavior, monitor
  network requests, or automate browser interactions.
---

# Playwright Debug Skill

A CLI tool for running Playwright browser automation scripts with automatic setup.

## Commands

### run — Run a Playwright script

Runs a JavaScript snippet in a headless Chromium browser. The script is automatically wrapped with browser/page setup.

Available variables in your script:
- `browser` — Chromium browser instance
- `page` — A new browser page
- `chromium` — Playwright's chromium launcher

```bash
# Navigate to a page and print the title
playwright-debug 'await page.goto("https://example.com"); console.log(await page.title());'

# Monitor network requests
playwright-debug '
  page.on("request", req => console.log("[REQ]", req.method(), req.url()));
  page.on("requestfailed", req => console.log("[FAIL]", req.url(), req.failure()?.errorText));
  await page.goto("https://example.com");
  await page.waitForTimeout(5000);
'

# Fill a form and submit
playwright-debug '
  await page.goto("https://example.com/login");
  await page.fill("input[name=email]", "test@example.com");
  await page.fill("input[name=password]", "secret");
  await page.click("button[type=submit]");
  await page.waitForURL("**/dashboard");
  console.log("Logged in:", await page.title());
'
```