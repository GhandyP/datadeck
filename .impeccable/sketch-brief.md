# Sketch brief — DataDeck decision round (code-led)

You write ONE first-viewport HTML sketch per assigned card. These are decision
aids the user compares side by side, then throw away. Nothing here ships.

## Hard rules

- One self-contained `.html` file per card. Inline CSS only. No JavaScript. No
  images. No fonts fetched except Google Fonts.
- Exactly the first viewport: `1440 x 900`, fixed height, `overflow: hidden`.
  `html, body { margin:0; height:900px; }`.
- Under 200 lines. Do not add a page header, a caption, or a label naming the
  direction — the decision page supplies all framing.
- Real product content, identical across every card (below). Never write
  "Lorem ipsum", "placeholder", or grey skeleton tiles.
- These are WORKING SCREENS. The world lends only four things: type, palette,
  density, and one signature move. It never supplies the page's organisation
  and never becomes a costume: every control is a standard web control the
  visitor already knows (button, input, `<table>` with a header row, tab,
  checkbox), navigation reads as navigation, and a table is a table.
- A screen an engineer could not build from standard web components is a
  picture of a tool, not the tool. Do not draw a costume of the world's own
  instrument.

## The product content — identical in every sketch

Header band: product name `DataDeck`, and the deck's derived code `EX-7F3A21`.
Navigation: three steps, `1 Source`, `2 Plan`, `3 Deck`. Step 2 is active.

Deck identity: title `Q3 2026 revenue review`, audience `Board`.

Parsed source table (this is what the user pasted):

| Region        | Revenue   | Change |
|---------------|-----------|--------|
| North America | 4,820,000 | +6.4%  |
| Europe        | 3,140,000 | −2.1%  |
| LATAM         | 1,275,000 | +11.8% |
| APAC          | 2,460,000 | +0.4%  |

The claim list, with each claim carrying its provenance code:

- `C1  Total revenue reached 11,695,000.` — provenance `Σ Revenue` — verified
- `C2  LATAM grew fastest at +11.8%.` — provenance `LATAM · Change` — verified
- `C3  Europe declined 7.3%.` — provenance `Europe · Change` — **REJECTED**:
  the number 7.3 does not appear in the source (`−2.1%` does). This is the
  alert state and it must be visible in every sketch.

Deck preview: 5 slides, and the primary action `Download deck.html`.

## What every sketch must contain

The band with the product name and the code; the three-step navigation with
step 2 active; the source table above; the claim list carrying the provenance
codes with `C3` visibly in its alert state; the deck summary with the download
action. What changes per card is type, palette, density, and the signature move.

## Per-card specification

### assigned.html — Exhibit & Chain of Custody

- Signature move: selecting a claim lights its source rows. Show `C2` selected:
  its row in the claim list is the active one and the `LATAM` source row is lit
  in manila.
- Palette: shell `#20348F`, ground `#E7E8E6`, ink `#15181B`, manila `#C98B2E`
  (provenance only, never elsewhere), alert `#B3261E`, verified `#1E6B4F`.
- Type: `Archivo` for everything; `Archivo Narrow` for the code, the step
  numbers, and column headers, uppercase and tracked. Tabular figures
  (`font-variant-numeric: tabular-nums`).
- Density: high and ruled. Hairline rules between rows, a stamped block for the
  code, no rounded corners, no shadows.

### model-pick.html — Investor Annual Report

- Palette: ink `#14213D`, paper `#F5F2EC`, gold rule `#D4A017`, teal `#2E7D6F`
  for verified, alert `#A4262C`.
- Type: `Source Serif 4` for the deck title and the claim sentences; `Archivo`
  for controls and table.
- Density: generous. Large title, wide margins, one hero figure set large, the
  source table on a ruled baseline.
- Signature move: the deck title and the hero figure are set at statement
  scale — the number is the composition.

### challenger-orienteering.html — Orienteering Course

- Palette: `#FFFFFF` ground, `#C9D96B` open land, `#8B5E34` contour brown,
  `#2C7FB8` marsh blue, `#7A2EBF` course purple reserved for the narrative
  layer, `#C0392B` alert.
- Type: `Barlow Condensed` for control numerals and headers; `Barlow` for text.
- Density: dense and layered.
- Signature move: the source table is the terrain and it never changes; the
  claims are numbered purple controls riding above it. On the left, a ranked
  symbol legend orders the rows by magnitude.

### challenger-wayfinding.html — Terminal Wayfinding

- Palette: panel yellow `#F2C200`, black `#111111`, white `#FFFFFF`, alert red
  `#C0392B` used only for the rejected claim.
- Type: `Fira Sans` with `Fira Sans Condensed` for headers.
- Density: very high contrast, one message per band, monumental numerals.
- Signature move: navigation is a chain of yellow sign bands, one message per
  band, with the step number set large; the yellow never touches data ink.

### challenger-splitflap.html — Split-Flap Board

- Palette: flap black `#0E0F11`, white paint `#F2F2EF`, steel `#9AA0A6`, amber
  `#D98A00` reserved for a changed value, alert `#C0392B`.
- Type: `Saira Condensed` for the board rows, uppercase; `Archivo` for the rest.
- Density: strict ruled rows and fixed columns; columns never move.
- Signature move: the deck's slide index is a board of ruled rows, one row per
  slide, each row's status restyling in place; the source table keeps the same
  ruled rhythm.

### canon.html — The category standard

- Palette: white `#FFFFFF`, ink `#111827`, blue `#2563EB`, alert `#DC2626`.
- Type: `Public Sans` throughout; rounded 8px corners on cards; one subtle
  shadow on the primary card.
- Density: comfortable, card-based, left rail navigation.
- Signature move: none by design — this is the arrangement a visitor expects,
  executed at full craft. Do not add a fifth thing.
