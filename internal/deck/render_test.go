package deck

import (
	"errors"
	"strings"
	"testing"

	"github.com/GhandyP/datadeck/internal/data"
	"github.com/GhandyP/datadeck/internal/plan"
)

func samplePlan() plan.Plan {
	return plan.Plan{
		Code:     "EX-7F3A21",
		Title:    "Q3 2026 revenue review",
		Audience: "Board",
		Status:   plan.StatusDraft,
		Facts: []data.Fact{
			{ID: "F-05", Label: "Revenue · total", Value: "11,695,000", Row: data.WholeColumn, Column: 1},
			{ID: "F-07", Label: "Europe · Change", Value: "-2.1%", Row: 1, Column: 2},
		},
		Slides: []plan.Slide{{
			Number: 1,
			Title:  "Revenue grew in three of four regions",
			Notes:  "Say the caveat before the number.",
			Claims: []plan.Claim{
				{ID: "C1", Text: "Total revenue reached 11,695,000.", EvidenceIDs: []string{"F-05"}, Status: plan.StatusVerified},
				{ID: "C2", Text: "Europe declined 7.3%.", EvidenceIDs: []string{"F-07"}, Status: plan.StatusRejected, Reason: "the cited data says -2.1%"},
			},
		}},
		RejectedCount: 1,
	}
}

func TestRenderProducesASelfContainedDeck(t *testing.T) {
	out, err := Render(samplePlan())
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	deck := string(out)

	wanted := []string{
		"<!doctype html",
		"EX-7F3A21",
		"Q3 2026 revenue review",
		"Total revenue reached 11,695,000.",
		"Revenue grew in three of four regions",
		"Say the caveat before the number.",
		"window.__DATADECK__",
		"--manila",
	}
	for _, want := range wanted {
		if !strings.Contains(deck, want) {
			t.Errorf("the deck does not contain %q", want)
		}
	}

	// The shell's markers are the renderer's contract with the assets: none may
	// survive into the output.
	for _, marker := range []string{styleMarker, dataMarker, scriptMarker} {
		if strings.Contains(deck, marker) {
			t.Errorf("the deck still contains the marker %q", marker)
		}
	}

	// A rejected claim is shown to the user in the app and never on a slide.
	if strings.Contains(deck, "Europe declined 7.3%.") {
		t.Error("the deck contains a rejected claim")
	}

	// The file is opened by double-clicking it, so it must reach nothing.
	for _, external := range []string{"http://", "https://", "//fonts."} {
		if strings.Contains(deck, external) {
			t.Errorf("the deck reaches an external host: %q", external)
		}
	}
}

func TestRenderCannotBeClosedByItsOwnContent(t *testing.T) {
	p := samplePlan()
	p.Title = `Bad </script><script>alert(1)</script> title`

	out, err := Render(p)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	deck := string(out)

	if strings.Contains(deck, `Bad </script>`) {
		t.Error("the title closed the script element it is written into")
	}
	if !strings.Contains(deck, `\u003c/script\u003e`) {
		t.Error("the title was not escaped into the payload")
	}
}

func TestRenderRefusesAPlanWithNoVerifiedClaims(t *testing.T) {
	p := samplePlan()
	p.Slides[0].Claims = []plan.Claim{{
		ID: "C1", Text: "Europe declined 7.3%.", EvidenceIDs: []string{"F-07"},
		Status: plan.StatusRejected, Reason: "the cited data says -2.1%",
	}}

	_, err := Render(p)
	if !errors.Is(err, ErrNoVerifiedClaims) {
		t.Fatalf("error = %v, want %v", err, ErrNoVerifiedClaims)
	}
}

func TestShellDeclaresEachMarkerExactlyOnce(t *testing.T) {
	shell, err := assets.ReadFile("assets/deck.html")
	if err != nil {
		t.Fatalf("read deck shell: %v", err)
	}
	for _, marker := range []string{styleMarker, dataMarker, scriptMarker} {
		if got := strings.Count(string(shell), marker); got != 1 {
			t.Errorf("marker %q appears %d times, want 1", marker, got)
		}
	}
}
