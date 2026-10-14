package data

import (
	"fmt"
	"strconv"
	"strings"
)

// WholeColumn is the Row of a fact computed over an entire column rather than
// over one cell. A reader that shows provenance lights the whole column for it.
const WholeColumn = -1

// Fact is one value a deck may state, together with the source cell it came
// from. Value is always either a cell the user supplied or an exact sum over
// one of its columns, formatted in that column's own convention.
type Fact struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Value  string `json:"value"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
}

// ComputeFacts derives the facts of a table. It never invents a value: a cell
// that is not a number produces no fact at all, and a column with no numbers is
// skipped rather than guessed at.
func ComputeFacts(table Table) ([]Fact, error) {
	var facts []Fact

	for column := range table.Columns {
		literals := make([]string, len(table.Rows))
		numbers, percents := 0, 0
		for row := range table.Rows {
			literal, percent, ok := numericLiteral(table.Rows[row][column])
			if !ok {
				continue
			}
			literals[row] = literal
			numbers++
			if percent {
				percents++
			}
		}
		if numbers == 0 {
			continue
		}

		// A column is a percentage column only when every number in it is one.
		style := inferStyle(literals, percents == numbers)

		values := make([]int64, len(table.Rows))
		for row, literal := range literals {
			if literal == "" {
				continue
			}
			value, ok := parseScaled(literal, style)
			if !ok {
				continue
			}
			values[row] = value
			facts = append(facts, Fact{
				Label:  rowLabel(table, row, column) + " · " + table.Columns[column],
				Value:  style.format(value),
				Row:    row,
				Column: column,
			})
		}

		if !style.percent {
			var total int64
			for row, literal := range literals {
				if literal == "" {
					continue
				}
				total += values[row]
			}
			facts = append(facts, Fact{
				Label:  table.Columns[column] + " · total",
				Value:  style.format(total),
				Row:    WholeColumn,
				Column: column,
			})
		}

		if row, value, ok := extreme(literals, values, true); ok {
			facts = append(facts, Fact{
				Label:  table.Columns[column] + " · highest",
				Value:  style.format(value),
				Row:    row,
				Column: column,
			})
		}
		if row, value, ok := extreme(literals, values, false); ok {
			facts = append(facts, Fact{
				Label:  table.Columns[column] + " · lowest",
				Value:  style.format(value),
				Row:    row,
				Column: column,
			})
		}
	}

	if len(facts) == 0 {
		return nil, ErrNoNumericColumn
	}
	for i := range facts {
		facts[i].ID = fmt.Sprintf("F-%02d", i+1)
	}
	return facts, nil
}

// extreme finds the row holding the largest or smallest value of a column.
func extreme(literals []string, values []int64, highest bool) (int, int64, bool) {
	row, found := 0, false
	for i, literal := range literals {
		if literal == "" {
			continue
		}
		if !found {
			row, found = i, true
			continue
		}
		if (highest && values[i] > values[row]) || (!highest && values[i] < values[row]) {
			row = i
		}
	}
	if !found {
		return 0, 0, false
	}
	return row, values[row], true
}

// rowLabel names the row a fact belongs to, using the table's own text rather
// than a row number whenever the table has a text column to take it from.
func rowLabel(table Table, row, column int) string {
	candidate := ""
	if column != 0 {
		candidate = table.Rows[row][0]
	} else {
		for i := 1; i < len(table.Columns); i++ {
			if _, _, ok := numericLiteral(table.Rows[row][i]); !ok {
				candidate = table.Rows[row][i]
				break
			}
		}
	}
	if candidate == "" {
		return fmt.Sprintf("row %d", row+1)
	}
	return candidate
}

// numberStyle is how one column writes its numbers. It is inferred from the
// column itself, so a table written in either convention keeps it.
type numberStyle struct {
	decimal  byte
	grouping byte
	decimals int
	percent  bool
}

// inferStyle reads the column's own convention: the last separator is the
// decimal separator when another separator precedes it or when fewer than three
// digits follow it, and a grouping separator otherwise.
func inferStyle(literals []string, percent bool) numberStyle {
	style := numberStyle{decimal: '.', grouping: ',', percent: percent}
	decimalSeen := false
	for _, literal := range literals {
		if literal == "" {
			continue
		}
		separator, decimals, hasSeparator, earlier := lastSeparator(literal)
		if !hasSeparator {
			continue
		}
		if earlier || decimals != 3 {
			style.decimal = separator
			decimalSeen = true
			if decimals > style.decimals {
				style.decimals = decimals
			}
		}
	}
	if decimalSeen && style.decimal == ',' {
		style.grouping = '.'
	}
	return style
}

func lastSeparator(literal string) (separator byte, decimals int, hasSeparator, earlier bool) {
	last := -1
	for i := 0; i < len(literal); i++ {
		if literal[i] == '.' || literal[i] == ',' {
			last = i
		}
	}
	if last < 0 {
		return 0, 0, false, false
	}
	for i := 0; i < last; i++ {
		if literal[i] == '.' || literal[i] == ',' {
			earlier = true
			break
		}
	}
	return literal[last], len(literal) - last - 1, true, earlier
}

// parseScaled reads a literal as an integer scaled by the column's decimal
// count, which keeps sums exact instead of rounding them through a float.
func parseScaled(literal string, style numberStyle) (int64, bool) {
	negative := false
	body := literal
	if strings.HasPrefix(body, "-") {
		negative = true
		body = body[1:]
	} else if strings.HasPrefix(body, "+") {
		body = body[1:]
	}
	if style.grouping != 0 {
		body = strings.ReplaceAll(body, string(style.grouping), "")
	}

	whole, fraction := body, ""
	if i := strings.IndexByte(body, style.decimal); i >= 0 {
		whole, fraction = body[:i], body[i+1:]
	}
	if strings.ContainsAny(whole, ".,") || strings.ContainsAny(fraction, ".,") {
		return 0, false
	}
	if whole == "" {
		whole = "0"
	}
	if len(fraction) > style.decimals {
		fraction = fraction[:style.decimals]
	}
	for len(fraction) < style.decimals {
		fraction += "0"
	}

	digits := strings.TrimLeft(whole, "0") + fraction
	if digits == "" {
		digits = "0"
	}
	value, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return 0, false
	}
	if negative {
		value = -value
	}
	return value, true
}

// format writes a scaled value back in the column's own convention.
func (s numberStyle) format(value int64) string {
	negative := value < 0
	if negative {
		value = -value
	}
	scale := pow10(s.decimals)
	out := group(strconv.FormatInt(value/scale, 10), s.grouping)
	if s.decimals > 0 {
		out += string(s.decimal) + fmt.Sprintf("%0*d", s.decimals, value%scale)
	}
	if s.percent {
		out += "%"
	}
	if negative {
		out = "-" + out
	}
	return out
}

// group inserts a grouping separator every three digits.
func group(digits string, separator byte) string {
	if separator == 0 || len(digits) <= 3 {
		return digits
	}
	var out strings.Builder
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out.WriteByte(separator)
		}
		out.WriteByte(digits[i])
	}
	return out.String()
}

func pow10(exponent int) int64 {
	result := int64(1)
	for i := 0; i < exponent; i++ {
		result *= 10
	}
	return result
}

// numericLiteral reports whether a cell holds a number and returns it without
// its currency symbol, thousands separators kept, so the column's convention can
// still be read from it.
func numericLiteral(cell string) (literal string, percent bool, ok bool) {
	body := strings.TrimSpace(cell)
	if body == "" {
		return "", false, false
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
	if body == "" {
		return "", false, false
	}

	digits := strings.TrimLeft(body, "+-")
	if digits == "" {
		return "", false, false
	}
	for i := 0; i < len(digits); i++ {
		c := digits[i]
		if (c < '0' || c > '9') && c != '.' && c != ',' {
			return "", false, false
		}
	}
	return body, percent, true
}
