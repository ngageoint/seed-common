package util

import (
	"testing"
	"fmt"
)

func init() {
	InitPrinter(PrintErr)
}

func TestGetNormalizedVariable(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"abc-def", "ABC_DEF"},
		{"ABC_def", "ABC_DEF"},
		{"12345", ""},
		{"123_ab-CD", "_AB_CD"},
	}

	for _, c := range cases {
		out := GetNormalizedVariable(c.input)
		if out != c.output {
			t.Errorf("GetNormalizedVariable(%q) returned %s, expected %s", c.input, out, c.output)
		}
	}
}

func TestIsReserved(t *testing.T) {
	cases := []struct {
		name      string
		allocated []string
		reserved  bool
	}{
		{"abc-def", []string{""}, false},
		{"OUTPUT_DIR", []string{""}, true},
		{"test1", []string{"TEST"}, false},
		{"test", []string{"TEST", "TEST1"}, true},
	}

	for _, c := range cases {
		reserved := IsReserved(c.name, c.allocated)
		if reserved != c.reserved {
			t.Errorf("IsReserved(%q, %q) returned %s, expected %s", c.name, c.allocated, reserved, c.reserved)
		}
	}
}

func TestIsInUse(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		in_use bool
	}{
		{"test", "path.one", false},
		{"test", "path.two", true},
		{"test1", "path.three", true},
		{"test_one", "path.four", false},
	}

	vars := make(map[string][]string)
	for _, c := range cases {
		used := IsInUse(c.name, c.path, vars)
		if used != c.in_use {
			t.Errorf("IsInUse(%q, %q, %q) returned %s, expected %s", c.name, c.path, vars, used, c.in_use)
		}
	}
}

func TestRemoveString(t *testing.T) {
	cases := []struct {
		s      string
		list   []string
		result []string
	}{
		{"test", []string{}, []string{}},
		{"test", []string{"test"}, []string{}},
		{"test1", []string{"test", "test1"}, []string{"test"}},
	}

	for _, c := range cases {
		result := RemoveString(c.list, c.s)

		if fmt.Sprintf("%s", result) != fmt.Sprintf("%s", c.result) {
			t.Errorf("RemoveString(%q, %q) returned %s, expected %s", c.list, c.s, result, c.result)
		}
	}
}

func TestContainsString(t *testing.T) {
	cases := []struct {
		s      string
		list   []string
		result bool
	}{
		{"test", []string{}, false},
		{"test", []string{"test"}, true},
		{"test1", []string{"test", "test1"}, true},
	}

	for _, c := range cases {
		result := ContainsString(c.list, c.s)

		if result != c.result {
			t.Errorf("RemoveString(%q, %q) returned %s, expected %s", c.list, c.s, result, c.result)
		}
	}
}
