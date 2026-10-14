// Package plan holds the deck's contract and the rule that makes it honest: a
// slide may only state a figure the supplied data contains.
//
// The model writes language. The program owns every value: the facts a plan
// carries are the ones computed before the call, and a claim that cites a value
// its evidence does not hold is rejected with the reason instead of being
// printed on a slide.
package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/GhandyP/datadeck/internal/data"
)

// StatusDraft is the only status a deck can have. Presenting it to an audience
// needs a human to approve its claims, which this program cannot supply.
const StatusDraft = "draft"

// ErrNoSlides reports an answer that carried no slides at all.
var ErrNoSlides = errors.New("the model returned no slides")

// ClaimStatus is the verdict on one claim.
type ClaimStatus string

const (
	// StatusVerified means every figure in the claim came from the data it cites.
	StatusVerified ClaimStatus = "verified"
	// StatusRejected means the claim is shown to the user and kept out of the deck.
	StatusRejected ClaimStatus = "rejected"
)

// Claim is one statement on a slide with the source values it rests on.
type Claim struct {
	ID          string      `json:"id"`
	Text        string      `json:"text"`
	EvidenceIDs []string    `json:"evidenceIds"`
	Status      ClaimStatus `json:"status"`
	Reason      string      `json:"reason,omitempty"`
}

// Slide is one page of the deck.
type Slide struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	Notes  string  `json:"notes"`
	Claims []Claim `json:"claims"`
}

// Plan is the whole deliverable: the deck's identity, its slides, the facts it
// may state, and how many claims were kept out of it.
type Plan struct {
	Code          string      `json:"code"`
	Title         string      `json:"title"`
	Audience      string      `json:"audience"`
	Status        string      `json:"status"`
	Slides        []Slide     `json:"slides"`
	Facts         []data.Fact `json:"facts"`
	RejectedCount int         `json:"rejectedCount"`
}

// Draft is the model's answer before any of it is trusted.
type Draft struct {
	Slides []DraftSlide `json:"slides"`
}

// DraftSlide is one slide as the model wrote it.
type DraftSlide struct {
	Title  string       `json:"title"`
	Notes  string       `json:"notes"`
	Claims []DraftClaim `json:"claims"`
}

// DraftClaim is one claim as the model wrote it: text and the ids it cites.
type DraftClaim struct {
	Text        string   `json:"text"`
	EvidenceIDs []string `json:"evidenceIds"`
}

// Code derives the deck's identity from the request that produced it. The same
// title and data always give the same code, and a different table never does.
func Code(title, raw string) string {
	sum := sha256.Sum256([]byte(title + "\x00" + raw))
	return "EX-" + strings.ToUpper(hex.EncodeToString(sum[:3]))
}

// VerifiedClaims returns the claims that may be shown, in deck order.
func (p Plan) VerifiedClaims() []Claim {
	var out []Claim
	for _, slide := range p.Slides {
		for _, claim := range slide.Claims {
			if claim.Status == StatusVerified {
				out = append(out, claim)
			}
		}
	}
	return out
}
