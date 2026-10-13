# DataDeck — portfolio-ready data-to-presentations web app
Status: Implementation authorized; repository rename is the first work unit.
Target: public GitHub repository `GhandyP/datadeck`; project content has not been pushed.

## Specs

- **S1:** “1 el reconocimiento y explorar , crear el repo de github con git y gh cli , crear un plan odd para crear desde cero en este repo el proyecto completo y lista para mi portafolio en github”
- **S2:** “app web + htlm de la mejor calidad”
- **S3:** “Solo en esta carpeta”
- **S4:** “Público”
- **S5:** “puedespiedes cambiarle el nombre a algo mejor y luego implementar el plan”

## Tasks

- **T0 — S5 — Rename the public GitHub repository to DataDeck (`datadeck`) and verify the local `origin`. Route: parent inline, before T1. Proposed commit: `docs: rebrand project as DataDeck`.
- **T1 — S1-S2 — Define the web-app contract and architecture:** settle user input, generated-plan review, HTML deck packaging, and whether the existing Go/Gemini prototype is reusable. Route: parent inline, before implementation. Proposed commit: `docs(spec): define web-to-HTML presentation contract` (not authorized or created in this phase).
- **T2 — S1-S2 — Build the web experience:** implement the agreed responsive interface for supplying source data, requesting a presentation, and inspecting generation/errors. Route: inline after T1. Proposed commit: `feat(web): add presentation planning interface` (not authorized or created in this phase).
- **T3 — S1-S2 — Implement grounded slide planning:** produce structured slide content tied to supplied evidence and validate the agreed output contract. Route: inline after T1-T2. Proposed commit: `feat(planning): generate validated evidence-linked slide plans` (not authorized or created in this phase).
- **T4 — S1-S2 — Render high-quality HTML presentations:** create the agreed HTML deck output and its viewing/export flow, with visual quality and accessibility checks defined in T1. Route: inline after T3. Proposed commit: `feat(rendering): export polished HTML presentations` (not authorized or created in this phase).
- **T5 — S1-S2 — Verify the end-to-end product:** test the public user flow, invalid/incomplete input, evidence validation, HTML output, and relevant accessibility/responsive behavior against the contract. Route: inline after T2-T4. Proposed commit: `test: cover web-to-HTML presentation flow` (not authorized or created in this phase).
- **T6 — S1-S2 — Prepare the GitHub portfolio presentation:** document setup, architecture, usage, limitations, and provide representative product visuals after implementation is available. Route: parent inline after T5. Proposed commit: `docs: prepare project for portfolio review` (not authorized or created in this phase).

## Log

- **L1 — Original request:** “1 el reconocimiento y explorar , crear el repo de github con git y gh cli , crear un plan odd para crear desde cero en este repo el proyecto completo y lista para mi portafolio en github”
- **L2 — User clarification:** “app web + htlm de la mejor calidad”
- **L3 — User selections:** “Solo en esta carpeta”; “Público”.
- **L4 — Reconnaissance:** the target folder initially contained only `README.md` and `main.go`, with no local `.git`. Git had instead resolved to `/home/ghandy`, whose index already contained unrelated staged personal files. A new repository was initialized only in the target folder; the parent repository and its staging were left unchanged.
- **L5 — Existing project baseline:** `README.md` describes a text-only slide plan and calls out structured slide data, claim/evidence validation, and human review as missing or future work. `main.go` invokes a Go ADK helper and returns untyped prose; it imports `github.com/GhandyP/adk-go-enterprise-cookbook/internal/adkhelper`. There is no `go.mod` or web frontend in this directory. T1 must verify what can be reused before choosing implementation details.
- **L6 — GitHub setup:** `GhandyP/15-data-to-presentations` was created as a public repository and configured as `origin`. The remote has no refs; no project files were pushed and no commits were made. No deployment was requested.
- **L7 — Scope boundary at planning time:** the initial request authorized reconnaissance, isolated repository setup, and this roadmap only; implementation and publishing project content were future work at that point.
- **L8 — User authorization, verbatim:** “puedespiedes cambiarle el nombre a algo mejor y luego implementar el plan”
- **L9 — Naming decision:** chose DataDeck (`datadeck`) as a short, descriptive project name; the GitHub account namespace lookup showed that slug was available. This selects the name on the user's behalf as requested.
