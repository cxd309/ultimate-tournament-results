package uoclient

import (
	"fmt"
	"strings"
)

// parseExportCSV parses the output of UltiOrganizer's ResultsetToCsv, which every
// ext/*csv.php export uses. It is close to CSV but not RFC 4180:
//   - a non-empty value is wrapped in double quotes, an empty one is left bare
//   - a quote inside a value is escaped with a backslash (\"), not doubled
//   - nothing else is escaped, so a value can run over several lines
//   - every row, the header included, ends with "\n"
//
// encoding/csv rejects the first unescaped-by-its-rules quote, which a spirit comment
// quoting someone produces, so the whole file would be lost
//
// \" is always read as an escaped quote, so a comment quoting someone mid-sentence
// ("said \"go\", then") stays whole. The one input that can't be told apart is a
// value whose own last character is a backslash, which swallows its closing quote;
// getCSV catches that as a row with the wrong number of columns
func parseExportCSV(body string) ([][]string, error) {
	var records [][]string
	var record []string
	i, n := 0, len(body)
	for i < n {
		var field strings.Builder
		if body[i] == '"' {
			i++
			for {
				if i >= n {
					return nil, fmt.Errorf("row %d: unterminated quoted value", len(records)+1)
				}
				c := body[i]
				if c == '\\' && i+1 < n && body[i+1] == '"' {
					field.WriteByte('"')
					i += 2
					continue
				}
				if c == '"' && endsValue(body, i+1) {
					i++
					break
				}
				field.WriteByte(c)
				i++
			}
		} else {
			for i < n && body[i] != ',' && body[i] != '\n' {
				field.WriteByte(body[i])
				i++
			}
		}
		record = append(record, field.String())

		switch {
		case i >= n:
			records = append(records, record)
			return records, nil
		case body[i] == ',':
			i++
		case body[i] == '\n':
			i++
			records = append(records, record)
			record = nil
		default:
			return nil, fmt.Errorf("row %d: unexpected %q after a quoted value", len(records)+1, body[i])
		}
	}
	if record != nil {
		records = append(records, record)
	}
	return records, nil
}

// endsValue is whether position i is where a value can end: the separator, the row
// terminator, or the end of the file
func endsValue(body string, i int) bool {
	return i >= len(body) || body[i] == ',' || body[i] == '\n'
}
