package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GhandyP/datadeck/internal/agent"
	"github.com/GhandyP/datadeck/internal/plan"
)

const kpiData = `Region,Revenue,Change
North America,4820000,+6.4%
Europe,3140000,-2.1%
LATAM,1275000,+11.8%
APAC,2460000,+0.4%
`

// fakePlanner stands in for the model, so the whole HTTP surface is exercised
// without credentials or a network call.
type fakePlanner struct {
	draft plan.Draft
	err   error
	got   agent.Request
	calls int
}

func (f *fakePlanner) Draft(_ context.Context, req agent.Request) (plan.Draft, error) {
	f.calls++
	f.got = req
	return f.draft, f.err
}

func draftWith(claims ...plan.DraftClaim) plan.Draft {
	return plan.Draft{Slides: []plan.DraftSlide{{
		Title:  "Revenue grew in three of four regions",
		Notes:  "Say the caveat before the number.",
		Claims: claims,
	}}}
}

func post(t *testing.T, handler http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decode[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var body T
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return body
}

func TestGenerateReturnsAValidatedDeck(t *testing.T) {
	planner := &fakePlanner{draft: draftWith(
		plan.DraftClaim{Text: "Total revenue reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
		plan.DraftClaim{Text: "Europe declined 7.3%.", EvidenceIDs: []string{"F-07"}},
	)}
	handler := (&server{planner: planner, model: "test-model"}).routes()

	recorder := post(t, handler, `{"title":"Q3 2026 revenue review","audience":"Board","data":`+quote(kpiData)+`}`)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}

	body := decode[generateResponse](t, recorder)
	if body.Code == "" || !strings.HasPrefix(body.Code, "EX-") {
		t.Errorf("code = %q, want an EX- code", body.Code)
	}
	if body.Status != plan.StatusDraft {
		t.Errorf("status = %q, want %q", body.Status, plan.StatusDraft)
	}
	if body.RejectedCount != 1 {
		t.Errorf("rejectedCount = %d, want 1", body.RejectedCount)
	}
	if got, want := len(body.Source.Rows), 4; got != want {
		t.Errorf("source rows = %d, want %d", got, want)
	}
	if len(body.Facts) == 0 {
		t.Fatal("the response carries no facts, so nothing could be traced")
	}

	// The claim that cannot be traced is reported with its reason and kept out
	// of the deck; the one that can is on the deck.
	if !strings.Contains(body.Slides[0].Claims[1].Reason, "7.3") {
		t.Errorf("rejection reason = %q, want it to name 7.3", body.Slides[0].Claims[1].Reason)
	}
	if !strings.Contains(body.DeckHTML, "Total revenue reached 11,695,000.") {
		t.Error("the deck does not carry the verified claim")
	}
	if strings.Contains(body.DeckHTML, "Europe declined 7.3%.") {
		t.Error("the deck carries a rejected claim")
	}
}

func TestGenerateHandsThePlannerTheParsedTableAndFacts(t *testing.T) {
	planner := &fakePlanner{draft: draftWith(
		plan.DraftClaim{Text: "Total revenue reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
	)}
	handler := (&server{planner: planner, model: "test-model"}).routes()

	post(t, handler, `{"title":"Deck","data":`+quote(kpiData)+`}`)

	if planner.calls != 1 {
		t.Fatalf("planner calls = %d, want 1", planner.calls)
	}
	if got, want := planner.got.Table.Columns[1], "Revenue"; got != want {
		t.Errorf("planner column = %q, want %q", got, want)
	}
	if len(planner.got.Facts) == 0 {
		t.Error("the planner received no facts, so it could not cite anything")
	}
	if planner.got.Audience != "General" {
		t.Errorf("audience = %q, want the default General", planner.got.Audience)
	}
}

func TestGenerateStillShowsARejectedDeck(t *testing.T) {
	planner := &fakePlanner{draft: draftWith(
		plan.DraftClaim{Text: "Europe declined 7.3%.", EvidenceIDs: []string{"F-07"}},
	)}
	handler := (&server{planner: planner, model: "test-model"}).routes()

	recorder := post(t, handler, `{"title":"Deck","data":`+quote(kpiData)+`}`)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 so the reasons reach the user", recorder.Code)
	}
	body := decode[generateResponse](t, recorder)
	if body.DeckHTML != "" {
		t.Error("a deck with no verified claim must not be offered for download")
	}
	if !strings.Contains(body.Slides[0].Claims[0].Reason, "7.3") {
		t.Errorf("reason = %q, want it to name 7.3", body.Slides[0].Claims[0].Reason)
	}
}

func TestGenerateRefusals(t *testing.T) {
	planner := &fakePlanner{draft: draftWith(
		plan.DraftClaim{Text: "Total revenue reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
	)}
	handler := (&server{planner: planner, model: "test-model"}).routes()

	cases := []struct {
		name     string
		body     string
		wantKind string
		want     string
	}{
		{
			name:     "no title",
			body:     `{"title":"  ","data":` + quote(kpiData) + `}`,
			wantKind: "input",
			want:     "title",
		},
		{
			name:     "no data",
			body:     `{"title":"Deck","data":""}`,
			wantKind: "input",
			want:     "no data was supplied",
		},
		{
			name:     "ragged row",
			body:     `{"title":"Deck","data":` + quote("Region,Revenue\nNorth,1\nEurope\n") + `}`,
			wantKind: "input",
			want:     "row 2",
		},
		{
			name:     "no numbers to present",
			body:     `{"title":"Deck","data":` + quote("Region,Note\nNorth,ok\nEurope,ok\n") + `}`,
			wantKind: "input",
			want:     "numbers",
		},
		{
			name:     "unknown field",
			body:     `{"title":"Deck","data":` + quote(kpiData) + `,"extra":true}`,
			wantKind: "input",
			want:     "JSON",
		},
		{
			name:     "not json",
			body:     `not json at all`,
			wantKind: "input",
			want:     "JSON",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := post(t, handler, tc.body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", recorder.Code, recorder.Body.String())
			}
			body := decode[map[string]string](t, recorder)
			if body["kind"] != tc.wantKind {
				t.Errorf("kind = %q, want %q", body["kind"], tc.wantKind)
			}
			if !strings.Contains(body["error"], tc.want) {
				t.Errorf("error %q does not mention %q", body["error"], tc.want)
			}
		})
	}
}

func TestGenerateReportsWhatIsMissingOrBroken(t *testing.T) {
	handler := (&server{planner: nil, model: "test-model"}).routes()
	recorder := post(t, handler, `{"title":"Deck","data":`+quote(kpiData)+`}`)
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", recorder.Code)
	}
	body := decode[map[string]string](t, recorder)
	if body["kind"] != "missing_api_key" {
		t.Errorf("kind = %q, want missing_api_key", body["kind"])
	}
	if !strings.Contains(body["error"], agent.APIKeyEnv) {
		t.Errorf("error %q does not name %s", body["error"], agent.APIKeyEnv)
	}

	failing := &fakePlanner{err: errors.New("model unavailable")}
	handler = (&server{planner: failing, model: "test-model"}).routes()
	recorder = post(t, handler, `{"title":"Deck","data":`+quote(kpiData)+`}`)
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", recorder.Code)
	}
	body = decode[map[string]string](t, recorder)
	if body["kind"] != "model" {
		t.Errorf("kind = %q, want model", body["kind"])
	}
}

