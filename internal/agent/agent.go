// Package agent is the one place DataDeck talks to a model.
//
// The boundary is deliberately narrow: the prompt, the output schema and the
// parse of the answer are pure functions that can be tested without
// credentials, the agent is bound to no tools, and everything that can be
// checked - the facts, the validation, the rendering - happens outside it. The
// model contributes language; the program owns every value.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"github.com/GhandyP/datadeck/internal/data"
	"github.com/GhandyP/datadeck/internal/plan"
)

const (
	// DefaultModel is the Gemini model this program documents.
	DefaultModel = "gemini-flash-latest"
	// ModelEnv overrides DefaultModel.
	ModelEnv = "ADK_MODEL"
	// APIKeyEnv is the environment variable the model client reads.
	APIKeyEnv = "GOOGLE_API_KEY"
	// AppName identifies the ADK application that owns the sessions.
	AppName   = "datadeck"
	agentName = "slide_planner"
	userID    = "datadeck-web"
)

// ErrMissingAPIKey reports a generation attempted without credentials.
var ErrMissingAPIKey = fmt.Errorf("%s is required to call the model", APIKeyEnv)

// ErrNoAgentOutput reports an agent turn that produced no usable answer.
var ErrNoAgentOutput = errors.New("the agent finished without an answer")

// Request is everything one generation is allowed to see: the table the user
// supplied and the facts already computed from it.
type Request struct {
	Title    string
	Audience string
	Table    data.Table
	Facts    []data.Fact
}

// instruction is the whole contract handed to the model. The program never asks
// the model for a number: the facts already carry them.
const instruction = `You write the slides of ONE presentation deck from a table the user supplied and the facts already computed from it.

Rules:
- Only use the supplied facts. Never call a tool, and never read a file, a dashboard, a metrics store or the network on your own.
- Never invent, compute, round, re-derive or estimate a number: a claim's text may contain a supplied fact value written exactly as the facts state it, and no other digits at all. When you need a count, write it in words.
- Period labels such as 2026-Q3 are labels, not numbers, and may be written as they appear.
- Every claim needs at least one evidenceIds entry, copied exactly from the id of a fact that supports it. A claim without evidence is rejected and kept off the deck.
- Titles carry no figures at all: put every figure in a claim.
- Write three to five slides. Each slide has a title, one to three claims, and one short sentence of speaker notes.
- A fact whose value is missing has no value: say it is missing, never assume one.
- This deck is a draft for internal review: never claim it was approved, published or presented, and never address the audience directly.
- Write short, plain sentences: no markdown, no tables, no headings inside a claim, no bullet characters.
`

// Instruction returns the agent instruction. Tests pin the fidelity rules on it.
func Instruction() string { return instruction }

// OutputSchema describes the answer the model must return. The facts and the
// deck's identity are deliberately absent: the program supplies them.
func OutputSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"slides": {
				Type:        genai.TypeArray,
				Description: "Three to five slides, in presentation order.",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"title": {Type: genai.TypeString, Description: "The slide's title. It carries no figures."},
						"notes": {Type: genai.TypeString, Description: "One short sentence of speaker guidance."},
						"claims": {
							Type:        genai.TypeArray,
							Description: "One to three statements, each citing the facts that support it.",
							Items: &genai.Schema{
								Type: genai.TypeObject,
								Properties: map[string]*genai.Schema{
									"text": {
										Type:        genai.TypeString,
										Description: "One plain sentence. Any figure in it is copied exactly from a cited fact.",
									},
									"evidenceIds": {
										Type:        genai.TypeArray,
										Description: "Ids of the facts that support this claim, copied exactly.",
										Items:       &genai.Schema{Type: genai.TypeString},
									},
								},
								Required: []string{"text", "evidenceIds"},
							},
						},
					},
					Required: []string{"title", "notes", "claims"},
				},
			},
		},
		Required: []string{"slides"},
	}
}

