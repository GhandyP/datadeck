// Package deck renders a validated plan as one self-contained HTML file.
//
// The deck is a static shell plus the plan as JSON: the shell carries its own
// styles and script, the plan is embedded beside them, and the file fetches
// nothing when it is opened. A rejected claim is never rendered, because the
// deck only shows figures the supplied data supports.
package deck

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/GhandyP/datadeck/internal/data"
	"github.com/GhandyP/datadeck/internal/plan"
	"github.com/GhandyP/datadeck/internal/theme"
)

//go:embed assets/deck.html assets/deck.css assets/deck.js
var assets embed.FS

// The three markers the shell declares and this package fills.
const (
	styleMarker  = "<!--DATADECK:STYLE-->"
	dataMarker   = "<!--DATADECK:DATA-->"
	scriptMarker = "<!--DATADECK:SCRIPT-->"
)

// ErrNoVerifiedClaims reports a plan whose every claim was rejected. Rendering
// it would produce a deck with nothing in it, so it is refused instead.
var ErrNoVerifiedClaims = errors.New("no claim in this deck could be traced to the supplied data")

// ErrIncompleteShell reports a shell that no longer declares its markers.
var ErrIncompleteShell = errors.New("the deck shell does not declare its markers")

// deckData is what the exported file carries: the plan with every rejected
// claim removed, so nothing unsupported can reach a slide.
type deckData struct {
	Code          string       `json:"code"`
	Title         string       `json:"title"`
	Audience      string       `json:"audience"`
	Status        string       `json:"status"`
	Slides        []plan.Slide `json:"slides"`
	Facts         []data.Fact  `json:"facts"`
	RejectedCount int          `json:"rejectedCount"`
}

// Render composes the deck. The result is one HTML document that opens on any
// machine by double-clicking it.
func Render(p plan.Plan) ([]byte, error) {
	if len(p.VerifiedClaims()) == 0 {
		return nil, ErrNoVerifiedClaims
	}

	shell, err := assets.ReadFile("assets/deck.html")
	if err != nil {
		return nil, fmt.Errorf("read deck shell: %w", err)
	}
	stylesheet, err := assets.ReadFile("assets/deck.css")
	if err != nil {
		return nil, fmt.Errorf("read deck stylesheet: %w", err)
	}
	script, err := assets.ReadFile("assets/deck.js")
	if err != nil {
		return nil, fmt.Errorf("read deck script: %w", err)
	}

	payload := deckData{
		Code:          p.Code,
		Title:         p.Title,
		Audience:      p.Audience,
		Status:        p.Status,
		Slides:        verifiedSlides(p),
		Facts:         p.Facts,
		RejectedCount: p.RejectedCount,
	}
	// encoding/json escapes <, > and & by default, so the payload cannot close
	// the script element it is written into.
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode deck data: %w", err)
	}

	out := string(shell)
	for marker, replacement := range map[string]string{
		styleMarker:  "<style>\n" + theme.FontCSS(true) + theme.Tokens() + "\n" + string(stylesheet) + "\n</style>",
		dataMarker:   "<script>window.__DATADECK__ = " + string(encoded) + ";</script>",
		scriptMarker: "<script>\n" + string(script) + "\n</script>",
	} {
		if !strings.Contains(out, marker) {
			return nil, fmt.Errorf("%w: %s", ErrIncompleteShell, marker)
		}
		out = strings.Replace(out, marker, replacement, 1)
	}
	return []byte(out), nil
}

// verifiedSlides returns the plan's slides with only the claims that may be
// shown, which is what makes a rejected claim unable to reach a slide.
func verifiedSlides(p plan.Plan) []plan.Slide {
	slides := make([]plan.Slide, 0, len(p.Slides))
	for _, slide := range p.Slides {
		kept := slide
		kept.Claims = nil
		for _, claim := range slide.Claims {
			if claim.Status == plan.StatusVerified {
				kept.Claims = append(kept.Claims, claim)
			}
		}
		slides = append(slides, kept)
	}
	return slides
}
