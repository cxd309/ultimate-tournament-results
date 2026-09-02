package convert

import "fmt"

// IntBool decodes the Live! by BULA API's IntBool convention:
// a boolean carried as a JSON integer (0 or 1) rather than a JSON boolean,
// and converts it to the int64 a NOT NULL IntBool-derived DB column stores it as
//
// Falsy values are stripped from Live!'s JSON responses entirely
// so an IntBool field is often simply absent rather than sent as 0
// the zero value is false, so a missing key already decodes correctly
type IntBool bool

func (b *IntBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "0":
		*b = false
	case "1":
		*b = true
	default:
		return fmt.Errorf("IntBool: unexpected value %s", data)
	}
	return nil
}

// Int64 converts to the int64 (0 or 1) a NOT NULL IntBool-derived column stores it as.
func (b IntBool) Int64() int64 {
	if b {
		return 1
	}
	return 0
}

// IntBoolFromInt64 converts an int64 (0 or 1) column value back to IntBool.
func IntBoolFromInt64(i int64) IntBool {
	return i != 0
}
