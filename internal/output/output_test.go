package output

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func init() {
	// Disable color output in tests so assertions are deterministic.
	color.NoColor = true
}

// ---------------------------------------------------------------------------
// visibleLen
// ---------------------------------------------------------------------------

func TestVisibleLen_PlainString(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 5},
		{"hello world", 11},
		{"   ", 3},
	}
	for _, tt := range tests {
		if got := visibleLen(tt.input); got != tt.want {
			t.Errorf("visibleLen(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestVisibleLen_ANSIEscapes(t *testing.T) {
	// Simulated ANSI-colored string: ESC[32mhelloESC[0m  (green "hello")
	ansi := "\033[32mhello\033[0m"
	if got := visibleLen(ansi); got != 5 {
		t.Errorf("visibleLen(ansi green 'hello') = %d, want 5", got)
	}

	// Nested/multiple sequences: bold + color
	nested := "\033[1m\033[31mERROR\033[0m"
	if got := visibleLen(nested); got != 5 {
		t.Errorf("visibleLen(bold+red 'ERROR') = %d, want 5", got)
	}

	// String with escape in the middle
	mixed := "ok \033[33mwarn\033[0m end"
	// visible: "ok warn end" = 11
	if got := visibleLen(mixed); got != 11 {
		t.Errorf("visibleLen(mixed) = %d, want 11", got)
	}
}

// ---------------------------------------------------------------------------
// Truncate
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exact", 5, "exact"},
		// Size hint fits: "... [25]" = 8 chars, leaves 2 chars of prefix
		{"hello world, this is long", 10, "he... [25]"},
		{"abc", 3, "abc"},
		// maxLen=5, suffix "... [6]" = 7 chars > 5, falls back to plain "..."
		{"abcdef", 5, "ab..."},
		{"hello", 0, ""},
		{"hello", 1, "h"},
		{"hello", 2, "he"},
		{"hello", 3, "hel"},
		// Size hint with larger maxLen
		{"this is a much longer note that should be truncated", 30, "this is a much longer ... [51]"},
	}
	for _, tt := range tests {
		got := Truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// BoolToStatus / EnabledStatus
// ---------------------------------------------------------------------------

func TestBoolToStatus(t *testing.T) {
	yes := BoolToStatus(true)
	no := BoolToStatus(false)
	if yes != "yes" {
		t.Errorf("BoolToStatus(true) = %q, want %q", yes, "yes")
	}
	if no != "no" {
		t.Errorf("BoolToStatus(false) = %q, want %q", no, "no")
	}
}

func TestEnabledStatus(t *testing.T) {
	on := EnabledStatus(true)
	off := EnabledStatus(false)
	if on != "enabled" {
		t.Errorf("EnabledStatus(true) = %q, want %q", on, "enabled")
	}
	if off != "disabled" {
		t.Errorf("EnabledStatus(false) = %q, want %q", off, "disabled")
	}
}

// ---------------------------------------------------------------------------
// StringOrEmpty
// ---------------------------------------------------------------------------

func TestStringOrEmpty(t *testing.T) {
	val := "hello"
	if got := StringOrEmpty(&val); got != "hello" {
		t.Errorf("StringOrEmpty(&%q) = %q, want %q", val, got, "hello")
	}
	if got := StringOrEmpty(nil); got != "" {
		t.Errorf("StringOrEmpty(nil) = %q, want %q", got, "")
	}
}

// ---------------------------------------------------------------------------
// Table.Render
// ---------------------------------------------------------------------------

func TestTableRender_Basic(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable(&buf, []string{"Name", "Age"})
	tbl.Append([]string{"Alice", "30"})
	tbl.Append([]string{"Bob", "25"})
	tbl.Render()

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d:\n%s", len(lines), out)
	}

	// Header should contain both column names
	if !strings.Contains(lines[0], "Name") || !strings.Contains(lines[0], "Age") {
		t.Errorf("header missing column names: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Alice") || !strings.Contains(lines[1], "30") {
		t.Errorf("row 1 missing data: %q", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") || !strings.Contains(lines[2], "25") {
		t.Errorf("row 2 missing data: %q", lines[2])
	}
}

func TestTableRender_EmptyHeaders(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable(&buf, []string{})
	tbl.Append([]string{"data"})
	tbl.Render()
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty headers, got: %q", buf.String())
	}
}

func TestTableRender_ColumnWidths(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable(&buf, []string{"X", "Y"})
	tbl.Append([]string{"longvalue", "ZZZ"})
	tbl.Append([]string{"b", "QQQ"})
	tbl.Render()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// The column widths should be driven by "longvalue" (9 chars) in column 0.
	// Row 2: "b" should be padded so column Y aligns with row 1.
	// Verify that "QQQ" in row 2 starts at the same offset as "ZZZ" in row 1.
	idxZZZ := strings.Index(lines[1], "ZZZ")
	idxQQQ := strings.Index(lines[2], "QQQ")
	if idxZZZ != idxQQQ {
		t.Errorf("columns not aligned: 'ZZZ' at %d, 'QQQ' at %d", idxZZZ, idxQQQ)
	}
}

func TestTableRender_FewerCellsThanHeaders(t *testing.T) {
	// Rows with fewer cells than headers should not panic and should pad.
	var buf bytes.Buffer
	tbl := NewTable(&buf, []string{"A", "B", "C"})
	tbl.Append([]string{"only-one"})
	tbl.Render()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

// ---------------------------------------------------------------------------
// PrintJQ
// ---------------------------------------------------------------------------

func TestPrintJQ_ValidExpression(t *testing.T) {
	data := []byte(`{"name": "alice", "age": 30}`)

	// Capture stdout by redirecting os.Stdout temporarily is fragile,
	// so we just verify no error is returned for a valid expression.
	err := PrintJQ(data, ".name")
	if err != nil {
		t.Errorf("PrintJQ with valid expr returned error: %v", err)
	}
}

func TestPrintJQ_InvalidExpression(t *testing.T) {
	data := []byte(`{"a": 1}`)
	err := PrintJQ(data, ".[invalid")
	if err == nil {
		t.Error("PrintJQ with invalid expression should return error")
	}
	if !strings.Contains(err.Error(), "invalid jq expression") {
		t.Errorf("expected 'invalid jq expression' in error, got: %v", err)
	}
}

func TestPrintJQ_InvalidJSON(t *testing.T) {
	err := PrintJQ([]byte("not json"), ".foo")
	if err == nil {
		t.Error("PrintJQ with invalid JSON should return error")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("expected 'failed to parse JSON' in error, got: %v", err)
	}
}

func TestPrintJQ_ArrayIteration(t *testing.T) {
	data := []byte(`[1, 2, 3]`)
	err := PrintJQ(data, ".[]")
	if err != nil {
		t.Errorf("PrintJQ array iteration returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PrintJQ - verify output
// ---------------------------------------------------------------------------

func TestPrintJQ_VerifyOutput(t *testing.T) {
	data := []byte(`{"name": "alice", "age": 30}`)

	// Capture stdout using os.Pipe
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	err = PrintJQ(data, ".name")

	// Close writer and restore stdout before reading
	w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("PrintJQ returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	r.Close()

	got := strings.TrimSpace(buf.String())
	if got != "alice" {
		t.Errorf("PrintJQ(.name) output = %q, want %q", got, "alice")
	}
}

func TestPrintJQ_VerifyOutput_Number(t *testing.T) {
	data := []byte(`{"name": "alice", "age": 30}`)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	err = PrintJQ(data, ".age")

	w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("PrintJQ returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	r.Close()

	got := strings.TrimSpace(buf.String())
	if got != "30" {
		t.Errorf("PrintJQ(.age) output = %q, want %q", got, "30")
	}
}

// ---------------------------------------------------------------------------
// SelectColumns / FilterRow
// ---------------------------------------------------------------------------

func TestSelectColumns_Empty(t *testing.T) {
	headers := []string{"ID", "Email", "Status"}
	got := SelectColumns(headers, "")
	want := []int{0, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("SelectColumns(empty) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SelectColumns(empty)[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestSelectColumns_Subset(t *testing.T) {
	headers := []string{"ID", "Email", "Status", "Note"}
	got := SelectColumns(headers, "id,note")
	want := []int{0, 3}
	if len(got) != len(want) {
		t.Fatalf("SelectColumns = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SelectColumns[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestSelectColumns_CaseInsensitive(t *testing.T) {
	headers := []string{"Reverse Alias", "Blocked"}
	got := SelectColumns(headers, "reverse-alias")
	if len(got) != 1 || got[0] != 0 {
		t.Errorf("SelectColumns(reverse-alias) = %v, want [0]", got)
	}
}

func TestSelectColumns_AllInvalid(t *testing.T) {
	headers := []string{"ID", "Email"}
	got := SelectColumns(headers, "bogus,nope")
	// Should fall back to all columns
	if len(got) != 2 {
		t.Errorf("SelectColumns(all invalid) returned %d indices, want 2 (fallback)", len(got))
	}
}

func TestFilterRow(t *testing.T) {
	row := []string{"1", "test@example.com", "enabled", "my note"}
	got := FilterRow(row, []int{0, 3})
	if len(got) != 2 || got[0] != "1" || got[1] != "my note" {
		t.Errorf("FilterRow = %v, want [1, my note]", got)
	}
}

// ---------------------------------------------------------------------------
// Truncate - edge cases
// ---------------------------------------------------------------------------

func TestTruncate_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"maxLen=0", "hello", 0, ""},
		{"maxLen=1", "hello", 1, "h"},
		{"maxLen=2", "hello", 2, "he"},
		{"maxLen=3", "hello", 3, "hel"},
		{"maxLen=5_exact_length", "hello", 5, "hello"},
		{"maxLen=10_longer_than_string", "hello", 10, "hello"},
		{"empty_string_maxLen=5", "", 5, ""},
		{"size_hint_fits", "abcdefghijklmnopqrstuvwxyz", 15, "abcdefg... [26]"},
		{"size_hint_boundary", "abcdefghij", 9, "a... [10]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
