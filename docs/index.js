// @ts-check

/**
 * One row of tournaments.csv, as parsed by Papa Parse with header: true
 * @typedef {Object} Tournament
 * @property {string} slug
 * @property {string} event
 * @property {string} start_date
 * @property {string} host
 * @property {string} version
 * @property {string} original_version
 * @property {string} legacy_flags
 * @property {string} notes
 */

/**
 * The slice of Papa Parse's API this file uses
 * typed by hand since there's no @types/papaparse install here
 * Papa is loaded globally from a CDN script tag rather than imported as a module
 * @typedef {Object} PapaStatic
 * @property {(input: string, config: {header?: boolean, skipEmptyLines?: boolean}) => {data: Tournament[]}} parse
 */

/**
 * @returns {PapaStatic}
 */
function papa() {
  // @ts-ignore -- Papa is a global from the CDN script tag, see PapaStatic above
  return window.Papa;
}

/**
 * Sets data-bs-theme before the page paints, so it never flashes the
 * wrong theme
 * priority:
 * 1. ?theme= in the URL
 * 2. system preference
 * 3. light
 * this is why index.js loads synchronously rather than deferred, and why
 * this call sits at the top of the file rather than in the
 * DOMContentLoaded block below: it has to run before <body> renders
 * @returns {"light" | "dark"}
 */
function applyInitialTheme() {
  const param = new URLSearchParams(location.search).get("theme");
  const theme = param === "dark" || param === "light"
    ? param
    : matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
  document.documentElement.setAttribute("data-bs-theme", theme);
  return theme;
}

applyInitialTheme();

/**
 * Shows the -version flag this tournament was archived with, plus the
 * deployment's own original_version in parens when it's not identical
 * e.g. "v1.9.14 (v1.9.17)" for WMUCC-2026
 * original_version in form x.y.z is wrapped with leading "v"
 * otherwise shown as-is
 * @param {string} version
 * @param {string} original
 * @returns {string} an HTML snippet
 */
function formatVersion(version, original) {
  if (!original) return `<code>${version}</code>`;
  const parts = original.split(".");
  const isRelease = parts.length === 3 && parts.every((p) => /^\d+$/.test(p));
  const originalDisplay = isRelease ? `v${original}` : original;
  if (originalDisplay === version) return `<code>${version}</code>`;
  return `<code>${version} (${originalDisplay})</code></span>`;
}

/**
 * @param {HTMLElement} tbody
 * @param {HTMLElement} status
 */
async function loadTournaments(tbody, status) {
  try {
    const res = await fetch("tournaments.csv");
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
    const text = await res.text();
    const tournaments = papa().parse(text, { header: true, skipEmptyLines: true }).data;

    tbody.innerHTML = tournaments
      .map((t) => {
        const refURL = `archive/${t.slug}/${t.slug}_reference.json`;
        return `
          <tr>
            <td>${t.event}</td>
            <td>${t.start_date}</td>
            <td><code>${t.host}</code></td>
            <td>${formatVersion(t.version, t.original_version)}</td>
            <td><a href="${refURL}"><code>${t.slug}</code></a></td>
            <td>${t.notes}</td>
          </tr>
        `;
      })
      .join("");
  } catch (err) {
    tbody.innerHTML = "";
    status.style.display = "block";
    status.textContent = `Couldn't load tournaments.csv: ${err instanceof Error ? err.message : String(err)}`;
  }
}

/** @returns {"light" | "dark"} */
function getTheme() {
  return document.documentElement.getAttribute("data-bs-theme") === "dark" ? "dark" : "light";
}

/**
 * @param {"light" | "dark"} theme
 * @param {HTMLElement} toggle
 */
function setTheme(theme, toggle) {
  document.documentElement.setAttribute("data-bs-theme", theme);
  toggle.textContent = theme === "dark" ? "☀️ Light" : "🌙 Dark";
}

/**
 * theme toggle button up
 * flipping it mirrors the choice into ?theme=
 * so the current view is a shareable, reload-stable link
 * @param {HTMLElement} toggle
 */
function initThemeToggle(toggle) {
  setTheme(getTheme(), toggle);
  toggle.addEventListener("click", () => {
    const next = getTheme() === "dark" ? "light" : "dark";
    setTheme(next, toggle);
    const url = new URL(location.href);
    url.searchParams.set("theme", next);
    history.replaceState(null, "", url);
  });
}

// index.js loads synchronously (see applyInitialTheme above), so the table
// and toggle button don't exist yet at this point in the file
// wait for DOMContentLoaded rather than relying on script placement/defer
document.addEventListener("DOMContentLoaded", () => {
  const tournamentsBody = document.getElementById("tournaments-body");
  const tournamentsStatus = document.getElementById("tournaments-status");
  if (tournamentsBody && tournamentsStatus) {
    loadTournaments(tournamentsBody, tournamentsStatus);
  }

  const themeToggle = document.getElementById("theme-toggle");
  if (themeToggle) {
    initThemeToggle(themeToggle);
  }
});
