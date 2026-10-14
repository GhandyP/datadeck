package data

import (
	"errors"
	"strings"
	"testing"
)

// kpiTable is the fixture the README and the decision round both use.
const kpiTable = `Region,Revenue,Change
North America,4820000,+6.4%
Europe,3140000,-2.1%
LATAM,1275000,+11.8%
APAC,2460000,+0.4%
`

func TestParseDetectsDelimiters(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"comma", "Region,Revenue\nNorth,4820000\n"},
		{"semicolon", "Region;Revenue\nNorth;4820000\n"},
		{"tab", "Region\tRevenue\nNorth\t4820000\n"},
		{"pipe", "Region|Revenue\nNorth|4820000\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse(%s) returned error: %v", tc.name, err)
			}
			if got, want := len(table.Columns), 2; got != want {
				t.Fatalf("columns = %d, want %d", got, want)
			}
			if got, want := table.Columns[0], "Region"; got != want {
				t.Errorf("first column = %q, want %q", got, want)
			}
			if got, want := len(table.Rows), 1; got != want {
				t.Fatalf("rows = %d, want %d", got, want)
			}
			if got, want := table.Rows[0][1], "4820000"; got != want {
				t.Errorf("cell = %q, want %q", got, want)
			}
		})
	}
}

func TestParseMarkdownTable(t *testing.T) {
	raw := "| Region | Revenue |\n| --- | ---: |\n| North | 4820000 |\n| Europe | 3140000 |\n"
	table, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got, want := len(table.Columns), 2; got != want {
		t.Fatalf("columns = %d, want %d", got, want)
	}
	if got, want := len(table.Rows), 2; got != want {
		t.Fatalf("rows = %d, want %d", got, want)
	}
	if got, want := table.Rows[1][0], "Europe"; got != want {
		t.Errorf("cell = %q, want %q", got, want)
	}
}

func TestParseJSONRecords(t *testing.T) {
	raw := `[{"Region":"North","Revenue":4820000},{"Region":"Europe","Revenue":3140000}]`
	table, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got, want := strings.Join(table.Columns, "|"), "Region|Revenue"; got != want {
		t.Fatalf("columns = %q, want %q", got, want)
	}
	if got, want := len(table.Rows), 2; got != want {
		t.Fatalf("rows = %d, want %d", got, want)
	}
	// A JSON number must keep the text the user wrote, not a re-formatted float.
	if got, want := table.Rows[0][1], "4820000"; got != want {
		t.Errorf("cell = %q, want %q", got, want)
	}
}

func TestParseRejections(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr error
		wantMsg string
	}{
		{
			name:    "empty input",
			raw:     "   \n\n",
			wantErr: ErrEmptyInput,
		},
		{
			name:    "header only",
			raw:     "Region,Revenue\n",
			wantErr: ErrNoRows,
		},
		{
			name:    "single column",
			raw:     "Region\nNorth\nEurope\n",
			wantErr: ErrTooFewColumns,
		},
		{
			name:    "ragged row names the row",
			raw:     "Region,Revenue,Change\nNorth,4820000,+6.4%\nEurope,3140000\n",
			wantErr: ErrRaggedRow,
			wantMsg: "row 2",
		},
		{
			name:    "duplicate header names the column",
			raw:     "Region,Region\nNorth,South\n",
			wantErr: ErrDuplicateColumn,
			wantMsg: "Region",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.raw)
			if err == nil {
				t.Fatalf("Parse accepted input it must reject")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantMsg != "" && !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error %q does not name %q", err.Error(), tc.wantMsg)
			}
		})
	}
}

func TestParseTrimsAndDropsBlankRows(t *testing.T) {
	raw := " Region , Revenue \n North , 4820000 \n\n , \n"
	table, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got, want := table.Columns[0], "Region"; got != want {
		t.Errorf("column = %q, want %q", got, want)
	}
	if got, want := len(table.Rows), 1; got != want {
		t.Fatalf("rows = %d, want %d (blank rows must not become data)", got, want)
	}
	if got, want := table.Rows[0][0], "North"; got != want {
		t.Errorf("cell = %q, want %q", got, want)
	}
}

