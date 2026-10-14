package plan

import (
	"errors"
	"strings"
	"testing"

	"github.com/GhandyP/datadeck/internal/data"
)

// facts mirrors what ComputeFacts produces for the KPI fixture: one value per
// row plus the column's total, which is what a claim has to cite.
func facts() []data.Fact {
	return []data.Fact{
		{ID: "F-01", Label: "North America · Revenue", Value: "4,820,000", Row: 0, Column: 1},
		{ID: "F-02", Label: "Europe · Revenue", Value: "3,140,000", Row: 1, Column: 1},
		{ID: "F-03", Label: "LATAM · Revenue", Value: "1,275,000", Row: 2, Column: 1},
		{ID: "F-04", Label: "APAC · Revenue", Value: "2,460,000", Row: 3, Column: 1},
		{ID: "F-05", Label: "Revenue · total", Value: "11,695,000", Row: data.WholeColumn, Column: 1},
		{ID: "F-06", Label: "North America · Change", Value: "6.4%", Row: 0, Column: 2},
		{ID: "F-07", Label: "Europe · Change", Value: "-2.1%", Row: 1, Column: 2},
		{ID: "F-08", Label: "LATAM · Change", Value: "11.8%", Row: 2, Column: 2},
		{ID: "F-09", Label: "APAC · Change", Value: "0.4%", Row: 3, Column: 2},
	}
}

func draft(claims ...DraftClaim) Draft {
	return Draft{Slides: []DraftSlide{{
		Title:  "Revenue grew in three of four regions",
		Notes:  "One sentence of speaker guidance.",
		Claims: claims,
	}}}
}

func TestValidateKeepsAClaimThatCitesItsOwnValue(t *testing.T) {
	got, err := Validate("Q3 2026 revenue review", "Board", facts(), draft(
		DraftClaim{Text: "Total revenue reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
	))
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if got, want := len(got.Slides), 1; got != want {
		t.Fatalf("slides = %d, want %d", got, want)
	}
	claim := got.Slides[0].Claims[0]
	if claim.Status != StatusVerified {
		t.Fatalf("status = %q (reason %q), want %q", claim.Status, claim.Reason, StatusVerified)
	}
	if claim.ID != "C1" {
		t.Errorf("claim id = %q, want C1", claim.ID)
	}
	if got.RejectedCount != 0 {
		t.Errorf("rejectedCount = %d, want 0", got.RejectedCount)
	}
	if got.Status != StatusDraft {
		t.Errorf("plan status = %q, want %q", got.Status, StatusDraft)
	}
}

func TestValidateRejectsUnsupportedClaims(t *testing.T) {
	cases := []struct {
		name       string
		claim      DraftClaim
		wantReason string
	}{
		{
			name:       "cites no source value",
			claim:      DraftClaim{Text: "Revenue grew.", EvidenceIDs: nil},
			wantReason: "cites no source value",
		},
		{
			name:       "cites an unknown source",
			claim:      DraftClaim{Text: "Revenue grew.", EvidenceIDs: []string{"F-99"}},
			wantReason: "F-99",
		},
		{
			name:       "states a number no cited source contains",
			claim:      DraftClaim{Text: "Europe declined 7.3%.", EvidenceIDs: []string{"F-07"}},
			wantReason: "7.3",
		},
		{
			name:       "states a number that belongs to a different source",
			claim:      DraftClaim{Text: "Europe reached 4,820,000.", EvidenceIDs: []string{"F-07"}},
			wantReason: "4,820,000",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Validate("Q3 2026 revenue review", "Board", facts(), draft(tc.claim))
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}
			claim := got.Slides[0].Claims[0]
			if claim.Status != StatusRejected {
				t.Fatalf("status = %q, want %q", claim.Status, StatusRejected)
			}
			if !strings.Contains(claim.Reason, tc.wantReason) {
				t.Errorf("reason %q does not name %q", claim.Reason, tc.wantReason)
			}
			if got.RejectedCount != 1 {
				t.Errorf("rejectedCount = %d, want 1", got.RejectedCount)
			}
		})
	}
}

func TestValidateAcceptsCountsWrittenInWords(t *testing.T) {
	got, err := Validate("Q3 2026 revenue review", "Board", facts(), draft(
		DraftClaim{
			Text:        "Revenue grew in three of the four regions.",
			EvidenceIDs: []string{"F-05"},
		},
	))
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if claim := got.Slides[0].Claims[0]; claim.Status != StatusVerified {
		t.Fatalf("status = %q (reason %q), want verified", claim.Status, claim.Reason)
	}
}

func TestValidateAcceptsAClaimThatKeepsTheSignOfANegativeValue(t *testing.T) {
	got, err := Validate("Deck", "Board", facts(), draft(
		DraftClaim{Text: "Europe moved to -2.1%.", EvidenceIDs: []string{"F-07"}},
	))
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if claim := got.Slides[0].Claims[0]; claim.Status != StatusVerified {
		t.Fatalf("status = %q (reason %q), want verified", claim.Status, claim.Reason)
	}
}

