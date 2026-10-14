// Package data turns the text a user pastes into a table and then into the
// facts a deck may state.
//
// Parsing accepts delimited text (comma, semicolon, tab or pipe), a Markdown
// table, and a JSON array of records. Computing facts is deterministic: every
// value a fact carries is either a cell the user supplied or an exact sum over
// one of its columns, and every fact remembers the source cell it came from so
// a claim that cites it can be traced back and shown.
package data

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Errors a parse can report. Each is a distinct value so a caller can tell an
// unusable input from an empty one, and each message names the row or column
// the user has to look at.
var (
	ErrEmptyInput      = errors.New("no data was supplied")
	ErrNoRows          = errors.New("the input has a header but no data rows")
	ErrTooFewColumns   = errors.New("the input needs at least two columns")
	ErrRaggedRow       = errors.New("a row has a different number of cells than the header")
	ErrDuplicateColumn = errors.New("two columns share the same name")
	ErrNoNumericColumn = errors.New("no column holds numbers")
	ErrNotRecords      = errors.New("the JSON must be an array of records")
)

// Table is the parsed input: a header row and the data rows under it.
type Table struct {
	Columns []string
	Rows    [][]string
}

// Parse reads the pasted text and returns the table it holds.
func Parse(raw string) (Table, error) {
	if strings.TrimSpace(raw) == "" {
		return Table{}, ErrEmptyInput
	}

	var (
		columns []string
		rows    [][]string
		err     error
	)

	trimmed := strings.TrimSpace(raw)
	switch {
	case strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{"):
		columns, rows, err = parseJSON(trimmed)
	case isMarkdownTable(trimmed):
		columns, rows, err = parseMarkdown(trimmed)
	default:
		columns, rows, err = parseDelimited(trimmed)
	}
	if err != nil {
		return Table{}, err
	}
	return normalize(columns, rows)
}

// isMarkdownTable reports whether the first line is a pipe row, which is what
// distinguishes a pasted Markdown table from pipe-delimited text.
func isMarkdownTable(raw string) bool {
	first, _, _ := strings.Cut(raw, "\n")
	return strings.HasPrefix(strings.TrimSpace(first), "|")
}

func parseDelimited(raw string) ([]string, [][]string, error) {
	reader := csv.NewReader(strings.NewReader(raw))
	reader.Comma = detectDelimiter(raw)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("read delimited text: %w", err)
	}
	if len(records) == 0 {
		return nil, nil, ErrEmptyInput
	}
	return records[0], records[1:], nil
}

// detectDelimiter picks the candidate that appears most often in the header
// line, which is the line a user's paste always gets right.
func detectDelimiter(raw string) rune {
	header, _, _ := strings.Cut(raw, "\n")
	best, bestCount := ',', 0
	for _, candidate := range []rune{',', ';', '\t', '|'} {
		if count := strings.Count(header, string(candidate)); count > bestCount {
			best, bestCount = candidate, count
		}
	}
	return best
}

func parseMarkdown(raw string) ([]string, [][]string, error) {
	lines := strings.Split(raw, "\n")
	cells := func(line string) []string {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSuffix(line, "|")
		parts := strings.Split(line, "|")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}

	var (
		columns []string
		rows    [][]string
	)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		switch {
		case columns == nil:
			columns = cells(line)
		case i == 1 && isSeparatorRow(line):
			// The dashes under a Markdown header are not data.
		default:
			rows = append(rows, cells(line))
		}
	}
	if columns == nil {
		return nil, nil, ErrEmptyInput
	}
	return columns, rows, nil
}

// isSeparatorRow reports whether a Markdown table's second line is its dashes.
func isSeparatorRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	for _, r := range trimmed {
		switch r {
		case '-', ':', '|', ' ', '\t':
		default:
			return false
		}
	}
	return strings.ContainsRune(trimmed, '-')
}

// parseJSON reads an array of records. Column order follows the first
// appearance of each key, and a number keeps the text the user wrote because
// the decoder is asked for json.Number rather than float64.
func parseJSON(raw string) ([]string, [][]string, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()

	token, err := decoder.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("read JSON: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return nil, nil, ErrNotRecords
	}

	var (
		columns []string
		seen    = map[string]bool{}
		rows    [][]string
	)
	for decoder.More() {
		record := map[string]string{}
		order := []string{}

		if token, err = decoder.Token(); err != nil {
			return nil, nil, fmt.Errorf("read JSON record: %w", err)
		}
		if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
			return nil, nil, ErrNotRecords
		}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, nil, fmt.Errorf("read JSON key: %w", err)
			}
			key, _ := keyToken.(string)
			var value any
			if err := decoder.Decode(&value); err != nil {
				return nil, nil, fmt.Errorf("read JSON value for %q: %w", key, err)
			}
			record[key] = jsonCell(value)
			order = append(order, key)
			if !seen[key] {
				seen[key] = true
				columns = append(columns, key)
			}
		}
		if _, err := decoder.Token(); err != nil { // closing '}'
			return nil, nil, fmt.Errorf("read JSON record end: %w", err)
		}

		row := make([]string, len(columns))
		for i, column := range columns {
			row[i] = record[column]
		}
		rows = append(rows, row)
	}
	return columns, rows, nil
}

func jsonCell(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case bool:
		return fmt.Sprintf("%t", typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(encoded)
	}
}

// normalize trims the header and the cells, rejects the shapes a deck cannot
// use, and drops rows that carry nothing.
func normalize(columns []string, rows [][]string) (Table, error) {
	cleaned := make([]string, len(columns))
	seen := make(map[string]bool, len(columns))
	for i, column := range columns {
		name := strings.TrimSpace(column)
		if name == "" {
			name = fmt.Sprintf("Column %d", i+1)
		}
		if seen[name] {
			return Table{}, fmt.Errorf("%w: %q", ErrDuplicateColumn, name)
		}
		seen[name] = true
		cleaned[i] = name
	}
	if len(cleaned) < 2 {
		return Table{}, ErrTooFewColumns
	}

	var kept [][]string
	for index, row := range rows {
		cells := make([]string, len(row))
		blank := true
		for i, cell := range row {
			cells[i] = strings.TrimSpace(cell)
			if cells[i] != "" {
				blank = false
			}
		}
		if blank {
			continue
		}
		if len(cells) != len(cleaned) {
			return Table{}, fmt.Errorf("%w: row %d has %d cells but the header has %d",
				ErrRaggedRow, index+1, len(cells), len(cleaned))
		}
		kept = append(kept, cells)
	}
	if len(kept) == 0 {
		return Table{}, ErrNoRows
	}
	return Table{Columns: cleaned, Rows: kept}, nil
}
