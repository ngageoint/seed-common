package util

import (
	"testing"
	"fmt"
	"strings"
)

func init() {
	InitPrinter(PrintErr)
}

func TestUnescapeManifestLabel(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"\\\"\\\\", "\\"},
		{"\\\\\\$", "\\\\$"},
		{"\\/", "/"},
		{"\"12\\\"3\"", "12\"3"},
	}

	for _, c := range cases {
		out := UnescapeManifestLabel(c.input)
		if out != c.output {
			t.Errorf("UnescapeManifestLabel(%q) returned %s, expected %s", c.input, out, c.output)
		}
	}
}

func TestParseSeedImageName(t *testing.T) {
	cases := []struct {
		input  string
		output string
		errStr string
	}{
		{"no-colon-1.0.0-seed", "[   ]", "ERROR: No colons in seed image name"},
		{"colon-blow:1.0.0-seed:1.0.0", "[   ]", "ERROR: More than one colon in seed image name"},
		{"seedless-1.0.0:1.0.0", "[   1.0.0]", "ERROR: Expected -seed, found 1.0.0"},
		{"extractor-0.1.0-seed:0.1.0", "[extractor extractor 0.1.0 0.1.0]", ""},
		{"docker.io/geointseed/extractor-0.1.0-seed:0.1.0", "[extractor docker.io/geointseed/extractor 0.1.0 0.1.0]", ""},
	}

	for _, c := range cases {
		result, err := ParseSeedImageName(c.input)
		resultStr := fmt.Sprintf("%s", result)
		if resultStr != c.output {
			t.Errorf("ParseSeedImageName(%q) returned %s, expected %s", c.input, resultStr, c.output)
		}

		if err == nil && c.errStr != "" {
			t.Errorf("ParseSeedImageName(%q) did not return an error when one was expected", resultStr, c.output)
		}

		if err != nil && !strings.Contains(err.Error(), c.errStr) {
			t.Errorf("ParseSeedImageName returned an error: %v\n expected %v", err, c.errStr)
		}
	}
}