# DataDeck — frozen contract for implementers

Read `PRODUCT.md`, `.impeccable/surface-brief.md` and `web/tokens.css` first.
This file is the interface contract: do not invent values, endpoints, field
names or file paths outside it.

## 1. Visual system — Exhibit & Chain of Custody

Every colour, font and spacing value comes from `web/tokens.css`. Never write a
raw hex value in a component: use the token.

- **Shell** `--shell`: the exhibit band across the top, navigation, the primary
  action, selection. This is the page's one saturated field.
- **Ground** `--ground` with `--panel` areas separated by hairline `--rule`.
- **Manila** `--manila` / `--manila-wash`: **provenance only.** It marks where a
  figure came from and appears nowhere else on either surface.
- **States**: `--alert` for a rejected claim, `--verified` for a validated one.
- **Type**: `--font-text` (Archivo) for all prose and controls. `--font-code`
  (Archivo Narrow) only for codes, step numbers, and column headers, uppercase
  and tracked — use the `.dd-code` primitive. Numbers use `.dd-num`.
- **Density**: square corners (`--radius: 0`), hairline rules, **no shadows**,
  no gradients, no rounded cards. Rows are ruled like a case file.
- **Signature move — visible provenance**: a figure or claim carries a code
  (`F-01`, `C2`) in a `.dd-prov` mark. Selecting it reveals exactly which source
  cells support it, and those cells light in manila. This is the product's whole
  claim made visible; it must work by keyboard, not only by pointer.

Accessibility floor: WCAG 2.2 AA contrast, visible focus via `:focus-visible`,
every control reachable and operable by keyboard, status never carried by colour
alone (always with a word: `Verified`, `Rejected`), `prefers-reduced-motion`
respected. Manila is too light to read as text on its own wash, so text that
must be manila sets in `--manila-ink` while the base keeps the rule and the wash
fills.

Neither surface fetches a font. `internal/theme/fonts/` holds Archivo 400, Archivo
600 and Archivo Narrow 600; the app serves them from its own origin at
`/fonts/…` and the exported deck inlines them as data URIs, so the deck carries
its own voice inside the one file it ships as. Text set in `--font-code` is
uppercase and tracked; everything else uses `--font-text`.

## 2. HTTP API — implemented by `main.go`, consumed by `web/app.js`

`GET /api/health` → `200`

```json
{ "status": "ok", "model": "gemini-flash-latest", "apiKeyConfigured": true }
```

`POST /api/generate` request body:

```json
{ "title": "Q3 2026 revenue review", "audience": "Board", "data": "<raw pasted text>" }
```

`200` response:

```json
{
  "code": "EX-7F3A21",
  "title": "Q3 2026 revenue review",
  "audience": "Board",
  "status": "draft",
  "source": {
    "columns": ["Region", "Revenue", "Change"],
    "rows": [["North America", "4820000", "+6.4%"]]
  },
  "facts": [
    { "id": "F-01", "label": "North America · Revenue", "value": "4,820,000",
      "row": 0, "column": 1 }
  ],
  "slides": [
    { "number": 1, "title": "Revenue grew in three of four regions",
      "notes": "One sentence of speaker guidance.",
      "claims": [
        { "id": "C1", "text": "Total revenue reached 11,695,000.",
          "evidenceIds": ["F-09"], "status": "verified" },
        { "id": "C3", "text": "Europe declined 7.3%.", "evidenceIds": ["F-04"],
          "status": "rejected",
          "reason": "the claim states 7.3%, which the cited source data does not contain; it says -2.1%" }
      ] }
  ],
  "rejectedCount": 1,
  "deckHTML": "<!doctype html>…"
}
```

- `source` is the parsed table, so the app can show it and light the cells a
  claim rests on.
- `facts[].row` / `facts[].column` are the zero-based indices of the source cell
  the fact came from. `row: -1` means the fact is computed over the whole
  column, so the app lights that column's header and every cell in it.
- `claims[].status` is `"verified"` or `"rejected"`. A rejected claim is shown in
  the app with its `reason` and is **excluded from the deck**; a deck whose every
  claim was rejected still answers `200` with `"deckHTML": ""`, so the reasons
  reach the user instead of an error replacing them.
- Errors: `400 {"error":"…","kind":"input"}` for unusable input;
  `502 {"error":"…","kind":"missing_api_key"}` when `GOOGLE_API_KEY` is absent;
  `502 {"error":"…","kind":"model"}` for a model failure.
  Every error body is JSON and `error` is a full sentence naming what failed and
  what the user can do.

The shared tokens live in `internal/theme/tokens.css` and are served at
`/tokens.css`, so the app and the exported deck cannot drift apart.

## 3. App shell files — `web/`

- `web/index.html` — one page, three steps in the exhibit band:
  `1 Source`, `2 Plan`, `3 Deck`. Step 2 active while claims are shown.
- `web/app.css` — layout and components, tokens only.
- `web/app.js` — no framework, no build step, ES modules not required. Fetches
  `/api/health` on load and shows the exact configuration message when
  `apiKeyConfigured` is false; posts `/api/generate`; renders the source table
  (parsed from the pasted text client-side for the live preview is **not**
  required — the server response's `facts` carry the row/column indices), the
  claim list with provenance marks and states, and the deck summary with the
  download action. The download is a `Blob` of `deckHTML` saved as
  `datadeck.html`.

## 4. Exported deck — `internal/deck/assets/`

The deck is one self-contained file. It is a static shell that the Go renderer
composes; the shell declares exactly three markers and the renderer replaces them:

- `<!--DATADECK:STYLE-->` — replaced with the inlined token and deck CSS.
- `<!--DATADECK:DATA-->` — replaced with `window.__DATADECK__ = <json>;`.
- `<!--DATADECK:SCRIPT-->` — replaced with the inlined deck script.

Files:

- `deck.html` — the shell, with those three markers, and nothing fetched from
  any external host.
- `deck.css` — the deck's own styles, tokens only, wider measure than the app.
- `deck.js` — reads `window.__DATADECK__`, renders the slides into the shell,
  and implements: one slide owning the viewport, keyboard navigation
  (`ArrowLeft`/`ArrowRight`/`Home`/`End`), the provenance reveal on each claim
  (pointer and keyboard), a slide index, and a print stylesheet hook. No
  framework, no network.

The deck's data shape is exactly the `200` response above, minus `deckHTML`.
