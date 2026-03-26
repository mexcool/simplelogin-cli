package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/itchyny/gojq"
	"github.com/mattn/go-isatty"
)

var (
	Green  = color.New(color.FgGreen)
	Red    = color.New(color.FgRed)
	Yellow = color.New(color.FgYellow)
	Bold   = color.New(color.Bold)
	Cyan   = color.New(color.FgCyan)
)

// PrintError prints an error message to stderr in red.
func PrintError(format string, a ...interface{}) {
	Red.Fprintf(os.Stderr, "Error: "+format+"\n", a...)
}

// PrintSuccess prints a success message to stderr in green.
func PrintSuccess(format string, a ...interface{}) {
	Green.Fprintf(os.Stderr, format+"\n", a...)
}

// PrintWarning prints a warning message to stderr in yellow.
func PrintWarning(format string, a ...interface{}) {
	Yellow.Fprintf(os.Stderr, format+"\n", a...)
}

// PrintJSON pretty-prints JSON data to stdout.
func PrintJSON(data []byte) error {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		fmt.Println(string(data))
		return nil
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}

// PrintJQ applies a jq expression to JSON data and prints the result.
func PrintJQ(data []byte, expr string) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid jq expression: %w", err)
	}

	var input interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	iter := query.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return fmt.Errorf("jq error: %w", err)
		}
		switch val := v.(type) {
		case string:
			fmt.Println(val)
		default:
			b, err := json.MarshalIndent(val, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal jq result: %w", err)
			}
			fmt.Println(string(b))
		}
	}
	return nil
}

// Table is a simple table renderer.
type Table struct {
	writer  io.Writer
	headers []string
	rows    [][]string
}

// NewTable creates a new table with standard formatting.
func NewTable(writer io.Writer, headers []string) *Table {
	return &Table{
		writer:  writer,
		headers: headers,
	}
}

// Append adds a row to the table.
func (t *Table) Append(row []string) {
	t.rows = append(t.rows, row)
}

// Render renders the table to the writer.
func (t *Table) Render() {
	if len(t.headers) == 0 {
		return
	}

	// Calculate column widths (using visible length, ignoring ANSI codes)
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = visibleLen(h)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) {
				vl := visibleLen(cell)
				if vl > widths[i] {
					widths[i] = vl
				}
			}
		}
	}

	// Print header
	printRow(t.writer, t.headers, widths, true)

	// Print rows
	for _, row := range t.rows {
		printRow(t.writer, row, widths, false)
	}
}

func printRow(w io.Writer, cells []string, widths []int, isHeader bool) {
	parts := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		vl := visibleLen(cell)
		padding := widths[i] - vl
		if padding < 0 {
			padding = 0
		}
		parts[i] = cell + strings.Repeat(" ", padding)
	}
	line := strings.Join(parts, "  ")
	if isHeader {
		fmt.Fprintln(w, Bold.Sprint(strings.TrimRight(line, " ")))
	} else {
		fmt.Fprintln(w, strings.TrimRight(line, " "))
	}
}

// visibleLen returns the visible length of a string, ignoring ANSI escape codes.
func visibleLen(s string) int {
	length := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		length++
	}
	return length
}

// BoolToStatus converts a bool to a colored status string.
func BoolToStatus(b bool) string {
	if b {
		return Green.Sprint("yes")
	}
	return Red.Sprint("no")
}

// EnabledStatus converts an enabled bool to a colored status string.
func EnabledStatus(enabled bool) string {
	if enabled {
		return Green.Sprint("enabled")
	}
	return Red.Sprint("disabled")
}

// StringOrEmpty returns the dereferenced string or empty string for nil.
func StringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Truncate truncates a string to maxLen characters.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// IsInteractive reports whether stdin is an interactive terminal.
func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// ConfirmAction prompts the user for confirmation.
// Returns false immediately in non-interactive mode (safe default).
func ConfirmAction(prompt string) bool {
	if !IsInteractive() {
		return false
	}
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