func TestValidateRejectsAClaimThatDropsTheSignOfANegativeValue(t *testing.T) {
	got, err := Validate("Deck", "Board", facts(), draft(
		DraftClaim{Text: "Europe moved to 2.1%.", EvidenceIDs: []string{"F-07"}},
	))
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	claim := got.Slides[0].Claims[0]
	if claim.Status != StatusRejected {
		t.Fatalf("status = %q, want rejected: a dropped sign states a figure the data does not contain", claim.Status)
	}
	if !strings.Contains(claim.Reason, "2.1") {
		t.Errorf("reason %q does not name 2.1", claim.Reason)
	}
}

func TestValidateTreatsAPeriodLabelAsALabel(t *testing.T) {
	got, err := Validate("Deck", "Board", facts(), draft(
		DraftClaim{Text: "In 2026-Q3 the total reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
	))
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if claim := got.Slides[0].Claims[0]; claim.Status != StatusVerified {
		t.Fatalf("status = %q (reason %q), want verified: a period label is not a figure", claim.Status, claim.Reason)
	}
}

func TestValidateNumbersClaimsAndSlidesInOrder(t *testing.T) {
	got, err := Validate("Deck", "Board", facts(), Draft{Slides: []DraftSlide{
		{Title: "One", Claims: []DraftClaim{
			{Text: "Total reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
			{Text: "LATAM reached 1,275,000.", EvidenceIDs: []string{"F-03"}},
		}},
		{Title: "Two", Claims: []DraftClaim{
			{Text: "APAC reached 2,460,000.", EvidenceIDs: []string{"F-04"}},
		}},
	}})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if got, want := got.Slides[0].Number, 1; got != want {
		t.Errorf("first slide number = %d, want %d", got, want)
	}
	if got, want := got.Slides[1].Number, 2; got != want {
		t.Errorf("second slide number = %d, want %d", got, want)
	}
	ids := []string{
		got.Slides[0].Claims[0].ID,
		got.Slides[0].Claims[1].ID,
		got.Slides[1].Claims[0].ID,
	}
	if want := []string{"C1", "C2", "C3"}; strings.Join(ids, ",") != strings.Join(want, ",") {
		t.Errorf("claim ids = %v, want %v", ids, want)
	}
}

func TestValidateKeepsTheSuppliedFacts(t *testing.T) {
	supplied := facts()
	got, err := Validate("Deck", "Board", supplied, draft(
		DraftClaim{Text: "Total reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
	))
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(got.Facts) != len(supplied) {
		t.Fatalf("facts = %d, want %d", len(got.Facts), len(supplied))
	}
	for i := range supplied {
		if got.Facts[i] != supplied[i] {
			t.Errorf("fact %d = %+v, want %+v", i, got.Facts[i], supplied[i])
		}
	}
}

func TestValidateRejectsAnEmptyDraft(t *testing.T) {
	_, err := Validate("Deck", "Board", facts(), Draft{})
	if !errors.Is(err, ErrNoSlides) {
		t.Fatalf("error = %v, want %v", err, ErrNoSlides)
	}
}

func TestVerifiedClaimsExcludesRejectedOnes(t *testing.T) {
	got, err := Validate("Deck", "Board", facts(), draft(
		DraftClaim{Text: "Total reached 11,695,000.", EvidenceIDs: []string{"F-05"}},
		DraftClaim{Text: "Europe declined 7.3%.", EvidenceIDs: []string{"F-07"}},
	))
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	verified := got.VerifiedClaims()
	if len(verified) != 1 {
		t.Fatalf("verified claims = %d, want 1", len(verified))
	}
	if verified[0].ID != "C1" {
		t.Errorf("verified claim = %q, want C1", verified[0].ID)
	}
	if got.RejectedCount != 1 {
		t.Errorf("rejectedCount = %d, want 1", got.RejectedCount)
	}
}

func TestCodeIsStableAndDistinct(t *testing.T) {
	first := Code("Q3 2026 revenue review", "Region,Revenue\nNorth,1\n")
	second := Code("Q3 2026 revenue review", "Region,Revenue\nNorth,1\n")
	other := Code("Q3 2026 revenue review", "Region,Revenue\nNorth,2\n")

	if first != second {
		t.Errorf("the same input produced %q and %q", first, second)
	}
	if first == other {
		t.Errorf("different inputs produced the same code %q", first)
	}
	if !strings.HasPrefix(first, "EX-") {
		t.Errorf("code %q does not carry the EX- prefix", first)
	}
	if got, want := len(first), len("EX-")+6; got != want {
		t.Errorf("code %q has length %d, want %d", first, got, want)
	}
}
