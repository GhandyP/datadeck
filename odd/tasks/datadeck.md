# DataDeck — portfolio-ready data-to-presentations web app
Status: complete. Branch `feature/datadeck`; nothing pushed.
Target: public GitHub repository `GhandyP/datadeck`.

## Specs

- **S1:** “1 el reconocimiento y explorar , crear el repo de github con git y gh cli , crear un plan odd para crear desde cero en este repo el proyecto completo y lista para mi portafolio en github”
- **S2:** “app web + htlm de la mejor calidad”
- **S3:** “Solo en esta carpeta”
- **S4:** “Público”
- **S5:** “puedespiedes cambiarle el nombre a algo mejor y luego implementar el plan”
- **S6 (answered):** stack “Go + HTML/CSS/JS”; generation “IA obligatoria”; execution “Local, un binario”.

## Tasks

- **T0 — S5 — Rename the public repository to DataDeck (`datadeck`).** Done. Evidence: `gh repo rename` produced `GhandyP/datadeck`; `origin` updated; remote still has no refs.
- **T1 — S1-S2 — Define the web-app contract and architecture.** Done. Evidence: `PRODUCT.md`, `.impeccable/surface-brief.md` (direction contract), `.impeccable/direction.md` (frozen interface contract), commit `8727b90`.
- **T2 — S1-S2 — Build the web experience.** Done. Evidence: `9cb48d8`; `main.go` serves the embedded app and the HTTP API, `internal/agent` is the only place a model is called, `web/` is the app shell. Proven against a running server: `/` 200 `text/html`, `/tokens.css` 200 `text/css`, `/fonts/…` 200 `font/woff2`, generation without a key 502 `missing_api_key`, unusable input 400 `input`.
- **T3 — S1-S2 — Implement grounded slide planning.** Done. Evidence: `f4ee77c` (`internal/data`) and `e5de00d` (`internal/plan`), both green under `go test ./...`.
- **T4 — S1-S2 — Render high-quality HTML presentations.** Done. Evidence: `a4eee82`; `internal/deck` composes one self-contained file. Tests assert no external host, marker removal, escaping, and that a rejected claim never reaches a slide.
- **T5 — S1-S2 — Verify the end-to-end product.** Done. Evidence: `go build`, `go vet` and `go test ./... -count=1` all exit 0; `main_test.go` covers the whole HTTP surface with a fake planner (no credentials, no network); the rendered deck returns zero findings from the design detector; captures at 1440 and 390 wide were inspected, and the five finish-review fixes were re-verified in them.
- **T6 — S1-S2 — Prepare the GitHub portfolio presentation.** Done. README with the architecture and the guarantee, `DESIGN.md` with the token ramp and the named rules, `.impeccable/design.json` sidecar.

## Log

