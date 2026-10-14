// Command datadeck turns a table of numbers into a finished HTML slide deck
// whose every figure traces back to the data the user supplied.
//
// It serves one page on localhost: the visitor pastes a table, names the deck,
// and gets a plan whose every claim is checked against the input plus one
// self-contained HTML file to download. Generation requires GOOGLE_API_KEY;
// without it the server still starts and the page explains exactly what is
// missing rather than pretending to work.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GhandyP/datadeck/internal/agent"
	"github.com/GhandyP/datadeck/internal/data"
	"github.com/GhandyP/datadeck/internal/deck"
	"github.com/GhandyP/datadeck/internal/plan"
	"github.com/GhandyP/datadeck/internal/theme"
)

//go:embed web
var webFiles embed.FS

// maxBodyBytes bounds one pasted table. Anything larger is refused instead of
// being read into memory.
const maxBodyBytes = 2 << 20

// modelTimeout bounds one generation, so a stalled model cannot hold a request
// open indefinitely.
const modelTimeout = 120 * time.Second

// Planner is the one thing this server needs from the model layer. Tests
// substitute a fake, so the whole HTTP surface is exercised without credentials.
type Planner interface {
	Draft(ctx context.Context, req agent.Request) (plan.Draft, error)
}

type server struct {
	planner Planner
	model   string
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "address to listen on")
	modelName := flag.String("model", "", "Gemini model name (default: ADK_MODEL, then "+agent.DefaultModel+")")
	flag.Parse()

	name := firstNonEmpty(*modelName, os.Getenv(agent.ModelEnv), agent.DefaultModel)

	var planner Planner
	created, err := agent.NewAgent(context.Background(), agent.Config{ModelName: name})
	switch {
	case err == nil:
		planner = created
	case errors.Is(err, agent.ErrMissingAPIKey):
		log.Printf("warning: %v", err)
		log.Printf("the server starts anyway; the page explains what is missing instead of building a deck")
	default:
		log.Fatalf("model: %v", err)
	}

	srv := &server{planner: planner, model: name}
	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           srv.routes(),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      modelTimeout + 30*time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("DataDeck listening on http://%s (model %s)", *addr, name)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve: %v", err)
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	static, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatalf("read the embedded app: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(static)))

	mux.HandleFunc("/tokens.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		io.WriteString(w, theme.FontCSS(false)+theme.Tokens())
	})
	// The fonts are served from this origin rather than a font host, so the app
	// renders in its own voice and reaches nothing external.
	for _, name := range theme.FontNames() {
		mux.HandleFunc("/fonts/"+name, func(w http.ResponseWriter, r *http.Request) {
			data, ok := theme.Font(name)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "font/woff2")
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			w.Write(data)
		})
	}
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/generate", s.generate)
	return mux
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method", "this endpoint answers GET")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":           "ok",
		"model":            s.model,
		"apiKeyConfigured": s.planner != nil,
	})
}

type generateRequest struct {
	Title    string `json:"title"`
	Audience string `json:"audience"`
	Data     string `json:"data"`
}

type generateResponse struct {
	Code          string       `json:"code"`
	Title         string       `json:"title"`
	Audience      string       `json:"audience"`
	Status        string       `json:"status"`
	Source        data.Table   `json:"source"`
	Facts         []data.Fact  `json:"facts"`
	Slides        []plan.Slide `json:"slides"`
	RejectedCount int          `json:"rejectedCount"`
	DeckHTML      string       `json:"deckHTML"`
}

func (s *server) generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method", "this endpoint answers POST")
		return
	}

	var request generateRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "input", "the request body is not the JSON this endpoint expects")
		return
	}

	title := strings.TrimSpace(request.Title)
	audience := strings.TrimSpace(request.Audience)
	if audience == "" {
		audience = "General"
	}
	if title == "" {
		writeError(w, http.StatusBadRequest, "input", "give the deck a title first")
		return
	}

	table, err := data.Parse(request.Data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "input", err.Error())
		return
	}
	facts, err := data.ComputeFacts(table)
	if err != nil {
		writeError(w, http.StatusBadRequest, "input", err.Error())
		return
	}

	if s.planner == nil {
		writeError(w, http.StatusBadGateway, "missing_api_key",
			"GOOGLE_API_KEY is not set, so no deck can be built. Set it in the environment you started the server from, then reload this page.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), modelTimeout)
	defer cancel()

	draft, err := s.planner.Draft(ctx, agent.Request{
		Title:    title,
		Audience: audience,
		Table:    table,
		Facts:    facts,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, "model", fmt.Sprintf("the model did not produce a usable answer: %v", err))
		return
	}

	validated, err := plan.Validate(title, audience, facts, draft)
	if err != nil {
		writeError(w, http.StatusBadGateway, "model", fmt.Sprintf("the model did not produce a usable answer: %v", err))
		return
	}
	validated.Code = plan.Code(title, request.Data)

	// A deck whose every claim was rejected is still worth showing: the user
	// needs the reasons. It simply has nothing to download.
	rendered, err := deck.Render(validated)
	if err != nil && !errors.Is(err, deck.ErrNoVerifiedClaims) {
		writeError(w, http.StatusInternalServerError, "render", fmt.Sprintf("the deck could not be rendered: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, generateResponse{
		Code:          validated.Code,
		Title:         validated.Title,
		Audience:      validated.Audience,
		Status:        validated.Status,
		Source:        table,
		Facts:         validated.Facts,
		Slides:        validated.Slides,
		RejectedCount: validated.RejectedCount,
		DeckHTML:      string(rendered),
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, kind, message string) {
	writeJSON(w, status, map[string]string{"error": message, "kind": kind})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
