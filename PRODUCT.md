# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Go standard-library HTTP server serving embedded static HTML/CSS/JS (no frontend
framework, no npm). Generation uses Google ADK v2 (`google.golang.org/adk/v2`)
with a Gemini model and requires `GOOGLE_API_KEY`. One binary, `go run .`.

## Users

Primary: an analyst, consultant, or founder who has a table of numbers (KPI
snapshot, survey result, monthly metrics) and has to present it to people who
will not read the table. They are at a desk, minutes before a review, and the
deck is the deliverable.

Secondary: an engineer or hiring reviewer reading this repository as part of a
public portfolio, deciding in one pass whether the author ships careful work.

## Product Purpose

Turn supplied tabular data into a finished, self-contained HTML slide deck whose
every stated number traces back to a supplied value. Success means the deck can
be opened by double-clicking one file on any machine, and no slide states a
figure the input does not contain.

## Positioning

The model writes the story; it never writes the numbers. Every figure in a deck
is copied from the supplied data by the program and re-checked against the input
before the deck is rendered, so a model that hallucinates a value produces a
visible validation failure instead of a confident slide.

## Operating Context

The user pastes or uploads CSV, TSV, JSON records, or a Markdown table, names
the deck and its audience, and picks a visual theme. The server parses and
normalizes the data, computes deterministic facts (totals, shares, changes),
asks one ADK agent for a slide narrative constrained by an output schema,
validates every claim against the evidence, and renders the deck in memory. The
result is previewed in the browser and downloaded as one `.html` file.

## Capabilities and Constraints

- Input: delimited text (comma, semicolon, tab, or pipe), JSON arrays of
  records, and Markdown tables.
- Output: one self-contained HTML file with inline CSS and JS, keyboard
  navigation, and a print stylesheet. No `.pptx`, no PDF, no publishing step,
  no external asset hosts.
- Generation requires `GOOGLE_API_KEY`; without it the app starts and reports
  the missing configuration instead of pretending to work. There is no offline
  generation path by explicit product decision.
- The deck is always marked as a draft requiring human review before it is
  presented.
- Runs locally: `go run .` on a single port. No deployment, no database, no
  user accounts, no telemetry.

## Brand Commitments

Name: DataDeck. The tone is technical and plain: the product states what it did
and what it refused to do, and never advertises itself on the deck's slides.

## Evidence on Hand

None. There are no customers, no testimonials, no benchmarks, and no published
performance numbers; the repository must never invent them. Example fixtures in
`testdata/` are synthetic and are labelled as such.

## Product Principles

1. A number that is not in the input is never on a slide.
2. The model contributes language, never figures; the program owns every value.
3. One portable file beats a pipeline of integrations.
4. Every refusal is explained to the user in the terms of their data.
5. A draft is a draft: the product never claims a human approved the deck.

## Accessibility & Inclusion

The app and the rendered deck target WCAG 2.2 AA: keyboard-operable controls and
slide navigation, visible focus, AA contrast in every shipped theme, no
information carried by colour alone, and `prefers-reduced-motion` respected.
