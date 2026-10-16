package uoclient

import (
	"reflect"
	"testing"
)

func TestParseExportCSV(t *testing.T) {
	// ResultsetToCsv output: bare empty values, \" escapes, a multi-line value
	body := "\"a\",\"b\",\"comments\"\n" +
		"\"1\",,\"plain\"\n" +
		"\"2\",\"0\",\"said \\\"go\\\", then\nleft\"\n" +
		"\"3\",\"x\",\n"
	got, err := parseExportCSV(body)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"a", "b", "comments"},
		{"1", "", "plain"},
		{"2", "0", "said \"go\", then\nleft"},
		{"3", "x", ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if _, err := parseExportCSV("\"a\",\"never closed\n"); err == nil {
		t.Error("accepted an unterminated value")
	}
}

func TestLooksLikeHeader(t *testing.T) {
	if looksLikeHeader([]string{"You don't have permission to access this page"}) {
		t.Error("a permission message passed as a header")
	}
	if !looksLikeHeader([]string{"division", "day"}) {
		t.Error("a real header rejected")
	}
}
