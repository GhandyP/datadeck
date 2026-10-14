# DataDeck

Turn a table of numbers into a finished HTML slide deck whose every figure
traces back to the data you supplied.

You paste a table, name the deck, and get a plan in which each claim carries the
code of the source values it rests on — plus one self-contained `.html` file you
open by double-clicking it.

```text
Region,Revenue,Change
North America,4820000,+6.4%
Europe,3140000,-2.1%
LATAM,1275000,+11.8%
APAC,2460000,+0.4%
```

## The guarantee

A model writes the story. It never writes the numbers.

Every figure on a slide is placed by the program from your table and re-checked
against it before the deck is rendered. A claim that states a number your data
does not contain is **rejected with the reason** and kept off the deck:

```text
C1  Total revenue reached 11,695,000.        Verified   F-05
C2  LATAM grew fastest, at 11.8%.            Verified   F-10
C3  Europe declined 7.3%.                    Rejected   F-09
    the claim states 7.3%, which the cited source data
    does not contain; it says -2.1%
```

The deck only ever shows `Verified` claims. You see the rejected ones in the app,
with the reason, because a silent deletion would hide exactly what you need to
know before presenting.

## How it works

```text
  paste ──▶ internal/data ──▶ internal/agent ──▶ internal/plan ──▶ internal/deck
            parse + facts     ADK v2 + Gemini     validate every     one self-contained
            (deterministic)   (language only)     claim vs. facts    HTML file
```

| Package | Job | Why it is separate |
|---|---|---|
| `internal/data` | Parses delimited text, Markdown tables and JSON records; computes facts and infers each column's own number convention. | Deterministic. No credentials needed to test it, and it is where every value is born. |
| `internal/agent` | The only place a model is called: one Google ADK v2 agent, one output schema, no tools. | The prompt and the schema are pure functions, so the whole model boundary is testable without a network. |
| `internal/plan` | Holds the deck contract and the rule that makes it honest: a claim may only state a figure its cited evidence contains. | The validation is the product. It is pure, table-driven and exhaustively tested. |
| `internal/deck` | Renders the plan as one HTML file: a shell plus the plan as JSON. | A rejected claim cannot reach a slide because the renderer only ever receives verified ones. |
| `internal/theme` | The tokens, the shared primitives and the two embedded faces both surfaces are built from. | The two surfaces cannot drift apart, and neither depends on a font host. |

Facts are computed with exact integer arithmetic scaled by each column's decimal
count, not through floats, so a total is the sum of what you supplied. A column
written `1,234.50` and one written `1.234,50` each keep their own convention, and
a column of percentages never gets a meaningless sum.

The exported deck carries its fonts inside the file as data URIs, so it renders
in its own voice on a machine that has none of them installed — and still reaches
no host.

## Requirements

- Go 1.26 or newer.
- `GOOGLE_API_KEY` for a Gemini model. **Generation is credential-bound by
  product decision**: without a key DataDeck cannot build a deck by any other
  route. It still starts, and the page states exactly what is missing rather
  than offering a control that cannot work.

## Run

```bash
export GOOGLE_API_KEY="your-key"
go run .
```

Open `http://127.0.0.1:8080`. Flags: `-addr` (listen address), `-model`
(defaults to `ADK_MODEL`, then `gemini-flash-latest`).

## Test

```bash
go test ./...
```

The suite needs no credentials and makes no network call: the HTTP surface is
exercised with a fake planner, and the model boundary is tested through its
prompt and schema.

## What it is not

- Not a `.pptx` or PDF generator, and not a publishing integration.
- Not a dashboard: it produces a document you review and present.
- Not autonomous: every deck is marked a draft, and a human has to approve its
  claims before anyone sees it.

## Accessibility

The app and the exported deck target WCAG 2.2 AA: keyboard-operable controls and
slide navigation, visible focus, AA contrast in every shipped colour, status
carried by a word as well as a colour, and `prefers-reduced-motion` respected.
The deck prints one slide per page.

## Licence

MIT — see [`LICENSE`](LICENSE).