// Config selects the model and the credentials. Both default to the environment.
type Config struct {
	// ModelName defaults to ModelEnv and then to DefaultModel.
	ModelName string
	// APIKey defaults to APIKeyEnv.
	APIKey string
}

// Agent runs one ADK agent over one request.
type Agent struct {
	runner *runner.Runner
	runs   atomic.Uint64
}

// NewAgent builds the single-agent slide planner. It fails before any network
// call when no credentials are available, and it binds no tools, so the agent
// cannot read or write anything beyond the table and facts it is given.
func NewAgent(ctx context.Context, cfg Config) (*Agent, error) {
	modelName := firstNonEmpty(cfg.ModelName, os.Getenv(ModelEnv), DefaultModel)
	apiKey := firstNonEmpty(cfg.APIKey, os.Getenv(APIKeyEnv))
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	model, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("create model %q: %w", modelName, err)
	}

	planner, err := llmagent.New(llmagent.Config{
		Name:         agentName,
		Model:        model,
		Description:  "Writes the slides of one deck from a table the user supplied and the facts computed from it.",
		Instruction:  Instruction(),
		OutputSchema: OutputSchema(),
	})
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	run, err := runner.NewInMemory(AppName, planner)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	return &Agent{runner: run}, nil
}

// Draft asks the model for the slides of one request. The answer it returns is
// untrusted: the caller validates every claim against the facts before anything
// reaches a slide.
func (a *Agent) Draft(ctx context.Context, req Request) (plan.Draft, error) {
	// A fresh session per run keeps one deck's answer out of the next prompt.
	sessionID := fmt.Sprintf("deck-%d", a.runs.Add(1))
	prompt := genai.NewContentFromText(RenderPrompt(req), genai.RoleUser)

	var raw, text string
	for event, err := range a.runner.Run(ctx, userID, sessionID, prompt, agent.RunConfig{}) {
		if err != nil {
			return plan.Draft{}, fmt.Errorf("run agent: %w", err)
		}
		if event == nil {
			continue
		}
		// ADK stores the value parsed against OutputSchema on the terminal
		// event. Read it wherever it appears and keep the last one, so a
		// differently marked final event cannot make the answer disappear.
		if event.Output != nil {
			encoded, err := json.Marshal(event.Output)
			if err != nil {
				return plan.Draft{}, fmt.Errorf("encode agent output: %w", err)
			}
			raw = string(encoded)
		}
		if event.IsFinalResponse() {
			if answer := eventText(event); answer != "" {
				text = answer
			}
		}
	}

	if raw == "" {
		if text == "" {
			return plan.Draft{}, ErrNoAgentOutput
		}
		raw = text
	}

	var draft plan.Draft
	if err := json.Unmarshal([]byte(raw), &draft); err != nil {
		return plan.Draft{}, fmt.Errorf("decode agent answer: %w", err)
	}
	return draft, nil
}

// RenderPrompt is the whole user turn: the table, the facts, and nothing else.
func RenderPrompt(req Request) string {
	var b strings.Builder
	b.WriteString("Deck title: " + req.Title + "\n")
	b.WriteString("Audience: " + req.Audience + "\n\n")

	b.WriteString("Source table:\n")
	b.WriteString(strings.Join(req.Table.Columns, " | "))
	b.WriteString("\n")
	for _, row := range req.Table.Rows {
		b.WriteString(strings.Join(row, " | "))
		b.WriteString("\n")
	}

	b.WriteString("\nFacts (id | label | value):\n")
	for _, fact := range req.Facts {
		b.WriteString(fact.ID + " | " + fact.Label + " | " + fact.Value + "\n")
	}

	b.WriteString("\nWrite the deck. Every figure in a claim must be one of the values above, copied exactly, and the claim must cite the id it came from.")
	return b.String()
}

func eventText(event *session.Event) string {
	if event == nil || event.Content == nil {
		return ""
	}
	var b strings.Builder
	for _, part := range event.Content.Parts {
		if part != nil {
			b.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
