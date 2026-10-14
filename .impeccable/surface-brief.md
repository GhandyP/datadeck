# Surface: app shell and deck review (operate) + the exported deck

## Direction contract

THESIS: DataDeck is a case file, not a dashboard. Every figure on screen carries
the label that says where it came from, and the category's default arrangement —
a grid of neutral cards with one blue accent — is refused.

OWN-WORLD: Exhibit blue `#20348F` for the shell band, navigation and the primary
action; file grey `#E7E8E6` ground with white panels; hairline rules `#C9CCC7`;
ink `#15181B`; manila `#C98B2E` reserved exclusively for provenance and spent
nowhere else; stamp red `#B3261E` for a rejected claim; archive green `#1E6B4F`
for a verified one. Archivo for text, Archivo Narrow uppercase and tracked for
codes and column headers. Density is high and ruled: square corners, hairline
rules, no shadows, tabular figures throughout.

STORY: The visitor pastes a table, sees it parsed as ruled rows, watches the
claims arrive each carrying its provenance code, selects a claim to see exactly
which source cells support it, watches a claim get rejected because its number is
not in the data, and downloads one file.

FIRST VIEWPORT: An exhibit band crosses the top with the product name, the
derived deck code set in Archivo Narrow, and the three steps with the active one
marked by a solid manila rule. Below it, three ruled work areas: the source table
at the left, the claim list in the centre — each claim showing its provenance code
and its status — and the deck summary with the download action at the right.

FORM: grounded candidate 6 of 7, the court-exhibit and chain-of-custody world;
seed key `f101b6bf`.

RAISED: from the Saville catalog sleeve, decode-on-demand — the provenance code
translates to the exact source cells the moment it is selected. From the boarding
pass and gate board, persistent change state — a figure that moved against the
previous period stays marked until it is looked at. From the vertical media feed,
one item owns the frame — in the deck viewer, a slide owns the whole viewport and
the chrome retreats.

FINISH: unreviewed and undocumented is unfinished; this build ends with the
finish review, the verdict, and DESIGN.md.

## Scope

Two developer-facing surfaces share one system: the local app shell (Operate) and
the self-contained deck it exports. The deck is a document, so it wears the same
type, palette and density with a wider measure; it ships no navigation chrome
beyond the slide controls that belong to it.

## Undecided

None. The world, the palette, the type and the signature move are settled; the
implementation follows the frozen contract in `.impeccable/direction.md`.
