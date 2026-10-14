---
name: DataDeck
description: A case file for numbers: every figure on screen carries the label that says where it came from.
colors:
  shell: "#20348F"
  shell-deep: "#17265F"
  shell-ink: "#FFFFFF"
  shell-wash: "#E8ECF8"
  ground: "#E7E8E6"
  panel: "#FFFFFF"
  rule: "#C9CCC7"
  rule-strong: "#9AA19B"
  ink: "#15181B"
  ink-soft: "#5A6068"
  manila: "#C98B2E"
  manila-wash: "#F2E3C6"
  alert: "#B3261E"
  alert-wash: "#F7DEDA"
  verified: "#1E6B4F"
  verified-wash: "#DCEBE3"
typography:
  slide-title:
    fontFamily: "Archivo, system-ui, sans-serif"
    fontSize: "clamp(1.9rem, 3.6vw, 3rem)"
    fontWeight: 600
    lineHeight: 1.1
    letterSpacing: "-0.01em"
  claim:
    fontFamily: "Archivo, system-ui, sans-serif"
    fontSize: "clamp(1.05rem, 1.6vw, 1.35rem)"
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: "normal"
  lead:
    fontFamily: "Archivo, system-ui, sans-serif"
    fontSize: "17px"
    fontWeight: 700
    lineHeight: 1.3
    letterSpacing: "-0.01em"
  body:
    fontFamily: "Archivo, system-ui, sans-serif"
    fontSize: "16px"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  read:
    fontFamily: "Archivo, system-ui, sans-serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  ui:
    fontFamily: "Archivo, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: "normal"
  meta:
    fontFamily: "Archivo, system-ui, sans-serif"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: "normal"
  data-input:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: "normal"
  code:
    fontFamily: "Archivo Narrow, Archivo, system-ui, sans-serif"
    fontSize: "12px"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "0.09em"
  micro:
    fontFamily: "Archivo Narrow, Archivo, system-ui, sans-serif"
    fontSize: "11px"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "0.09em"
rounded:
  none: "0"
spacing:
  step: "4px"
  gutter: "16px"
components:
  button-primary:
    backgroundColor: "{colors.shell}"
    textColor: "{colors.shell-ink}"
    rounded: "{rounded.none}"
    padding: "9px 14px"
  button-primary-hover:
    backgroundColor: "{colors.shell-deep}"
    textColor: "{colors.shell-ink}"
  control:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "6px 14px"
  control-active:
    backgroundColor: "{colors.shell}"
    textColor: "{colors.shell-ink}"
  claim-provenance:
    backgroundColor: "{colors.manila-wash}"
    textColor: "{colors.manila}"
    rounded: "{rounded.none}"
    padding: "3px 8px"
  claim-provenance-open:
    backgroundColor: "{colors.manila}"
    textColor: "{colors.panel}"
  field:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "7px 9px"
---

# Design System: DataDeck

## Overview

DataDeck is a case file, not a dashboard. The visitor pastes a table, and every
figure that appears afterwards carries the label saying which cells it came from.
The category's default — a grid of neutral cards with one blue accent — is
refused on purpose.

Two surfaces share this system: the local app shell, where the work happens, and
the deck it exports, which is a document and wears a wider measure. They load the
same tokens from `internal/theme/tokens.css`, so they cannot drift apart.

## Colors

### Primary

`shell #20348F` is the page's one saturated field. It carries the exhibit band
across the top, the primary action, the active step, and selection. `shell-deep
#17265F` is its pressed state and the band's bottom rule; `shell-wash #E8ECF8`
is the inactive step label on the band; `shell-ink #FFFFFF` is text on it.

### Secondary

`manila #C98B2E` with its wash `#F2E3C6` means provenance and nothing else. It
marks the code on a claim, the block that reveals which source values support it,
and the source cells a selected claim rests on.

### Tertiary

`verified #1E6B4F` and `alert #B3261E` are claim states, each with its own wash
(`#DCEBE3`, `#F7DEDA`). They are never used decoratively, and a state is always
carried by a word as well as the colour.

### Neutral

`ground #E7E8E6` is the file-room floor; `panel #FFFFFF` is every work area;
`rule #C9CCC7` and `rule-strong #9AA19B` are the hairlines that do all the
separating; `ink #15181B` and `ink-soft #5A6068` are text.

### Named Rules

**The manila rule.** Manila is spent on provenance and nowhere else. The moment
it decorates something that is not a source citation, the system's one
distinctive idea stops meaning anything.

**The word rule.** Colour never carries a state alone. Every verified or rejected
claim also says so in words, in `code` type.

## Typography

