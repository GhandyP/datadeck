# DataDeck

Turn a table of numbers into a finished HTML slide deck whose every figure
traces back to the data you supplied.

**Status: implementation in progress.** The repository currently holds the
product contract and the work plan
([`odd/tasks/datadeck.md`](odd/tasks/datadeck.md)); the Go source lands as the
plan's tasks complete. Nothing is published or pushed yet.

## What it does

You paste or upload tabular data — delimited text, a JSON array of records, or a
Markdown table — and name the deck. DataDeck parses and normalizes the data,
computes deterministic facts over it, and asks one Google ADK agent for a slide
narrative constrained by an output schema. Every claim the agent returns is then
re-checked against the input: a slide may only state a figure that the supplied
data contains. A claim that cites anything else is rejected with the reason,
instead of being quietly printed.

The result is one self-contained `.html` file: inline CSS and JS, keyboard
navigation, a print stylesheet, and no external asset hosts. Open it by
double-clicking it.

## Position

The model writes the story. It never writes the numbers. Every figure in a deck
is placed by the program from your data and validated before rendering, so a
model that invents a value produces a visible failure rather than a confident
slide.

## Requirements

- Go 1.26 or newer.
- `GOOGLE_API_KEY` for a Gemini model. Generation is deliberately
  credential-bound: without a key the app starts and reports the missing
  configuration rather than generating a deck by another route.

## Run

```bash
export GOOGLE_API_KEY="your-key"
go run .
```

Then open the address the server prints.

`ADK_MODEL` overrides the default model name.

## What it is not

- Not a `.pptx` or PDF generator, and not a publishing integration.
- Not a dashboard: it produces a document you review and present.
- Not autonomous: every deck is marked a draft and requires a human to approve
  its claims before it is shown to anyone.

## Licence

MIT. See [`LICENSE`](LICENSE).
