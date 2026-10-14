package plan

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GhandyP/datadeck/internal/data"
)

// Validate turns the model's answer into the deck that will be shown, keeping
// every claim whose figures the cited data supports and rejecting the rest with
// a reason the user can read.
//
// The facts on the returned plan are always the ones passed in: the model
// contributes prose, never numbers.
func Validate(title, audience string, facts []data.Fact, draft Draft) (Plan, error) {
	if len(draft.Slides) == 0 {
		return Plan{}, ErrNoSlides
	}

	byID := make(map[string]data.Fact, len(facts))
	for _, fact := range facts {
		byID[fact.ID] = fact
	}

	result := Plan{
		Title:    title,
		Audience: audience,
		Status:   StatusDraft,
		Facts:    append([]data.Fact(nil), facts...),
	}

	claims := 0
	for index, draftSlide := range draft.Slides {
		slide := Slide{
			Number: index + 1,
			Title:  strings.TrimSpace(draftSlide.Title),
			Notes:  strings.TrimSpace(draftSlide.Notes),
		}
		for _, draftClaim := range draftSlide.Claims {
			claims++
			claim := Claim{
				ID:          fmt.Sprintf("C%d", claims),
				Text:        strings.TrimSpace(draftClaim.Text),
				EvidenceIDs: draftClaim.EvidenceIDs,
			}
			claim.Status, claim.Reason = judge(claim, byID)
			if claim.Status == StatusRejected {
				result.RejectedCount++
			}
			slide.Claims = append(slide.Claims, claim)
		}
		result.Slides = append(result.Slides, slide)
	}
	return result, nil
}

// judge decides whether a claim may be shown and, when it may not, says why in
// the user's terms rather than the validator's.
func judge(claim Claim, byID map[string]data.Fact) (ClaimStatus, string) {
	if len(claim.EvidenceIDs) == 0 {
		return StatusRejected, "the claim cites no source value, so nothing in the supplied data supports it"
	}

	cited := make([]data.Fact, 0, len(claim.EvidenceIDs))
	for _, id := range claim.EvidenceIDs {
		fact, ok := byID[strings.TrimSpace(id)]
		if !ok {
			return StatusRejected, fmt.Sprintf("the claim cites %s, which is not one of the supplied values", strings.TrimSpace(id))
		}
		cited = append(cited, fact)
	}

	accepted := make([]string, 0, len(cited))
	for _, fact := range cited {
		accepted = append(accepted, fact.Value)
	}
	for _, token := range numbersIn(claim.Text) {
		if !supported(token, cited) {
			return StatusRejected, fmt.Sprintf(
				"the claim states %s, which the cited source data does not contain; it says %s",
				token, strings.Join(accepted, ", "))
		}
	}
	return StatusVerified, ""
}

// supported reports whether one figure in a claim matches a value of the data
// the claim cites.
func supported(token string, cited []data.Fact) bool {
	value, percent, ok := canonicalNumber(token)
	if !ok {
		return false
	}
	for _, fact := range cited {
		factValue, factPercent, ok := canonicalNumber(fact.Value)
		if !ok || factPercent != percent {
			continue
		}
		if value == factValue {
			return true
		}
	}
	return false
}

// canonicalNumber reduces a figure to a comparable value. It is used for
// comparison only: the text a deck shows is always the source's own formatting.
func canonicalNumber(token string) (value float64, percent bool, ok bool) {
	body := strings.TrimSpace(token)
	if body == "" {
		return 0, false, false
	}
	body = strings.NewReplacer("\u2212", "-", "\u2013", "-", "\u2014", "-").Replace(body)
	if strings.HasSuffix(body, "%") {
		percent = true
		body = strings.TrimSuffix(body, "%")
	}
	body = strings.Map(func(r rune) rune {
		switch r {
		case '$', '\u20ac', '\u00a3', '\u00a5', ' ', '\u00a0', '\u202f', '\'':
			return -1
		}
		return r
	}, body)
	body = strings.TrimPrefix(body, "+")

	// Keep only the decimal separator: a separator followed by exactly three
	// digits with no separator before it is grouping, not a decimal point.
	if last := lastSeparatorIndex(body); last >= 0 {
		digitsAfter := len(body) - last - 1
		earlier := strings.ContainsAny(body[:last], ".,")
		body = strings.NewReplacer(",", "", ".", "").Replace(body)
		if !(digitsAfter == 3 && !earlier) {
			at := len(body) - digitsAfter
			body = body[:at] + "." + body[at:]
		}
	}

	parsed, err := strconv.ParseFloat(body, 64)
	if err != nil {
		return 0, false, false
	}
	return parsed, percent, true
}

func lastSeparatorIndex(body string) int {
	last := -1
	for i := 0; i < len(body); i++ {
		if body[i] == '.' || body[i] == ',' {
			last = i
		}
	}
	return last
}

// numbersIn returns the figures a claim states, sign included: the sign is part
// of the figure, so a claim that drops it states a different number. A run of
// digits that belongs to a word or to a label such as 2026-Q3 is not a figure.
func numbersIn(text string) []string {
	var tokens []string
	runes := []rune(text)
	for i := 0; i < len(runes); {
		if !isDigit(runes[i]) {
			i++
			continue
		}
		start := i
		for i < len(runes) {
			if isDigit(runes[i]) {
				i++
				continue
			}
			if (runes[i] == '.' || runes[i] == ',') && i+1 < len(runes) && isDigit(runes[i+1]) {
				i += 2
				continue
			}
			break
		}
		end := i
		if i < len(runes) && runes[i] == '%' {
			i++
		}
		from := signStart(runes, start)
		if isLabel(runes, from, end, i) {
			continue
		}
		tokens = append(tokens, string(runes[from:i]))
	}
	return tokens
}

// signStart extends a numeric run to its sign when the sign is a real unary
// sign rather than the hyphen inside a label such as Q3-2026.
func signStart(runes []rune, start int) int {
	at := start - 1
	if at < 0 || (runes[at] != '-' && runes[at] != '+') {
		return start
	}
	if at > 0 {
		before := runes[at-1]
		if isLetter(before) || isDigit(before) || before == '%' {
			return start
		}
	}
	return at
}

// isLabel reports whether a numeric run belongs to a word or to a label such as
// 2026-Q3 or Q3-2026.
func isLabel(runes []rune, start, end, after int) bool {
	if start > 0 && isLetter(runes[start-1]) {
		return true
	}
	if after < len(runes) && isLetter(runes[after]) {
		return true
	}
	if after+1 < len(runes) && runes[after] == '-' && isLetter(runes[after+1]) {
		return true
	}
	if start > 1 && runes[start-1] == '-' && (isLetter(runes[start-2]) || isDigit(runes[start-2])) {
		return true
	}
	_ = end
	return false
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r > 127
}