Archivo carries all prose and every control. Archivo Narrow appears only for
codes, step numbers and column headers, always uppercase and tracked at
`0.09em` — the `.dd-code` primitive. All figures use tabular numerals
(`.dd-num`), so columns of numbers line up the way a case file's do.

### Hierarchy

One ramp, seven steps, and nothing between them. The two fluid steps are the
deck's voices; the fixed ones are the app's working sizes.

| Step | Size | Where it is used |
|---|---|---|
| `slide-title` | `clamp(1.9rem, 3.6vw, 3rem)` | The deck's slide heading, the largest voice in the system |
| `claim` | `clamp(1.05rem, 1.6vw, 1.35rem)` | The statement a slide makes |
| `lead` | 17px | The product name in the exhibit band |
| `body` | 16px | The app's working size |
| `read` | 15px | Deck band, speaker notes, notices — anything read once |
| `ui` | 14px | Controls, fields, table cells, claim text in the app |
| `meta` | 13px | Labels, hints, reasons, summaries |
| `code` | 12px | Provenance marks, source ids, area sub-headings |
| `micro` | 11px | Column headers, the claim's status word |
| `data-input` | 13px mono | The pasted-data field only |

`data-input` is the system's one monospace voice, and it is justified: the field
is where a user pastes columns of numbers, and monospace is what keeps those
columns readable while they are still text. It is the platform stack, not a
fetched face, because the field holds the user's own characters and nothing
about the design should reinterpret them.

### Named Rules

**The narrow rule.** Archivo Narrow marks machine identity — a code, a step, a
column header. Prose never sets in it.

**The ramp rule.** A size outside the table above is a new step: add it here or
do not use it. The system has no in-between sizes.

## Layout

The app is three ruled work areas on one row — source, plan, deck — collapsing to
a single stacked column under 980px. The deck shows one slide owning the viewport
with the chrome retreated to the band and the controls, and prints one slide per
page.

Spacing is a 4px step with a 16px gutter. Areas are separated by a hairline
border rather than by space or shadow.

## Elevation & Depth

There is no elevation. No shadow, no blur, no gradient, no translucency anywhere
in either surface. Depth is expressed the way a case file expresses it: hairline
rules separate, and a wash marks the thing you are looking at. A selected claim
takes the manila wash; a rejected one takes the alert wash; a source cell takes
the manila wash with a manila left edge.

## Shapes

Square. `--radius: 0` everywhere: inputs, buttons, panels, the provenance marks,
the slide index. Borders are 1px; the only heavier rule in the system is the 2px
`shell-deep` line under the exhibit band. Nothing is pill-shaped, and nothing
floats.

## Components

### Buttons

`button-primary` is a full-width shell block with white text — the one action per
area. `control` is a square panel button used for deck navigation and the slide
index; `control-active` inverts it to shell. Both are keyboard-operable with a
visible `:focus-visible` ring at 2px shell.

### Chips

`claim-provenance` is the signature control: a manila-outlined mark in `code`
type holding the fact ids a claim cites. Opened, it inverts to solid manila with
panel text. It is a real `<button>` with `aria-pressed` / `aria-expanded`, so the
reveal works from the keyboard.

### Cards / Containers

There are no cards. A work area is a panel separated by hairlines; a slide is a
full viewport; a source block is a hairline box in manila. Containers never round,
never lift, and never nest more than one level.

### Inputs / Fields

`field` is a square panel input with a `rule-strong` hairline. Labels sit above
in 13px semibold; hints below in 12px `ink-soft`. The pasted-data textarea sets in
the platform monospace stack, because that is what a user pastes into it.

### Navigation

The exhibit band carries three steps — Source, Plan, Deck — with the active one
marked by a solid `shell-ink` rule beneath it and a heavier weight. In the
exported deck, navigation is the slide index: square numbered buttons with
`aria-current` on the current slide, plus Previous/Next and the arrow keys.

### Signature Component

The provenance reveal. A claim shows its evidence ids as a manila mark; opening
it prints the cited facts as ruled rows — id, label, value — and lights exactly
those cells in the source table (a fact computed over a column lights the whole
column, header included). It is the product's guarantee made visible: the figure
and the cell it came from are one gesture apart.

## Do's and Don'ts

### Do:

- Use a token for every colour, size and rule width. Never write a raw hex value.
- Separate with hairlines and mark with washes.
- Keep manila for provenance, and always pair a state colour with its word.
- Keep tabular numerals on every figure.
- Keep the reveal operable by keyboard, and announce what it revealed.

### Don't:

- Don't add shadows, gradients, blur, rounded corners, or pill shapes.
- Don't spend manila on decoration, headings, or the band.
- Don't set prose in Archivo Narrow.
- Don't let a rejected claim reach a slide — the app shows it, the deck does not.
- Don't introduce a second saturated colour: shell is the only field colour.
