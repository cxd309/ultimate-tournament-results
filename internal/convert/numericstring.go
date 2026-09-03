package convert

import "encoding/json"

// NumericString decodes a Live! API field that can arrive as either a JSON
// string or a bare JSON number, depending on whether its value happens to
// look numeric
//
// the API's own DataGenerator coercion re-encodes any all-digit string as a
// number unless the field is blacklisted, a hex color is the common case
// ("0000FF" stays a string, "000000" becomes the number 0)
type NumericString string

func (n *NumericString) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*n = NumericString(s)
		return nil
	}
	*n = NumericString(data)
	return nil
}

// MarshalJSON replicates the same coercion on the way out
// an all-digit value is emitted as a bare number, matching what the live
// API itself would send, anything else is a quoted string
func (n NumericString) MarshalJSON() ([]byte, error) {
	if isAllDigits(string(n)) {
		return []byte(n), nil
	}
	return json.Marshal(string(n))
}

// isAllDigits mirrors PHP's ctype_digit: non-empty, every rune 0-9
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