- **L1 — Original request:** “1 el reconocimiento y explorar , crear el repo de github con git y gh cli , crear un plan odd para crear desde cero en este repo el proyecto completo y lista para mi portafolio en github”
- **L2 — Clarification:** “app web + htlm de la mejor calidad”
- **L3 — Selections:** “Solo en esta carpeta”; “Público”.
- **L4 — Reconnaissance:** the folder held only `README.md` and `main.go`, with no local `.git`; Git resolved to `/home/ghandy`, whose index held unrelated staged personal files. A repository was initialized only in the project folder and the parent index was left untouched.
- **L5 — Baseline:** the old `main.go` imports `github.com/GhandyP/adk-go-enterprise-cookbook/internal/adkhelper`, an internal package of a repository that does not exist on GitHub, so the prototype never compiled. It was replaced rather than repaired, and it is not part of any commit.
- **L6 — GitHub setup:** public `GhandyP/datadeck` created and configured as `origin`; no refs, no commits pushed, no deployment.
- **L7 — Scope boundary at planning time:** the first request authorized reconnaissance, isolated repository setup, and a roadmap only.
- **L8 — User authorization:** “puedespiedes cambiarle el nombre a algo mejor y luego implementar el plan”
- **L9 — Naming decision:** DataDeck (`datadeck`), chosen on the user's behalf as requested; the slug was verified free before the rename.
- **L10 — Product decisions, verbatim answers:** stack “Go + HTML/CSS/JS (recomendado)”; generation “IA obligatoria”; execution “Local, un binario (recomendado)”. Consequence recorded honestly: without `GOOGLE_API_KEY` the product cannot generate, by explicit product decision. The server therefore starts anyway and the page states exactly what is missing instead of failing silently.
- **L11 — Visual world:** the concept-seed roll assigned grounded candidate 6 of 7, the court-exhibit and chain-of-custody world. The decision page closed unanswered (exit 4), the structured question tool was used once as the documented fallback, and the user chose **Exhibit & Chain of Custody**. Seed key `f101b6bf`.
- **L12 — Design-round route:** the six first-viewport sketches were dropped, not skipped: they existed only to feed the decision page, and that page was closed. Recorded as dropped so no task is silently abandoned.
- **L13 — Incident, subagents unavailable:** every writer delegation was refused with “Writer session Git authority changed before admission”, and `session_worktree_register` confirmed the cause: “Session Git identity changed; cannot register a worktree”. The session's Git identity is bound to the `/home/ghandy` repository, while this project is a separate repository created inside it. The writer route is reported unavailable rather than retried further; all implementation was done inline.
- **L14 — Defect found and fixed in my own work:** the first commit staged `main.go` and a throwaway payload while its message claimed the prototype was excluded. The commit was amended (`8727b90`) so the message matches the diff and both files are out of it.
- **L15 — Verification evidence so far:** `go test ./internal/...` green for `internal/data`, `internal/plan`, `internal/deck`; RED was observed before each implementation (undefined symbols, then passing).
- **L16 — Defect found by the review fixture:** generating a real deck showed two rejected claims where the draft held one. The figure extractor dropped a leading minus sign, so a claim that correctly wrote `-2.1%` was rejected with the confusing reason “the claim states 2.1% … it says -2.1%”. Tests were added first (`TestValidateAcceptsAClaimThatKeepsTheSignOfANegativeValue`), RED was observed with exactly that message, and the extractor now treats the sign as part of the figure while still treating `2026-Q3` and `Q3-2026` as labels. The fixture now reports the one real rejection.
- **L17 — Design-system defect found by the finish detector:** the exhibit band and the active step spent `--manila` on decoration, contradicting this project's own rule that manila means provenance and nothing else. Both were corrected to `--shell-deep` and `--shell-ink`; the detector now returns no findings.
- **L18 — Finish review (degraded, in-thread, disclosed):** the shipped finish reviewer could not be spawned, so the review was run in this thread against the craft floor. It found five material items, all fixed: (1) the provenance mark set manila text on its own wash at about 2.3:1, below AA — a `--manila-ink` token was added for text while the base keeps the rule and the wash fills; (2) the deck carried a “SLIDE 01” kicker above its heading, which the craft floor bans outright — deleted, the position stays in the slide's `aria-label` and the counter; (3) the source-cell highlight carried a 3px inset edge, the same tell as the banned side-tab — removed, the wash marks the cells on its own; (4) browser surfaces were unthemed — `::selection` and `caret-color` now come from the palette on both surfaces; (5) the deck fell back to the platform sans, which the craft floor calls a failure rather than a fallback — Archivo 400/600 and Archivo Narrow 600 are now embedded and inlined as data URIs, so the deck carries its own voice and still reaches no host. The app now serves the same fonts from its own origin, so it no longer depends on a font host either.
- **L19 — Verification route unavailable:** `gentle-ai-verify` was attempted for the risk-item-5 verifier and failed with “could not register launched worktree: Session Git identity changed”, the same cause as L13. Both the writer and the verifier routes are reported unavailable for this workspace; no further retries were made.