func TestGenerateRefusesAnOversizedBody(t *testing.T) {
	planner := &fakePlanner{draft: draftWith(
		plan.DraftClaim{Text: "Total revenue reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
	)}
	handler := (&server{planner: planner, model: "test-model"}).routes()

	huge := strings.Repeat("x", maxBodyBytes+1024)
	recorder := post(t, handler, `{"title":"Deck","data":"`+huge+`"}`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for a body over the limit", recorder.Code)
	}
	if planner.calls != 0 {
		t.Error("an oversized body reached the model")
	}
}

func TestEndpointsAnswerOnlyTheirMethod(t *testing.T) {
	handler := (&server{planner: nil, model: "test-model"}).routes()

	get := httptest.NewRequest(http.MethodGet, "/api/generate", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, get)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/generate = %d, want 405", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/health", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/health = %d, want 405", recorder.Code)
	}
}

func TestHealthReportsTheCredentialState(t *testing.T) {
	handler := (&server{planner: nil, model: "test-model"}).routes()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	body := decode[map[string]any](t, recorder)
	if body["apiKeyConfigured"] != false {
		t.Errorf("apiKeyConfigured = %v, want false without a planner", body["apiKeyConfigured"])
	}
	if body["model"] != "test-model" {
		t.Errorf("model = %v, want test-model", body["model"])
	}

	handler = (&server{planner: &fakePlanner{}, model: "test-model"}).routes()
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	body = decode[map[string]any](t, recorder)
	if body["apiKeyConfigured"] != true {
		t.Errorf("apiKeyConfigured = %v, want true with a planner", body["apiKeyConfigured"])
	}
}

func TestTheAppAndItsTokensAreServed(t *testing.T) {
	handler := (&server{planner: nil, model: "test-model"}).routes()

	for _, tc := range []struct{ path, want string }{
		{"/", "DataDeck"},
		{"/app.js", "api/generate"},
		{"/app.css", ".claim__prov"},
		{"/tokens.css", "--manila"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", recorder.Code)
			}
			if !strings.Contains(recorder.Body.String(), tc.want) {
				t.Errorf("%s does not contain %q", tc.path, tc.want)
			}
		})
	}
}

// quote encodes a fixture as a JSON string literal.
func quote(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}