func TestComputeFactsOnKPITable(t *testing.T) {
	table, err := Parse(kpiTable)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	facts, err := ComputeFacts(table)
	if err != nil {
		t.Fatalf("ComputeFacts returned error: %v", err)
	}

	byID := make(map[string]Fact, len(facts))
	for _, f := range facts {
		if _, dup := byID[f.ID]; dup {
			t.Fatalf("duplicate fact id %q", f.ID)
		}
		byID[f.ID] = f
	}

	// The cell a claim cites must be locatable in the source table.
	var northRevenue *Fact
	for i := range facts {
		f := facts[i]
		if f.Label == "North America · Revenue" {
			northRevenue = &f
		}
	}
	if northRevenue == nil {
		t.Fatalf("no fact for North America · Revenue; facts: %v", labels(facts))
	}
	if got, want := northRevenue.Value, "4,820,000"; got != want {
		t.Errorf("value = %q, want %q", got, want)
	}
	if northRevenue.Row != 0 || northRevenue.Column != 1 {
		t.Errorf("source = (row %d, column %d), want (0, 1)", northRevenue.Row, northRevenue.Column)
	}

	// A total is computed over a column, so it cites the column, not a cell.
	var total *Fact
	for i := range facts {
		f := facts[i]
		if f.Label == "Revenue · total" {
			total = &f
		}
	}
	if total == nil {
		t.Fatalf("no fact for the Revenue total; facts: %v", labels(facts))
	}
	if got, want := total.Value, "11,695,000"; got != want {
		t.Errorf("total = %q, want %q", got, want)
	}
	if total.Row != WholeColumn {
		t.Errorf("total source row = %d, want WholeColumn (%d)", total.Row, WholeColumn)
	}
	if total.Column != 1 {
		t.Errorf("total source column = %d, want 1", total.Column)
	}

	// A percentage keeps the text the user wrote, sign included.
	var europeChange *Fact
	for i := range facts {
		f := facts[i]
		if f.Label == "Europe · Change" {
			europeChange = &f
		}
	}
	if europeChange == nil {
		t.Fatalf("no fact for Europe · Change; facts: %v", labels(facts))
	}
	if got, want := europeChange.Value, "-2.1%"; got != want {
		t.Errorf("value = %q, want %q", got, want)
	}
}

func TestComputeFactsKeepsEachColumnConvention(t *testing.T) {
	cases := []struct {
		name      string
		raw       string
		wantTotal string
	}{
		{
			name:      "comma groups, dot decimals",
			raw:       "Region,Revenue\nA,\"1,234.50\"\nB,\"2,000.25\"\n",
			wantTotal: "3,234.75",
		},
		{
			name:      "dot groups, comma decimals",
			raw:       "Region,Revenue\nA,\"1.234,50\"\nB,\"2.000,25\"\n",
			wantTotal: "3.234,75",
		},
		{
			name:      "plain integers keep the grouping convention of the column",
			raw:       "Region,Revenue\nA,4820000\nB,3140000\n",
			wantTotal: "7,960,000",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table, err := Parse(tc.raw)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			facts, err := ComputeFacts(table)
			if err != nil {
				t.Fatalf("ComputeFacts returned error: %v", err)
			}
			var total string
			for _, f := range facts {
				if f.Label == "Revenue · total" {
					total = f.Value
				}
			}
			if total != tc.wantTotal {
				t.Errorf("total = %q, want %q", total, tc.wantTotal)
			}
		})
	}
}

func TestComputeFactsSkipsNonNumericColumns(t *testing.T) {
	table, err := Parse(kpiTable)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	facts, err := ComputeFacts(table)
	if err != nil {
		t.Fatalf("ComputeFacts returned error: %v", err)
	}
	for _, f := range facts {
		if strings.Contains(f.Label, "Region") && f.Row == WholeColumn {
			t.Errorf("a text column produced a computed fact: %+v", f)
		}
	}
}

func TestComputeFactsRejectsTableWithNoNumbers(t *testing.T) {
	table, err := Parse("Region,Note\nNorth,ok\nEurope,ok\n")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	_, err = ComputeFacts(table)
	if !errors.Is(err, ErrNoNumericColumn) {
		t.Fatalf("error = %v, want %v", err, ErrNoNumericColumn)
	}
}

func labels(facts []Fact) []string {
	out := make([]string, 0, len(facts))
	for _, f := range facts {
		out = append(out, f.Label+"="+f.Value)
	}
	return out
}
